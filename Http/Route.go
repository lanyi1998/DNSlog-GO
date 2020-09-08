package Http

import (
	"../Core"
	"../Dns"
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

var resp RespData

func index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/template", http.StatusMovedPermanently)
}

func GetDnsData(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("token")
	if key == Core.Config.HTTP.Token {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            Dns.D.Get(),
		}))
	} else {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "403",
			Msg:            "false",
		}))
	}
}

func verifyToken(w http.ResponseWriter, r *http.Request) {
	var data map[string]string
	token, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(token, &data)
	if data["token"] == Core.Config.HTTP.Token {
		fmt.Fprintf(w, JsonRespData(RespData{
			HTTPStatusCode: "200",
			Msg:            "true",
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
	if key == Core.Config.HTTP.Token {
		Dns.D.Clear()
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
	type queryInfo struct {
		Query string // 首字母大写
	}
	var Q queryInfo
	key := r.Header.Get("token")
	if key == Core.Config.HTTP.Token {
		body, _ := ioutil.ReadAll(r.Body)
		json.Unmarshal(body, &Q)
		resp := RespData{
			HTTPStatusCode: "200",
			Msg:            "false",
		}
		for _, v := range Dns.DnsData {
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
