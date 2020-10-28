package component

import (
	"component/auth"
	"config"
	"encoding/binary"
	"fmt"
	"github.com/extrame/radius"
	"github.com/wonderivan/logger"
	"net"
	"strings"
)

var START = 1
var STOP = 2
var UPDATE = 3

type AuthService struct{}

var radius_service *AuthService = new(AuthService)

//radius 认证、授权服务
func StartRadiusAuth() {
	logger.Info("listening auth on %d\n", *config.RadiusAuthPort)
	// 生成Server结构体返回，包含ip地址和密码
	s := radius.NewServer(fmt.Sprintf(":%d", *config.RadiusAuthPort), *config.RadiusSecret)
	s.RegisterService(radius_service)
	err := s.ListenAndServe()
	logger.Error("Auth Err:", err)
}

//radius计费
func StartRadiusAcc() {
	if *config.RadiusAccPort != *config.RadiusAuthPort {
		logger.Info("listening acc on %d\n", *config.RadiusAccPort)
		s := radius.NewServer(fmt.Sprintf(":%d", *config.RadiusAccPort), *config.RadiusSecret)
		s.RegisterService(radius_service)
		err := s.ListenAndServe()
		logger.Error("Acc Err:", err)
	}
}

func (p *AuthService) Authenticate(request *radius.Packet) (*radius.Packet, error) {
	var username, userpwd []byte
	var chapid byte
	var chappwd []byte
	var chapmod = false
	var callingStationId net.HardwareAddr
	var chapcha = request.Authenticator[:]
	var userip net.IP
	var acctStatus int
	var acctSessionId string
	var inputOctets, outputOctets uint32
	var acctSessionTime uint32

	for _, v := range request.AVPs {
		if v.Type == radius.UserName {
			username = v.Value
		} else if v.Type == radius.UserPassword {
			userpwd = v.Value
		} else if v.Type == radius.CHAPPassword {
			chapmod = true
			chapid = v.Value[0]
			chappwd = v.Value[1:]
		} else if v.Type == radius.CHAPChallenge {
			chapcha = v.Value
		} else if v.Type == radius.FramedIPAddress {
			userip, _ = v.IP()
		} else if v.Type == radius.CallingStationId {
			callingStationId, _ = v.Mac()
		} else if v.Type == radius.AcctStatusType {
			acctStatus, _ = v.Integer()
		} else if v.Type == radius.AcctSessionId {
			acctSessionId, _ = v.Text()
		} else if v.Type == radius.AcctOutputOctets {
			outputOctets, _ = v.Uinteger32()
		} else if v.Type == radius.AcctInputOctets {
			inputOctets, _ = v.Uinteger32()
		} else if v.Type == radius.AcctSessionTime {
			acctSessionTime, _ = v.Uinteger32()
		}

	}

	//AC向radius-server发起计费请求
	// Accounting-Request (4)
	if request.Code == radius.AccountingRequest {
		var err error
		if acctStatus == START {
			if service, ok := auth.ExtraAuth.(auth.RadiusAcctStartService); ok {
				err = service.AcctStart(username, userip, request.NasIP(), callingStationId, acctSessionId)
			} else {
				err = BASIC_SERVICE.AcctStart(username, userip, request.NasIP(), callingStationId, acctSessionId)
			}

		} else if acctStatus == STOP {
			if service, ok := auth.ExtraAuth.(auth.RadiusAcctStopService); ok {
				err = service.AcctStop(username, userip, request.NasIP(), callingStationId, acctSessionId)
			} else {
				err = BASIC_SERVICE.AcctStop(username, userip, request.NasIP(), callingStationId, acctSessionId)
			}

		} else if acctStatus == UPDATE {
			if service, ok := auth.ExtraAuth.(auth.RadiusAcctUpdateService); ok {
				err = service.AcctUpdate(username, userip, request.NasIP(), inputOctets, outputOctets, acctSessionTime)
			} else {
				err = BASIC_SERVICE.AcctUpdate(username, userip, request.NasIP(), inputOctets, outputOctets, acctSessionTime)
			}
		}

		//radius-server响应计费请求
		//AccountingResponse PacketCode = 5
		npac := request.Reply()
		npac.Code = radius.AccountingResponse
		text := "OK!"
		if err != nil {
			text = err.Error()
		}
		npac.AVPs = append(npac.AVPs, radius.AVP{Type: radius.ReplyMessage, Value: []byte(text)})
		return npac, nil
	}

	//AC向radius-server发起radius请求
	//AccessRequest      PacketCode = 1
	logger.Debug("ac(ip:%s) username:%s : radius request, code:%s ", request.NasIP().String(), string(username), request.Code.String())
	npac := request.Reply()
	msg := "ok!"
	var err = fmt.Errorf("unhandled")
	var timeout uint32
	//for mac test
	testedUserName := strings.Replace(callingStationId.String(), ":", "", -1)
	if strings.ToLower(string(username)) == testedUserName {
		logger.Info("Request to auth mac %s\n", testedUserName)
		if auth, ok := auth.ExtraAuth.(auth.MacAuthService); ok {
			err, timeout = auth.AuthMac(callingStationId, userip)
		} else {
			err, timeout = BASIC_SERVICE.AuthMac(callingStationId, userip)
		}
	}
	//for user name test
	if err != nil {
		if chapmod {
			if auth, ok := auth.ExtraAuth.(auth.ChapAuthService); ok {
				err, timeout = auth.AuthChap(username, chapid, chappwd, chapcha, userip, callingStationId)
			} else {
				err, timeout = BASIC_SERVICE.AuthChap(username, chapid, chappwd, chapcha, userip, callingStationId)
			}
		} else {
			if auth, ok := auth.ExtraAuth.(auth.PapAuthService); ok {
				err, timeout = auth.AuthPap(username, userpwd, userip)
			} else {
				err, timeout = BASIC_SERVICE.AuthPap(username, userpwd, userip)
			}
		}
	}

	if err == nil {
		if timeout != 0 {
			var to_bts = make([]byte, 4)
			binary.BigEndian.PutUint32(to_bts, timeout)
			npac.AVPs = append(npac.AVPs, radius.AVP{Type: radius.SessionTimeout, Value: to_bts})
		}
		//radius-server响应radius认证请求
		//AccessAccept       PacketCode = 2
		npac.Code = radius.AccessAccept
	} else {
		//radius-server响应radius认证请求
		//AccessReject       PacketCode = 3
		npac.Code = radius.AccessReject
		logger.Error("radius error ", err)
		msg = err.Error()
	}
	npac.AVPs = append(npac.AVPs, radius.AVP{Type: radius.ReplyMessage, Value: []byte(msg)})

	return npac, nil
}
