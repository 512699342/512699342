package auth

import (
	"crypto/md5"
	"github.com/wonderivan/logger"
	"net"
	"net/http"
	"runtime/debug"
)

type ChapAuthService interface {
	AuthChap(username []byte, chapid byte, chappwd, chapcha []byte, userip net.IP, usermac net.HardwareAddr) (error, uint32)
}

type PapAuthService interface {
	AuthPap(username, userpwd []byte, userip net.IP) (error, uint32)
}

type MacAuthService interface {
	AuthMac(mac net.HardwareAddr, userip net.IP) (error, uint32)
}

//通过http方式请求Login
type HttpLoginHandler interface {
	HandleLogin(w http.ResponseWriter, r *http.Request)
}

//通过http方式发送验证码
type HttpSendValidateCodeHandler interface {
	HandleSendValidateCode(w http.ResponseWriter, r *http.Request)
}

type HttpRegisterPageHandler interface {
	HandleRegisterPage(w http.ResponseWriter, r *http.Request)
}

//通过http方式注册
type HttpRegisterHandler interface {
	HandleRegister(w http.ResponseWriter, r *http.Request)
}

//通过http方式请求Logout
type HttpLogoutHandler interface {
	HandleLogout(w http.ResponseWriter, r *http.Request)
}

type HttpRootHandler interface {
	HandleRoot(w http.ResponseWriter, r *http.Request)
}

//通过该接口监听更多的http方法
type ExtraHttpHandler interface {
	AddExtraHttp()
}

type RadiusAcctStartService interface {
	AcctStart(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error
}

type RadiusAcctStopService interface {
	AcctStop(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error
}

type RadiusAcctUpdateService interface {
	AcctUpdate(username []byte, userip net.IP, nasip net.IP, inputoctets, outputoctets, acctsessiontime uint32) error
}

var ExtraAuth interface{}

//UTILS for wrap the http error
func ErrorWrap(w http.ResponseWriter) {
	if e := recover(); e != nil {
		logger.Error("panic:", e, "\n", string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
		if err, ok := e.(error); ok {
			w.Write([]byte(err.Error()))
		}
	}
}

//utils function to test agent chap password
func TestChapPwd(chapid byte, testedpwd, chapcha, chappwd []byte) bool {
	hash := md5.New()
	hash.Write([]byte{chapid})
	hash.Write(testedpwd)
	hash.Write(chapcha)
	tested := hash.Sum(nil)
	for i := 0; i < len(tested); i++ {
		if tested[i] != chappwd[i] {
			return false
		}
	}
	return true
}
