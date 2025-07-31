package httpHandle

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"time"
)

type Client struct {
}

// BufferedConn 包装连接，保存已读取的数据
type BufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *BufferedConn) Read(b []byte) (int, error) {
	return c.reader.Read(b)
}

func (c *Client) HandleHTTP(conn net.Conn, reader *bufio.Reader, ginHandler http.Handler) {
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