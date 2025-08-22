package handler

import (
	"DnsLog/internal/config"
	"DnsLog/internal/ipwry"
	"DnsLog/internal/model"
	"DnsLog/pkg/httpHandle"
	"DnsLog/pkg/ldap"
	"DnsLog/pkg/rmi"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func MultiProtocolListener(ginEngine http.Handler, port int) error {
	// 启动监听器
	portStr := fmt.Sprintf("%d", port)
	listener, err := net.Listen("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		return err
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		// 并发处理连接
		go handleConnection(conn, ginEngine)
	}
}

// isHTTP 判断是否为HTTP请求
func isHTTP(data []byte) bool {
	if len(data) < 4 {
		return false
	}

	// 检查HTTP方法
	methods := []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS", "PATCH", "TRACE"}
	dataStr := string(data)

	for _, method := range methods {
		if strings.HasPrefix(dataStr, method+" ") {
			return true
		}
	}
	return false
}

func isJRMI(data []byte) bool {
	dataStr := string(data)
	if strings.HasPrefix(dataStr, "JRMI") {
		return true
	}
	return false
}
func isLDAP(data []byte) bool {
	if bytes.HasPrefix(data, []byte{48, 12, 2, 1}) {
		return true
	}
	return false
}

// handleConnection 处理单个连接
func handleConnection(conn net.Conn, ginHandler http.Handler) {
	defer func() {
		if r := recover(); r != nil {
			conn.Close()
		}
	}()

	// 设置读取超时
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	// 创建缓冲读取器
	reader := bufio.NewReader(conn)

	peekSize := 4
	data, err := reader.Peek(peekSize)
	if err != nil && err != io.EOF {
		conn.Close()
		return
	}

	if isHTTP(data) {
		conn.SetReadDeadline(time.Time{})
		httpClient := httpHandle.Client{}
		httpClient.HandleHTTP(conn, reader, ginHandler)
		return
	}
	if isLDAP(data) {
		ldapClient := ldap.Client{}
		searchReq, _ := ldapClient.HandleLDAP(conn, reader)
		if searchReq != nil {
			username := searchReq.BaseObject
			for k, v := range config.Config.User {
				if v == username {
					ipStr := strings.Split(conn.RemoteAddr().String(), ":")[0]
					IpLocation, _ := ipwry.Query(ipStr)
					model.UserDnsDataMap.Set(k, model.DnsInfo{
						Type:       "LDAP",
						Subdomain:  username,
						Ipaddress:  ipStr,
						Time:       time.Now().Unix(),
						IpLocation: IpLocation,
					})
					break
				}
			}
		}
	}
	if isJRMI(data) {
		rmiClient := rmi.Client{}
		path := rmiClient.HandleJRMI(conn, reader)
		if path != "" {
			for k, v := range config.Config.User {
				if v == path {
					ipStr := strings.Split(conn.RemoteAddr().String(), ":")[0]
					IpLocation, _ := ipwry.Query(ipStr)
					model.UserDnsDataMap.Set(k, model.DnsInfo{
						Type:       "RMI",
						Subdomain:  path,
						Ipaddress:  ipStr,
						Time:       time.Now().Unix(),
						IpLocation: IpLocation,
					})
				}
			}
		}
	}
	handleOtherProtocol(conn, reader)
}

// handleOtherProtocol 处理其他协议
func handleOtherProtocol(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()

	// 读取一行数据
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return
	}

	// 简单回复
	response := fmt.Sprintf("Echo: %s", line)
	conn.Write([]byte(response))
}