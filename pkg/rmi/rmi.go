package rmi

import (
	"bufio"
	"io"
	"net"
)

type Client struct {
}

func (c *Client) handleJRMI(conn net.Conn, reader *bufio.Reader) {
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