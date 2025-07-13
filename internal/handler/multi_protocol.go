package handler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// BufferedConn 包装连接，保存已读取的数据
type BufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (bc *BufferedConn) Read(b []byte) (int, error) {
	return bc.reader.Read(b)
}

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

func handleHTTP(conn net.Conn, reader *bufio.Reader, ginHandler http.Handler) {
	// 创建包装连接
	bufferedConn := &BufferedConn{
		Conn:   conn,
		reader: reader,
	}

	// 创建HTTP服务器
	server := &http.Server{
		Handler:      ginHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// 为单个连接创建临时监听器
	listener := &singleConnListener{conn: bufferedConn}

	// 处理连接
	server.Serve(listener)
}

// singleConnListener 单连接监听器
type singleConnListener struct {
	conn net.Conn
	used bool
}

func (l *singleConnListener) Accept() (net.Conn, error) {
	if l.used {
		return nil, io.EOF
	}
	l.used = true
	return l.conn, nil
}

func (l *singleConnListener) Close() error {
	return nil // 不关闭连接，让HTTP服务器处理
}

func (l *singleConnListener) Addr() net.Addr {
	return l.conn.LocalAddr()
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

	peekSize := 512
	data, err := reader.Peek(peekSize)
	if err != nil && err != io.EOF {
		conn.Close()
		return
	}

	// 判断协议类型
	if isHTTP(data) {
		// 重置读取超时
		conn.SetReadDeadline(time.Time{})
		handleHTTP(conn, reader, ginHandler)
		return
	}
	if isJRMI(data) {
		conn.SetReadDeadline(time.Time{})
		handleJRMI(conn, reader)
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

var (
	jrmiMagic       = []byte{0x4A, 0x52, 0x4D, 0x49} // "JRMI"
	jrmpVersion1    = []byte{0x00, 0x01}             // Version 1
	jrmpPingType    = byte(0x4B)                     // 'K'
	jrmpPingAckType = byte(0x4C)                     // 'L'
)

func handleJRMI(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()
	data := make([]byte, 1024)
	_, err := reader.Read(data)
	if err != nil && err != io.EOF {
		return
	}
	pingAckResponse := make([]byte, 7)
	copy(pingAckResponse[0:4], jrmiMagic)
	copy(pingAckResponse[4:6], jrmpVersion1)
	pingAckResponse[6] = jrmpPingAckType
	// 发送 PingAck
	_, err = conn.Write(pingAckResponse)
	_, err = reader.Read(data)

}