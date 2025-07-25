package handler

import (
	"DnsLog/internal/config"
	"DnsLog/internal/ipwry"
	"DnsLog/internal/model"
	"bufio"
	"bytes"
	"fmt"
	"github.com/go-asn1-ber/asn1-ber"
	"io"
	"log"
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
func isLDAP(data []byte) bool {
	if bytes.HasPrefix(data, []byte{48, 12, 2, 1}) {
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

	peekSize := 4
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
	if isLDAP(data) {
		conn.SetReadDeadline(time.Time{})
		handleLDAP(conn, reader)
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

// LDAPMessage LDAP请求结构
type LDAPMessage struct {
	MessageID  int64
	ProtocolOp *ber.Packet
	Controls   []*ber.Packet
}

func handleLDAP(conn net.Conn, reader *bufio.Reader) {
	buffer := make([]byte, 4096)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
				log.Printf("读取数据错误: %v\n", err)
			}
			break
		}

		if n > 0 {
			processLDAPMessage(buffer[:n], conn)
		}
	}
}

func processLDAPMessage(data []byte, conn net.Conn) {
	// 解析BER编码的数据
	packet := ber.DecodePacket(data)
	// 解析LDAP消息
	ldapMsg, err := parseLDAPMessage(packet)
	if err != nil {
		return
	}
	// 根据操作类型处理请求
	handleLDAPRequest(ldapMsg, conn)
}

func parseLDAPMessage(packet *ber.Packet) (*LDAPMessage, error) {
	if len(packet.Children) < 2 {
		return nil, fmt.Errorf("无效的LDAP消息格式")
	}

	// 消息ID
	messageID, ok := packet.Children[0].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("无效的消息ID")
	}

	// 协议操作
	protocolOp := packet.Children[1]

	// 控制信息（可选）
	var controls []*ber.Packet
	if len(packet.Children) > 2 {
		controls = packet.Children[2].Children
	}

	return &LDAPMessage{
		MessageID:  messageID,
		ProtocolOp: protocolOp,
		Controls:   controls,
	}, nil
}

// LDAP操作类型常量
const (
	ApplicationBindRequest      = 0
	ApplicationBindResponse     = 1
	ApplicationUnbindRequest    = 2
	ApplicationSearchRequest    = 3
	ApplicationSearchResultDone = 5
)

func handleLDAPRequest(msg *LDAPMessage, conn net.Conn) {
	opType := msg.ProtocolOp.Tag

	switch opType {
	case ApplicationBindRequest:
		handleBindRequest(msg, conn)
	case ApplicationSearchRequest:
		handleSearchRequest(msg, conn)
	}
}

// BindRequest LDAP绑定请求
type BindRequest struct {
	Version        int64
	Name           string
	Authentication interface{}
}

func parseBindRequest(packet *ber.Packet) (*BindRequest, error) {
	if len(packet.Children) < 3 {
		return nil, fmt.Errorf("无效的绑定请求格式")
	}

	version, ok := packet.Children[0].Value.(int64)
	if !ok {
		return nil, fmt.Errorf("无效的版本号")
	}

	name, ok := packet.Children[1].Value.(string)
	if !ok {
		return nil, fmt.Errorf("无效的绑定名称")
	}

	return &BindRequest{
		Version:        version,
		Name:           name,
		Authentication: packet.Children[2].Value,
	}, nil
}

// 处理绑定请求
func handleBindRequest(msg *LDAPMessage, conn net.Conn) {
	_, err := parseBindRequest(msg.ProtocolOp)
	if err != nil {
		return
	}
	// 发送绑定响应
	response := createBindResponse(msg.MessageID, 0, "", "绑定成功")
	sendResponse(response, conn)
}

func createBindResponse(messageID int64, resultCode int64, matchedDN, diagnosticMessage string) *ber.Packet {
	response := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	bindResponse := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationBindResponse, nil, "Bind Response")
	bindResponse.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultCode, "Result Code"))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, matchedDN, "Matched DN"))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, diagnosticMessage, "Diagnostic Message"))

	response.AppendChild(bindResponse)
	return response
}

func sendResponse(response *ber.Packet, conn net.Conn) {
	data := response.Bytes()
	_, err := conn.Write(data)
	if err != nil {
		log.Printf("发送响应错误: %v\n", err)
	}
}

// handleSearchRequest 处理搜索请求
func handleSearchRequest(msg *LDAPMessage, conn net.Conn) {
	searchReq, err := parseSearchRequest(msg.ProtocolOp)
	if err != nil {
		return
	}

	if strings.HasPrefix(searchReq.BaseObject, "s=") {
		username := searchReq.BaseObject[2:]
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

	// 发送搜索完成响应
	response := createSearchDoneResponse(msg.MessageID, 0, "", "搜索完成")
	sendResponse(response, conn)
}

type SearchRequest struct {
	BaseObject   string
	Scope        int64
	DerefAliases int64
	SizeLimit    int64
	TimeLimit    int64
	TypesOnly    bool
	Filter       *ber.Packet
	Attributes   []string
}

func parseSearchRequest(packet *ber.Packet) (*SearchRequest, error) {
	if len(packet.Children) < 7 {
		return nil, fmt.Errorf("无效的搜索请求格式")
	}

	baseObject, _ := packet.Children[0].Value.(string)
	scope, _ := packet.Children[1].Value.(int64)
	derefAliases, _ := packet.Children[2].Value.(int64)
	sizeLimit, _ := packet.Children[3].Value.(int64)
	timeLimit, _ := packet.Children[4].Value.(int64)
	typesOnly, _ := packet.Children[5].Value.(bool)
	filter := packet.Children[6]

	var attributes []string
	if len(packet.Children) > 7 {
		attrList := packet.Children[7]
		for _, attr := range attrList.Children {
			if attrName, ok := attr.Value.(string); ok {
				attributes = append(attributes, attrName)
			}
		}
	}

	return &SearchRequest{
		BaseObject:   baseObject,
		Scope:        scope,
		DerefAliases: derefAliases,
		SizeLimit:    sizeLimit,
		TimeLimit:    timeLimit,
		TypesOnly:    typesOnly,
		Filter:       filter,
		Attributes:   attributes,
	}, nil
}

func createSearchDoneResponse(messageID int64, resultCode int64, matchedDN, diagnosticMessage string) *ber.Packet {
	response := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	searchDone := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationSearchResultDone, nil, "Search Result Done")
	searchDone.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultCode, "Result Code"))
	searchDone.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, matchedDN, "Matched DN"))
	searchDone.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, diagnosticMessage, "Diagnostic Message"))

	response.AppendChild(searchDone)
	return response
}
