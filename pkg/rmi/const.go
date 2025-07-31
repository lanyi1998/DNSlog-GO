package rmi

var (
	jrmiMagic       = []byte{0x4A, 0x52, 0x4D, 0x49} // "JRMI"
	jrmpVersion1    = []byte{0x00, 0x01}             // Version 1
	jrmpPingType    = byte(0x4B)                     // 'K'
	jrmpPingAckType = byte(0x4C)                     // 'L'
)