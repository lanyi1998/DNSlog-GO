简介
---

DNSLog-GO 是一款golang编写的监控 DNS 解析记录的工具。

支持:

- docker一行命令部署
- WEB界面 或者 API 模式
- 多账号，可供多人使用。或者工具扫描和人工测试，数据互不干扰
- 支持通过置DNA A记录，用于临时测试或者ns服务自解析

演示截图:

![avatar](images/demo.png)

安装
---

详细图文教程:https://mp.weixin.qq.com/s/m_UXJa0imfOi721bkBpwFg

# 1.获取发行版

这里 https://github.com/lanyi1998/DNSlog-GO/releases 下载最新发行版,并解压

或者docker启动
```shell
wget https://raw.githubusercontent.com/lanyi1998/DNSlog-GO/master/config.yaml
#修改你的config.yaml文件
docker run -d -p 53:53 -p 53:53/udp -p 8000:8000 -v `pwd`/config.yaml:/DNSlog-GO/config.yaml --name dnslog --privileged --restart=always lanyi1998/dnslog-go:latest
```

# 2.域名与公网 IP 准备

```
搭建并使用 DNSLog，你需要拥有一个域名如(a.com),还需要有一个公网 IP 地址(如：1.1.1.1)
    
修改 a.com 的 NS 记录为 

NS1.vpsip.NIP.IO
NS2.vpsip.NIP.IO

如 
ns1.1.1.1.1.nip.io
ns1.1.1.1.1.nip.io

本步骤中，需要在域名提供商提供的页面进行设置，部分域名提供商只允许修改 NS 记录为已经认证过的 NS 地址。所以需要找一个支持修改 NS 记录为自己 NS 的域名提供商。
    
注意: NS 记录修改之后部分地区需要 24-48 小时会生效
```

# 3.修改配置文件 config.ini

```
Http:
  host: 0.0.0.0
  port: 8000
  consoleDisable: false

User:
  #"token":"用户对应子域名"
  'admin': 'dnslog'

Dns:
  domain: demo.com
  ARecord:
    'www': '1.1.1.1'
```

# 4.启动对应系统的客户端

**注意服务端重启以后，如果修改了用户对应子域名，必须清空一下浏览器中的localStorage,否则会获取不到数据**

---

API Python Demo

```python
import requests
import random
import json


class DnsLog():
    domain = ""
    token = ""
    Webserver = ""

    def __init__(self, Webserver, token):
        self.Webserver = Webserver  # dnslog的http监听地址，格式为 ip:端口
        self.token = token  # token
        # 检测DNSLog服务器是否正常
        res = requests.post("http://" + Webserver + "/api/verifyToken", json={"token": token}).json()
        if res['code'] == 401:
            exit(res['msg'])
        self.domain = res['data']['subdomain']


    # 生成随机子域名
    def randomSubDomain(self, length=5):
        subDomain = ''.join(random.sample('zyxwvutsrqponmlkjihgfedcba', length)) + '.' + self.domain
        return subDomain

    # 验证子域名是否存在
    def checkDomain(self, domain):
        res = requests.post("http://" + self.Webserver + "/api/verifyDns", json={"Query": domain},
                            headers={"token": self.token}).json()
        if res["msg"] == 'success':
            return True
        else:
            return False


url = "https://www.baidu.com/"

dns = DnsLog("1.1.1.1:8000", "admin")

subDomain = dns.randomSubDomain()

payload = {
    "b": {
        "@type": "java.net.Inet4Address",
        "val": subDomain
    }
}

requests.post(url, json=payload)

if dns.checkDomain(subDomain):
    print("存在FastJosn")
```

# 5. CLI

发布版本提供 `dnslog-cli-macos.zip`、`dnslog-cli-linux.zip` 和 `dnslog-cli-windows.zip`（均为 amd64）。解压后得到 `dnslog-cli`（Windows 为 `dnslog-cli.exe`），可将其放入 `PATH` 供 Skill 调用。

构建 CLI：

```bash
make build-cli
```

使用 CLI 写入连接配置，无需手动创建配置文件：

```bash
./build/dnslog-cli config set \
  --url "http://127.0.0.1:8000" \
  --token "admin"
```

token 可能被 shell 历史记录保存。共享机器上建议临时关闭历史记录，或改用环境变量。CLI 默认把配置写入 `~/.config/dnslog-go/config.yaml`，文件权限为 `0600`：

```yaml
base_url: http://127.0.0.1:8000
token: admin
```

也可以通过环境变量提供配置。环境变量的优先级高于配置文件：

```bash
export DNSLOG_BASE_URL="http://127.0.0.1:8000"
export DNSLOG_TOKEN="admin"
```

`DNSLOG_CONFIG` 可以指定其他配置文件路径。常用命令：

```bash
./build/dnslog-cli verify-token
./build/dnslog-cli config path
./build/dnslog-cli get-dns
./build/dnslog-cli verify-dns abc.dnslog.demo.com
./build/dnslog-cli verify-http /dnslog/test --timeout 15s --interval 1s
./build/dnslog-cli random-domain
./build/dnslog-cli help
```

`verify-dns` 和 `verify-http` 会先立即查询一次；未找到时默认每 `500ms` 重试，最多等待 `10s`。可以通过 `--timeout` 和 `--interval` 调整，或使用 `--timeout 0` 禁用轮询。

轮询期间成功查询但未命中时返回 `{"found":false}`；请求失败或在请求过程中超时时返回非零退出码，并将错误写入标准错误。

除帮助信息外，CLI 的标准输出均为 JSON，错误写入标准错误，适合 Agent 调用。

# 6. Agent Skill

仓库提供了可配合 CLI 使用的 Agent Skill：[skills/dnslog-go](skills/dnslog-go)。Skill 不保存或接收 token，CLI 会自动从环境变量或用户配置文件读取凭据。

安装 CLI：

```bash
go install github.com/lanyi1998/DNSlog-GO/cmd/dnslog-cli@latest
```

将 Skill 安装到 Codex：

```bash
mkdir -p ~/.codex/skills
cp -R skills/dnslog-go ~/.codex/skills/dnslog-go
```

之后可以通过 `$dnslog-go` 显式调用，也可以让 Agent 在处理 DNSlog-GO 任务时自动选择它。

# 7. Golang SDK

```golang
package main

import (
	"net/http"

	"github.com/lanyi1998/DNSlog-GO/pkg/sdk"
)

func main() {
	client, err := sdk.NewDnsLogClient("http://127.0.0.1:8000", "admin")
	if err != nil {
		panic(err)
	}

	dnslog :=  client.RandomSubDomain(10)
	client.RandomSSRFUrl(10)
	http.Get("http://" + dnslog)
	// 一个http请求一个dns
	if ok, _ := client.VerifyDns(dnslog); ok {
		println(">> [DNS] 验证成功", dnslog)
	}
	// 一个请求验证多个DNS,高效率,适用于扫描器等大批量发送dns请求的场景
	if ok, _ := client.VerifyDnsV2(dnslog); ok {
		println(">> [DNS] 验证成功", dnslog)
	}
	client.Clear()
}
```

## Stargazers over time

[![lanyi1998/DNSlog-GO GitStock K-Line Chart](https://gitstock.org/lanyi1998/DNSlog-GO/stock.svg)](https://gitstock.org/lanyi1998/DNSlog-GO)

## 404星链计划

[![Stargazers over time](https://github.com/knownsec/404StarLink-Project/raw/master/logo.png)](https://starchart.cc/lanyi1998/DNSlog-GO)

DNSlog-Go 现已加入 [404星链计划](https://github.com/knownsec/404StarLink)
