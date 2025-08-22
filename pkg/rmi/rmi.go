package rmi

import (
	"bufio"
	"bytes"
	"net"
)

type Client struct {
}

var (
	// rmi protocol
	// https://docs.oracle.com/javase/9/docs/specs/rmi/protocol.html
	rmireplay = []byte{
		0x4e, 0x00, 0x09, // 保证4e00开头
		0x31, 0x32, 0x37, 0x2e, 0x30, 0x2e, 0x30, 0x2e, 0x31, // 模拟 127.0.0.1
		0x00, 0x00, 0xc4, 0x12,
	}
)

func createHandshakeResponse() []byte {
	// 标准 Java RMI 响应
	response := []byte{
		0x4E, // 确认字节 (Protocol acknowledgment)
	}

	// Java 序列化流头 (ObjectStreamConstants)
	streamHeader := []byte{
		0xAC, 0xED, // STREAM_MAGIC
		0x00, 0x05, // STREAM_VERSION
	}

	// 组合响应
	fullResponse := append(response, streamHeader...)

	// 添加一些示例数据 (可选)
	additionalData := []byte{
		0x77, 0x22, // 块数据标记和长度
		// 后续可以添加更多服务端信息
	}

	return append(fullResponse, additionalData...)
}

func (c *Client) HandleJRMI(conn net.Conn, reader *bufio.Reader) string {
	buf := make([]byte, 1024)
	defer conn.Close()
	conn.Write(rmireplay)
	reader.Read(buf)
	conn.Write([]byte{})
	conn.Read(buf)
	conn.Write([]byte{})
	conn.Read(buf)
	var dataList []byte
	var flag bool
	// 从后往前读因为空都是00
	for i := len(buf) - 1; i >= 0; i-- {
		// 这里要用一个flag来区分
		// 因为正常数据中也会含有00
		if buf[i] != 0x00 || flag {
			flag = true
			dataList = append(dataList, buf[i])
		}
	}
	// 已读到的长度等于当前读到的字节代表的数字
	// 那么认为已读到的字符串翻转后是路径参数
	var j_ int
	for i := 0; i < len(dataList); i++ {
		if int(dataList[i]) == i {
			j_ = i
			break
		}
	}

	if len(dataList) < j_ {
		return ""
	}
	temp := dataList[0:j_]
	pathBytes := &bytes.Buffer{}
	// 翻转后拿到真正的路径参数
	for i := len(temp) - 1; i >= 0; i-- {
		pathBytes.Write([]byte{dataList[i]})
	}
	path := pathBytes.String()
	return path
}