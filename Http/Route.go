package Http

import (
	"DnsLog/Core"
	"DnsLog/Dns"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type RespData struct {
	HTTPStatusCode string
	Msg            string
}

type BulkRespData struct {
	HTTPStatusCode string
	Msg            []string
}

type queryInfo struct {
	Query string
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/template", http.StatusMovedPermanently)
}

func GetDnsData(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            Dns.D.Get(key),
		}))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyTokenApi(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	token, _ := io.ReadAll(r.Body)
	json.Unmarshal(token, &data)
	if Core.VerifyToken(data["token"]) {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            Core.Config.HTTP.User[data["token"]] + "." + Core.Config.DNS.Domain,
		}))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func JsonRespData(resp interface{}) string {
	rs, err := json.Marshal(resp)
	if err != nil {
		log.Fatalln(err)
	}
	return string(rs)
}

func Clean(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		Dns.D.Clear(key)
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            "success",
		}))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyDns(w http.ResponseWriter, r *http.Request) {
	Dns.DnsDataRwLock.RLock()
	defer Dns.DnsDataRwLock.RUnlock()
	var Q queryInfo
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &Q)
		resp := RespData{
			HTTPStatusCode: "200",
			Msg:            "false",
		}
		for _, v := range Dns.DnsData[key] {
			if v.Subdomain == Q.Query {
				resp.Msg = "true"
				break
			}

		}
		fmt.Fprint(w, JsonRespData(resp))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyHttp(w http.ResponseWriter, r *http.Request) {
	Dns.DnsDataRwLock.RLock()
	defer Dns.DnsDataRwLock.RUnlock()
	var Q queryInfo
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &Q)
		resp := RespData{
			HTTPStatusCode: "200",
			Msg:            "false",
		}
		for _, v := range Dns.DnsData[key] {
			if v.Subdomain == Q.Query && v.Type == "HTTP" {
				resp.Msg = "true"
				break
			}

		}
		fmt.Fprint(w, JsonRespData(resp))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func BulkVerifyDns(w http.ResponseWriter, r *http.Request) {
	Dns.DnsDataRwLock.RLock()
	defer Dns.DnsDataRwLock.RUnlock()
	var Q []string
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &Q)
		var result []string
		for _, v := range Dns.DnsData[key] {
			for _, q := range Q {
				if v.Subdomain == q {
					result = append(result, q)
				}
			}
		}
		var resp BulkRespData
		if len(result) == 0 {
			resp = BulkRespData{
				HTTPStatusCode: "200",
				Msg:            result,
			}
		} else {
			resp = BulkRespData{
				HTTPStatusCode: "200",
				Msg:            removeDuplication(result),
			}
		}
		fmt.Fprint(w, JsonRespData(resp))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func BulkVerifyHttp(w http.ResponseWriter, r *http.Request) {
	Dns.DnsDataRwLock.RLock()
	defer Dns.DnsDataRwLock.RUnlock()
	var Q []string
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &Q)
		var result []string
		for _, v := range Dns.DnsData[key] {
			for _, q := range Q {
				if v.Subdomain == q && v.Type == "HTTP" {
					result = append(result, q)
				}
			}
		}
		resp := BulkRespData{
			HTTPStatusCode: "200",
			Msg:            removeDuplication(result),
		}
		fmt.Fprint(w, JsonRespData(resp))
	} else {
		fmt.Fprint(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func removeDuplication(arr []string) []string {
	j := 0
	for i := 1; i < len(arr); i++ {
		if arr[i] == arr[j] {
			continue
		}
		j++
		arr[j] = arr[i]
	}
	return arr[:j+1]
}

func HttpRequestLog(w http.ResponseWriter, r *http.Request) {
	user := Core.GetUser(r.URL.Path)
	Dns.D.Set(user, Dns.DnsInfo{
		Type:      "HTTP",
		Subdomain: r.URL.Path,
		Ipaddress: r.RemoteAddr,
		Time:      time.Now().Unix(),
	})
}
