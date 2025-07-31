package ldap

import (
	"bufio"
	"errors"
	"fmt"
	ber "github.com/go-asn1-ber/asn1-ber"
	"net"
)

// Message LDAP请求结构
type Message struct {
	MessageID  int64
	ProtocolOp *ber.Packet
	Controls   []*ber.Packet
}

// BindRequest LDAP绑定请求
type BindRequest struct {
	Version        int64
	Name           string
	Authentication interface{}
}

// SearchRequest LDAP搜索请求
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

type Client struct {
}

func (c *Client) HandleLDAP(conn net.Conn, reader *bufio.Reader) (*SearchRequest, error) {
	buffer := make([]byte, 4096)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			if err.Error() != "EOF" {
			}
			break
		}

		if n > 0 {
			msg, err := c.processLDAPMessage(buffer[:n], conn)
			if msg != nil {
				return msg, err
			}
		}
	}
	return nil, errors.New("没有LDAP请求数据")
}

func (c *Client) processLDAPMessage(data []byte, conn net.Conn) (*SearchRequest, error) {
	// 解析BER编码的数据
	packet := ber.DecodePacket(data)
	// 解析LDAP消息
	ldapMsg, err := c.parseLDAPMessage(packet)
	if err != nil {
		return nil, err
	}
	// 根据操作类型处理请求
	return c.handleLDAPRequest(ldapMsg, conn)
}

func (c *Client) parseLDAPMessage(packet *ber.Packet) (*Message, error) {
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

	return &Message{
		MessageID:  messageID,
		ProtocolOp: protocolOp,
		Controls:   controls,
	}, nil
}

func (c *Client) createBindResponse(messageID int64, resultCode int64, matchedDN, diagnosticMessage string) *ber.Packet {
	response := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	bindResponse := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationBindResponse, nil, "Bind Response")
	bindResponse.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultCode, "Result Code"))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, matchedDN, "Matched DN"))
	bindResponse.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, diagnosticMessage, "Diagnostic Message"))

	response.AppendChild(bindResponse)
	return response
}

func (c *Client) sendResponse(response *ber.Packet, conn net.Conn) error {
	data := response.Bytes()
	_, err := conn.Write(data)
	if err != nil {
		return err
	}
	return nil
}

// handleSearchRequest 处理搜索请求
func (c *Client) handleSearchRequest(msg *Message, conn net.Conn) (*SearchRequest, error) {
	searchReq, err := c.parseSearchRequest(msg.ProtocolOp)
	if err != nil {
		return nil, err
	}

	// 发送搜索完成响应
	response := c.createSearchDoneResponse(msg.MessageID, 0, "", "搜索完成")
	err = c.sendResponse(response, conn)
	if err != nil {
		return nil, err
	}
	return searchReq, nil
}

func (c *Client) parseBindRequest(packet *ber.Packet) (*BindRequest, error) {
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

// handleBindRequest 处理绑定请求
func (c *Client) handleBindRequest(msg *Message, conn net.Conn) error {
	_, err := c.parseBindRequest(msg.ProtocolOp)
	if err != nil {
		return err
	}
	// 发送绑定响应
	response := c.createBindResponse(msg.MessageID, 0, "", "绑定成功")
	return c.sendResponse(response, conn)
}

func (c *Client) parseSearchRequest(packet *ber.Packet) (*SearchRequest, error) {
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

func (c *Client) createSearchDoneResponse(messageID int64, resultCode int64, matchedDN, diagnosticMessage string) *ber.Packet {
	response := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Response")
	response.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	searchDone := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationSearchResultDone, nil, "Search Result Done")
	searchDone.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, resultCode, "Result Code"))
	searchDone.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, matchedDN, "Matched DN"))
	searchDone.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, diagnosticMessage, "Diagnostic Message"))

	response.AppendChild(searchDone)
	return response
}

func (c *Client) handleLDAPRequest(msg *Message, conn net.Conn) (*SearchRequest, error) {
	opType := msg.ProtocolOp.Tag
	switch opType {
	case ApplicationBindRequest:
		return nil, c.handleBindRequest(msg, conn)
	case ApplicationSearchRequest:
		return c.handleSearchRequest(msg, conn)
	}
	return nil, fmt.Errorf("未支持的操作类型: %d", opType)
}