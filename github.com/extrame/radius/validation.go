package radius

import (
	"bytes"
	"crypto"
	_ "crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

type AttributeDataType uint8

const (
	UNSPECIFIED AttributeDataType = iota
	TEXT        AttributeDataType = iota
	STRING      AttributeDataType = iota
	ADDRESS     AttributeDataType = iota
	MAC         AttributeDataType = iota
	VALUE       AttributeDataType = iota
	INTEGER     AttributeDataType = iota
	TIME        AttributeDataType = iota
)

const (
	UNLIMITED = -1
)

type Validation struct {
	Kind      AttributeDataType
	MinLength int
	MaxLength int //if < 0 unlimited
	Decode    func(p *Packet, attr *AVP) error
}

func (a AVP) Binary() ([]byte, error)    { return a.Value, nil }
func (a AVP) Address() (net.Addr, error) { return nil, nil }
func (a AVP) Integer() (int, error) {
	i := binary.BigEndian.Uint32(a.Value)
	return int(i), nil
}
func (a AVP) Uinteger32() (uint32, error) {
	return binary.BigEndian.Uint32(a.Value), nil
}
func (a AVP) Time() (time.Time, error) { return time.Time{}, nil }
func (a AVP) Text() (string, error)    { return string(a.text), nil }
func (a AVP) Mac() (net.HardwareAddr, error) {
	s := string(a.Value)
	s = strings.Replace(s, "-", ".", -1)
	return net.ParseMAC(s)
}
func (a AVP) IP() (net.IP, error) {
	return net.IPv4(a.Value[0], a.Value[1], a.Value[2], a.Value[3]), nil
}

func (v Validation) Validate(p *Packet, attr *AVP) error {

	if len(attr.Value) < v.MinLength {
		return errors.New("value too short")
	}

	if v.MaxLength != UNLIMITED && len(attr.Value) > v.MaxLength {
		return fmt.Errorf("value too long: test %d against %d", len(attr.Value), v.MaxLength)
	}

	if v.Decode != nil {
		return v.Decode(p, attr)
	}
	return nil
}

func DecodeUserPassword(p *Packet, a *AVP) error {
	//Decode password. XOR against md5(p.server.secret+Authenticator)
	secAuth := append([]byte(nil), []byte(p.server.secret)...)
	secAuth = append(secAuth, p.Authenticator[:]...)
	m := crypto.Hash(crypto.MD5).New()
	m.Write(secAuth)
	md := m.Sum(nil)
	pass := a.Value
	if len(pass) == 16 {
		for i := 0; i < len(pass); i++ {
			pass[i] = pass[i] ^ md[i]
		}
		a.Value = bytes.TrimRight(pass, string([]rune{0}))
		return nil
	}
	return errors.New("not implemented for password > 16")
}

var validation = map[AttributeType]Validation{
	UserName:               {UNSPECIFIED, 1, UNLIMITED, nil},
	UserPassword:           {UNSPECIFIED, 16, 128, DecodeUserPassword},
	CHAPPassword:           {UNSPECIFIED, 17, 17, nil},
	NASIPAddress:           {ADDRESS, 4, 4, nil},
	NASPort:                {VALUE, 4, 4, nil},
	MessageAuthenticator:   {UNSPECIFIED, 1, UNLIMITED, nil},
	ServiceType:            {UNSPECIFIED, 1, UNLIMITED, nil},
	FramedProtocol:         {UNSPECIFIED, 1, UNLIMITED, nil},
	FramedIPAddress:        {UNSPECIFIED, 1, UNLIMITED, nil},
	FramedIPNetmask:        {},
	FramedRouting:          {},
	FilterId:               {},
	FramedMTU:              {},
	FramedCompression:      {},
	LoginIPHost:            {},
	LoginService:           {},
	LoginTCPPort:           {},
	ReplyMessage:           {},
	CallbackNumber:         {},
	CallbackId:             {},
	FramedRoute:            {},
	FramedIPXNetwork:       {},
	State:                  {},
	Class:                  {},
	VendorSpecific:         {UNSPECIFIED, 1, UNLIMITED, nil},
	SessionTimeout:         {UNSPECIFIED, 1, UNLIMITED, nil},
	IdleTimeout:            {UNSPECIFIED, 1, UNLIMITED, nil},
	TerminationAction:      {},
	CalledStationId:        {MAC, 1, UNLIMITED, nil},
	CallingStationId:       {UNSPECIFIED, 1, UNLIMITED, nil},
	NASIdentifier:          {UNSPECIFIED, 1, UNLIMITED, nil},
	ProxyState:             {},
	LoginLATService:        {},
	LoginLATNode:           {},
	LoginLATGroup:          {},
	FramedAppleTalkLink:    {},
	FramedAppleTalkNetwork: {},
	FramedAppleTalkZone:    {},
	AcctStatusType:         {INTEGER, 1, UNLIMITED, nil},
	AcctDelayTime:          {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctInputOctets:        {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctOutputOctets:       {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctSessionId:          {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctAuthentic:          {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctSessionTime:        {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctInputPackets:       {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctOutputPackets:      {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctTerminateCause:     {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctMultiSessionId:     {},
	AcctLinkCount:          {UNSPECIFIED, 1, UNLIMITED, nil},
	EventTimestamp:         {UNSPECIFIED, 1, UNLIMITED, nil},

	CHAPChallenge:       {UNSPECIFIED, 1, UNLIMITED, nil},
	NASPortType:         {UNSPECIFIED, 1, UNLIMITED, nil},
	NASPortId:           {UNSPECIFIED, 1, UNLIMITED, nil},
	LocationInformation: {UNSPECIFIED, 1, UNLIMITED, nil},
	LocationData:        {UNSPECIFIED, 1, UNLIMITED, nil},
	PortLimit:           {},
	LoginLATPort:        {},
	AcctInputGigawords:  {UNSPECIFIED, 1, UNLIMITED, nil},
	AcctOutputGigawords: {UNSPECIFIED, 1, UNLIMITED, nil},
	ConnectInfo:         {UNSPECIFIED, 1, UNLIMITED, nil},
}

/*
AccessRequest      PacketCode = 1
MUST NAS-IP-Address AND/OR NAS-Identifier

     User-Password XOR CHAP-Password XOR State XOR someother authentication





AccessAccept       PacketCode = 2
AccessReject       PacketCode = 3


AccountingRequest  PacketCode = 4
AccountingResponse PacketCode = 5
AccessChallenge    PacketCode = 11

*/
