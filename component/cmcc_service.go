package component

import (
	"component/cmcc/portal"
	"component/cmcc/portal/v1"
	"config"
	"fmt"
	"github.com/wonderivan/logger"
	"net"
	"net/http"
)

func StartCmcc() {
	portal.RegisterFallBack(func(msg portal.Message, src net.IP) {
		logger.Debug("type:", msg.Type())
		//Notify  LOGOUT   客户退出登录
		//portal.NTF_LOGOUT = 8
		if msg.Type() == portal.NTF_LOGOUT {		
			//portal通知用户下线处理函数
			BASIC_SERVICE.NotifyLogout(msg.UserIp(), src)
		}
	})

	//.cmcc版本
	if *config.CmccVersion == 1 {
		portal.SetVersion(new(v1.Version))
	} else {
		portal.SetVersion(new(v1.Version))
	}

	//监听CmccPortalPort
	//监听并且读取udp信息,监听
	//侦听并接收认证服务器的端口数据
	err := portal.ListenAndService(fmt.Sprintf(":%d", *config.CmccPort))
	if err != nil {
		logger.Error("cmcc server fail : %s", err.Error())
	}
}

//Notify  LOGOUT   客户退出登录
func callBackOffline(url string, userip, netip net.IP) {
	if url != "" {
		if resp, err := http.Get(url + "?userip=" + userip.String() + "&nas=" + netip.String()); err == nil {
			defer resp.Body.Close()
		}
	}
}

//发起Portal请求的全过程
//包括challenge和radius认证请求
func Auth(userip net.IP, basip net.IP, timeout uint32, username, userpwd []byte) (err error) {
	var res portal.Message
	//向AC请求Challenge
	logger.Debug("ac(ip:%s) user(ip:%s username:%s) : Cmcc portal Challenge start", basip.String(), userip.String(), string(username))
	//logger.Debug("userip:%s username:%s ac:%s  Cmcc portal Challenge start", userip, username, basip)
	if res, err = Challenge(userip, basip); err == nil {
		if cres, ok := res.(portal.ChallengeRes); ok {//如果路由已经分配好了Challenge
			logger.Debug("ac(ip:%s) user(ip:%s username:%s) : Cmcc portal ChapAuth start", basip.String(), userip.String(), string(username))

			//logger.Debug("userip:%s username:%s ac:%s  Cmcc portal ChapAuth start", userip, username, basip)
			//向AC发起radius认证请求
			res, err = portal.ChapAuth(userip, *config.CmccSecret, basip, *config.CmccNasPort, username, userpwd, res.ReqId(), cres.GetChallenge())
			//			if err == nil {
			//				res, err = portal.AffAckAuth(userip, *config.CmccSecret, basip, *config.CmccNasPort, res.SerialId(), res.ReqId())
			//			}
			if err != nil {
				err = fmt.Errorf("Cmcc portal ChapAuth fail: %s", err.Error())
				//logger.Error("ac(ip:%s) user(ip:%s username:%s) : Cmcc portal ChapAuth fail : %s", basip.String(), userip.String(), string(username), err.Error())
			}
		}
	} else if err.Error() == "No. 2:此链接已建立" {
		err = nil
		logger.Debug("ac(ip:%s) user(ip:%s username:%s) : 此链接已建立", basip.String(), userip.String(), string(username))
	} else {
		err = fmt.Errorf("Cmcc portal Challenge fail: %s", err.Error())
		//logger.Error("ac(ip:%s) user(ip:%s username:%s) : Cmcc portal Challenge fail : %s", basip.String(), userip.String(), string(username), err.Error())
	}
	return
}

//向AC请求Challenge
func Challenge(userip net.IP, basip net.IP) (response portal.Message, err error) {
	return portal.Challenge(userip, *config.CmccSecret, basip, *config.CmccNasPort)
}

//处理Logout请求
func Logout(userip net.IP, secret string, basip net.IP) (response portal.Message, err error) {
	return portal.Logout(userip, *config.CmccSecret, basip, *config.CmccNasPort)
}
