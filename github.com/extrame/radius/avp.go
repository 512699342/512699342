package radius

import (
	"errors"
	"net"
	"time"
)

type AVP struct {
	Type    AttributeType
	Value   []byte
	text    string
	address net.Addr
	integer uint32
	time    time.Time
}

func (a AVP) Encode(b []byte) (n int, err error) {
	fullLen := len(a.Value) + 2 //type and length
	if fullLen > 255 || fullLen < 2 {
		return 0, errors.New("value too big for attribute")
	}
	b[0] = uint8(a.Type)
	b[1] = uint8(fullLen)
	copy(b[2:], a.Value)
	return fullLen, err
}

func (a AVP) ValueStr() string {
	return string(a.Value)
}
