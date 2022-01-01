package Http

import (
	"DnsLog/Core"
	"DnsLog/Dns"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RespData struct {
	HTTPStatusCode string
	Msg            string
}

type queryInfo struct {
	Query string // 首字母大写
}

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/template", http.StatusMovedPermanently)
}

func GetDnsData(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            Dns.D.Get(key),
		}))
	} else {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyTokenApi(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	token, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(token, &data)
	if Core.VerifyToken(data["token"]) {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            Core.User[data["token"]] + "." + Core.Config.Dns.Domain,
		}))
	} else {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func JsonRespData(resp RespData) string {
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
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            "success",
		}))
	} else {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyDns(w http.ResponseWriter, r *http.Request) {
	var Q queryInfo
	key := r.Header.Get("token")
	if Core.VerifyToken(key) {
		body, _ := ioutil.ReadAll(r.Body)
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
		fmt.Fprintf(w, JsonRespData(resp))
	} else {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}
