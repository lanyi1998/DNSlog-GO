package ipwry

import (
	"net/http"
	"os"
)

const (
	// IndexLen 索引长度
	IndexLen = 7
	// RedirectMode1 国家的类型, 指向另一个指向
	RedirectMode1 = 0x01
	// RedirectMode2 国家的类型, 指向一个指向
	RedirectMode2 = 0x02
)

// ResultQQwry 归属地信息
type ResultQQwry struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

type fileData struct {
	Data  []byte
	Path  *os.File
	IPNum int64
}

// QQwry 纯真ip库
type QQwry struct {
	Data   *fileData
	Offset int64
}

// Response 向客户端返回数据的
type Response struct {
	r *http.Request
	w http.ResponseWriter
}
