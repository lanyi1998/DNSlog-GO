简介
---
DNSLog-GO 是一款golang编写的监控 DNS 解析记录的工具，自带WEB界面

安装
---

详细图文教程:https://mp.weixin.qq.com/s/m_UXJa0imfOi721bkBpwFg

1.获取发行版
    这里 https://github.com/lanyi1998/DNSlog-GO/releases 下载最新发行版,并解压
    
2.域名与公网 IP 准备
   
    搭建并使用 DNSLog，你需要拥有两个域名，一个域名作为 NS 服务器域名(例:a.com)，一个用于记录域名(例: b.com)。还需要有一个公网 IP 地址(如：1.1.1.1)

    注意：b.com 的域名提供商需要支持自定义 NS 记录, a.com 则无要求。
    
    在 a.com 中设置两条 A 记录：
    
    ns1.a.com  A 记录指向  1.1.1.1        
    ns2.a.com  A 记录指向  1.1.1.1
    修改 b.com 的 NS 记录为 1 中设定的两个域名
    
    本步骤中，需要在域名提供商提供的页面进行设置，部分域名提供商只允许修改 NS 记录为已经认证过的 NS 地址。所以需要找一个支持修改 NS 记录为自己 NS 的域名提供商。
    
    注意: NS 记录修改之后部分地区需要 24-48 小时会生效
    
3.修改配置文件 config.ini

    Port = 8080 //HTTP监听端口
    Token = admin //API token
    ConsoleDisable = false //禁用web控制台，设置为true以后无法访问web页面，只能通过API获取数据
    Domain = a.com //绑定自己的域名,避免无效域名和其他网络扫描
    

4.启动服务
    VPS上，root运行 ./main,即可启动DNS和HTTP监听
    
演示截图:

![avatar](https://github.com/lanyi1998/DNSlog-GO/raw/master/images/demo.png)


go依赖:

`go get gopkg.in/gcfg.v1`

`go get golang.org/x/net/dns/dnsmessage`

