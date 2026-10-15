package sdk

import (
	"encoding/json"
	"errors"
	"net/http"
)

// DnsRecord represents one interaction returned by DNSlog-GO.
type DnsRecord struct {
	Type       string `json:"type"`
	Subdomain  string `json:"subdomain"`
	Ipaddress  string `json:"ipaddress"`
	IpLocation string `json:"ip_location,omitempty"`
	Request    string `json:"request,omitempty"`
	Time       int64  `json:"time"`
}

// UnmarshalJSON accepts the server's IpLocation field while preserving the CLI's
// snake_case JSON representation.
func (r *DnsRecord) UnmarshalJSON(data []byte) error {
	type recordJSON DnsRecord
	var decoded struct {
		recordJSON
		ServerIPLocation string `json:"IpLocation"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*r = DnsRecord(decoded.recordJSON)
	if r.IpLocation == "" {
		r.IpLocation = decoded.ServerIPLocation
	}
	return nil
}

type responseMeta struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type dnsDataResponse struct {
	responseMeta
	Data []DnsRecord `json:"data"`
}

type stringListResponse struct {
	responseMeta
	Data []string `json:"data"`
}

type operationResponse struct {
	responseMeta
}

type bulkVerifyHttpRequest struct {
	Query []string `json:"query"`
}

type setARecordRequest struct {
	Domain string `json:"domain"`
	IP     string `json:"ip"`
}

type setTXTRecordRequest struct {
	Domain string `json:"domain"`
	TXT    string `json:"txt"`
}

func (d *DnsLogClient) request(method, path string, body, result any, meta *responseMeta) error {
	request := d.httpClient.R().
		SetHeader("Token", d.token).
		SetResult(result)
	if body != nil {
		request.SetBody(body)
	}

	response, err := request.Execute(method, path)
	if err != nil {
		return err
	}
	if response.IsError() {
		return errors.New(http.StatusText(response.StatusCode()))
	}
	if meta.Code != http.StatusOK {
		if meta.Msg == "" {
			return errors.New("DNSlog-GO API request failed")
		}
		return errors.New(meta.Msg)
	}
	return nil
}

// GetDnsData returns all interactions associated with the configured token.
func (d *DnsLogClient) GetDnsData() ([]DnsRecord, error) {
	var result dnsDataResponse
	err := d.request(http.MethodGet, "/api/getDnsData", nil, &result, &result.responseMeta)
	return result.Data, err
}

// GetDnsDataAndClear returns all interactions and removes them from the server.
func (d *DnsLogClient) GetDnsDataAndClear() ([]DnsRecord, error) {
	var result dnsDataResponse
	err := d.request(http.MethodGet, "/api/getDnsData_clear", nil, &result, &result.responseMeta)
	return result.Data, err
}

// BulkVerifyHttp returns the requested HTTP paths that have been observed.
func (d *DnsLogClient) BulkVerifyHttp(paths []string) ([]string, error) {
	var result stringListResponse
	err := d.request(
		http.MethodPost,
		"/api/bulkVerifyHttp",
		bulkVerifyHttpRequest{Query: paths},
		&result,
		&result.responseMeta,
	)
	return result.Data, err
}

// SetARecord sets an A record for the configured token.
func (d *DnsLogClient) SetARecord(domain, ip string) error {
	var result operationResponse
	return d.request(
		http.MethodPost,
		"/api/setARecord",
		setARecordRequest{Domain: domain, IP: ip},
		&result,
		&result.responseMeta,
	)
}

// SetTXTRecord sets a TXT record for the configured token.
func (d *DnsLogClient) SetTXTRecord(domain, txt string) error {
	var result operationResponse
	return d.request(
		http.MethodPost,
		"/api/setTXTRecord",
		setTXTRecordRequest{Domain: domain, TXT: txt},
		&result,
		&result.responseMeta,
	)
}

// Clear removes all interactions associated with the configured token.
func (d *DnsLogClient) ClearRecords() error {
	var result operationResponse
	return d.request(http.MethodGet, "/api/clean", nil, &result, &result.responseMeta)
}
