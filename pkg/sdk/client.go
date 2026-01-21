package sdk

import (
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type DnsLogClient struct {
	baseUrl   string
	token     string
	Subdomain string
	PreFix    string
	keyPool   *KeyPool
}

var httpClient = resty.New()

type VerifyTokenRequest struct {
	Token string `json:"token"`
}

type VerifyTokenResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Subdomain string `json:"subdomain"`
		Token     string `json:"token"`
	} `json:"data"`
}

func NewDnsLogClient(baseUrl, token string) (*DnsLogClient, error) {
	httpClient.SetTimeout(10 * time.Second)
	httpClient.SetBaseURL(baseUrl)
	dnsClient := DnsLogClient{
		baseUrl: baseUrl,
		token:   token,
	}
	var respBody VerifyTokenResponse
	resp, err := httpClient.R().
		SetBody(VerifyTokenRequest{Token: dnsClient.token}). // 设置请求体（JSON）
		SetResult(&respBody).                                // 绑定响应结构体（自动反序列化）
		Post("/api/verifyToken")
	if err != nil {
		return nil, err
	}
	if respBody.Code != 200 || resp.IsError() {
		return nil, errors.New(respBody.Msg)
	}
	dnsClient.Subdomain = respBody.Data.Subdomain
	dnsClient.PreFix = strings.Split(dnsClient.Subdomain, ".")[0]
	dnsClient.keyPool = NewKeyPool(9999, 5*time.Second, &dnsClient)
	go dnsClient.keyPool.Start()
	return &dnsClient, nil
}

// RandomSubDomain 随机生成子域名
func (d *DnsLogClient) RandomSubDomain(length int) string {
	return randStr(length) + "." + d.Subdomain
}

func randStr(length int) string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// RandomSSRFUrl 随机生成SSRF URL
func (d *DnsLogClient) RandomSSRFUrl(length int) string {
	return d.baseUrl + "/" + d.PreFix + "/" + randStr(length)
}

func (d *DnsLogClient) VerifyHttp(url string) (bool, error) {
	key := strings.Replace(url, d.baseUrl, "", 1)
	var respBody VerifyDnsResponse
	resp, err := httpClient.R().
		SetHeader("Token", d.token).
		SetBody(VerifyDnsReqeust{Query: key}).
		SetResult(&respBody).
		Post("/api/verifyHttp")
	if err != nil {
		return false, err
	}
	if respBody.Code != 200 || resp.IsError() {
		return false, errors.New(respBody.Msg)
	}
	if respBody.Data.Subdomain != "" {
		return true, nil
	}
	return false, nil
}

type VerifyDnsReqeust struct {
	Query string `json:"query"`
}
type VerifyDnsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Ipaddress string `json:"ipaddress"`
		Subdomain string `json:"subdomain"`
		Time      int    `json:"time"`
		Type      string `json:"type"`
	} `json:"data"`
}

// VerifyDns 验证DNS是否存在
func (d *DnsLogClient) VerifyDns(domain string) (bool, error) {
	var respBody VerifyDnsResponse
	resp, err := httpClient.R().
		SetHeader("Token", d.token).
		SetBody(VerifyDnsReqeust{Query: domain}).
		SetResult(&respBody).
		Post("/api/verifyDns")
	if err != nil {
		return false, err
	}
	if respBody.Code != 200 || resp.IsError() {
		return false, errors.New(respBody.Msg)
	}
	if respBody.Data.Subdomain != "" {
		return true, nil
	}
	return false, nil
}

func (d *DnsLogClient) VerifyDnsV2(domain string) (bool, error) {
	if d.keyPool.DoRequest(domain) {
		return true, nil
	} else {
		return false, nil
	}
}

type BulkVerifyDnsRequest struct {
	Subdomain []string `json:"subdomain"`
}

type BulkVerifyDnsResponse struct {
	Code int      `json:"code"`
	Msg  string   `json:"msg"`
	Data []string `json:"data"`
}

func (d *DnsLogClient) BulkVerifyDns(domains []string) ([]string, error) {
	var respBody BulkVerifyDnsResponse
	resp, err := httpClient.R().
		SetHeader("Token", d.token).
		SetBody(BulkVerifyDnsRequest{Subdomain: domains}).
		SetResult(&respBody).
		Post("/api/bulkVerifyDns")
	if err != nil {
		return nil, err
	}
	if respBody.Code != 200 || resp.IsError() {
		return nil, errors.New(respBody.Msg)
	}
	return respBody.Data, nil
}

// Clear 清空DNS记录clean
func (d *DnsLogClient) Clear() error {
	_, err := httpClient.R().
		SetHeader("Token", d.token).
		Get("/api/clean")
	return err
}