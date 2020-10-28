package component

import (
	"bytes"
	"component/auth"
	"component/mongodb"
	"config"
	"crypto/md5"
	"encoding/hex"
	"euht"
	"fmt"
	"net"
	"net/http"
	"strings"
	"utility"

	"github.com/wonderivan/logger"

	//"net/url"
	"path/filepath"
)

type AuthInfo struct {
	Name    []byte
	Pwd     []byte
	Mac     net.HardwareAddr
	Timeout uint32
}

type AuthServer struct {
	authing_user map[string]*AuthInfo
}

var BASIC_SERVICE = new(AuthServer)

var HTTP_PHONEAUTH_PORT = fmt.Sprintf("%d", *config.HttpPhoneAuthPort)

func InitBasic() {
	BASIC_SERVICE.authing_user = make(map[string]*AuthInfo)
}

func (a *AuthServer) AuthChap(username []byte, chapid byte, chappwd, chapcha []byte, userip net.IP, usermac net.HardwareAddr) (err error, to uint32) {
	//radius 协议
	logger.Debug("username:%s chap auth start", string(username))
	if auth.TestChapPwd(chapid, []byte(*config.RadiusSecret), chapcha, chappwd) == false {
		err = fmt.Errorf("password not match  %s", string(username))
		return
	}
	/*
		//EUHT 策略
		if euht.Auth(string(username)) == false {
			err = fmt.Errorf("euht auth fail  %s", string(username))
			return
		}*/
	logger.Debug("username:%s chap auth end", string(username))
	return nil, 0
}

func (a *AuthServer) AuthMac(mac net.HardwareAddr, userip net.IP) (error, uint32) {
	return fmt.Errorf("unsupported mac auth on %s", userip.String()), 0
}

func (a *AuthServer) AuthPap(username, userpwd []byte, userip net.IP) (err error, to uint32) {
	if info, ok := a.authing_user[userip.String()]; ok {
		if bytes.Compare(info.Pwd, userpwd) == 0 {
			to = info.Timeout
		}
	} else {
		err = fmt.Errorf("radius auth - no such user ", userip.String())
	}
	return
}

//重定向注册页面URL，发给用户URL访问
func (a *AuthServer) redirectRegisterPage(w http.ResponseWriter, r *http.Request) {
	index := strings.Index(r.RequestURI, "?")
	if index < 0 {
		index = 0
	}
	path := "/registerpage" + r.RequestURI[index:]
	http.Redirect(w, r, path, http.StatusFound)
}

//清除缓存用处
func (a *AuthServer) disableBrowerCache(w http.ResponseWriter) {
	//禁止浏览器缓存
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Cache-Control", "max-age=0")
	w.Header().Add("Cache-Control", "must-revalidate")
	w.Header().Add("Cache-Control", "private")
	w.Header().Add("Pragma", "no-cache")
}

//portal重定向登录路径处理信息
func (a *AuthServer) login(w http.ResponseWriter, r *http.Request) {
	var err error
	var relation mongodb.Relation
	//logger.Debug(r.RequestURI)
	//获取接入设备信息
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	//acName := r.URL.Query().Get("wlanacname")
	//acIp := r.URL.Query().Get("wlanacip")
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	_, port, _ := net.SplitHostPort(r.Host)
	wlanacip := r.URL.Query().Get("wlanacip")
	userip := net.ParseIP(userIP)
	acip := net.ParseIP(acIP)
	userPwd := []byte(string(*config.RadiusSecret))

	//logger.Info("ac(ip:%s) user(ip:%s mac:%s) : login request", acIP, userIP, userMac)

	defer func() {
		if err != nil {
			logger.Error("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s): login error : %s", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac, err.Error())
			path := filepath.FromSlash(*config.LoginFailPage)
			http.ServeFile(w, r, path)
		}
	}()

	if userip == nil {
		err = fmt.Errorf("user IP 配置错误")
		return
	}
	if acip == nil {
		err = fmt.Errorf("ac IP 配置错误")
		return
	}
	//根据端口号区分手机号绑定方式、Router绑定方式
	if port == HTTP_PHONEAUTH_PORT {
		relation.AcIp = acIP
		relation.ClientIp = userIP
		relation.ClientMac = userMac
		//为了代码更好兼容性，填充对应关系-Router信息
		relation.RouterSn = "SN1234567890"
		relation.RouterMac = "AA:BB:CC:DD:EE:FF"
		relation.RouterIp = "0.0.0.0"
		logger.Info("ac(ip:%s) user(ip:%s mac:%s)  login request", acIP, userIP, userMac)
		//查询绑定表，如果该手机未绑定就推送重定向注册页面
		if euht.IsClientRegistered(userMac) == false {
			logger.Warn("ac(ip:%s) user(ip:%s mac:%s) : not register", acIP, userIP, userMac)
			a.redirectRegisterPage(w, r)
			return
		}
	} else {
		relation, err = euht.GetRelationByClientMac(userMac)
		logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s) login request", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac)
		if err == nil {
			if euht.IsRealName(relation.RouterMac) == false {
				logger.Warn("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s) : not real name", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac)
				//生成校验码，防止用户修改url参数发起注册
				user_parameter := "nufront" + wlanacip + userIP + userMac + relation.RouterSn + relation.RouterMac
				hash := md5.New()
				hash.Write([]byte(user_parameter))
				checksum := hex.EncodeToString(hash.Sum(nil))
				r.RequestURI = r.RequestURI + "&routersn=" + relation.RouterSn + "&routermac=" + relation.RouterMac + "&checksum=" + checksum
				a.redirectRegisterPage(w, r)
				return
			}
		} else {
			err = nil
			relation.AcIp = acIP
			relation.ClientIp = userIP
			relation.ClientMac = userMac
		}
	}

	//发起Portal请求
	for i := 0; i < *config.CmccPortalTimes; i++ {
		err = Auth(userip, acip, uint32(0), []byte(userMac), userPwd)
		if err == nil {
			break
		} else {
			logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s) : auth fail %d times, %s", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac, i, err.Error())
		}
	}

	if err != nil {
		err = fmt.Errorf("auth fail: %s", err.Error())
		return
	}
	logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s) : login success", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac)
	path := filepath.FromSlash(*config.LoginSuccessPage)
	a.disableBrowerCache(w)
	http.ServeFile(w, r, path)
	//euht.RecordClientOnlineStatus(relation, true)

}

//处理主页请求
func (a *AuthServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	a.login(w, r)
}

//处理随机请求
func (a *AuthServer) HandleRandom(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}

//处理登录请求
func (a *AuthServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	a.login(w, r)
}

//处理验证码发送，根据前端发过来手机号
func (a *AuthServer) HandleSendValidateCode(w http.ResponseWriter, r *http.Request) {
	// 从用户发送的request中获取用户信息
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	user_phone_number := r.URL.Query().Get("user_phone_number")

	//根据acIP查询数据库得出对应的服务热线
	// acip为客户端的ip地址，config.ServicePhoneCollection记录着关系数据库的名称，GetServicePhoneByAcIP为获取对应关系的函数
	servicePhoneInfo, err := mongodb.Db_handler.GetServicePhoneByAcIP(*config.ServicePhoneCollection, acIP)
	service_phone_number := servicePhoneInfo.ServicePhone
	if err != nil || servicePhoneInfo.ServicePhone == "" {
		//查不到的ACIP，服务热线使用总机
		service_phone_number = *config.EuhtDefaultServicePhone
	}

	logger.Debug("ac(ip:%s) user(ip:%s mac:%s) : send sms code request : phone:%s service_phone_number(%s)", acIP, userIP, userMac, user_phone_number, service_phone_number)
	if euht.SendSms(user_phone_number, service_phone_number) == true {
		logger.Debug("ac(ip:%s) user(ip:%s mac:%s) : send sms code ok : phone:%s", acIP, userIP, userMac, user_phone_number)
		w.WriteHeader(http.StatusOK)
	} else {
		logger.Error("ac(ip:%s) user(ip:%s mac:%s) : send sms code fail : phone:%s", acIP, userIP, userMac, user_phone_number)
		w.WriteHeader(http.StatusBadRequest)
	}
}

//推送用户注册页面
func (a *AuthServer) HandleRegisterPage(w http.ResponseWriter, r *http.Request) {
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	logger.Info("ac(ip:%s) user(ip:%s mac:%s) : register page request", acIP, userIP, userMac)
	_, port, _ := net.SplitHostPort(r.Host)
	path := filepath.FromSlash(*config.RegisterNamePage)
	if port == HTTP_PHONEAUTH_PORT {
		path = filepath.FromSlash(*config.RegisterPhonePage)
	}
	// 清理浏览器缓存，禁止浏览器缓存
	a.disableBrowerCache(w)
	http.ServeFile(w, r, path)
}

//推送用户注册成功页面
func (a *AuthServer) HandleRegisterSuccess(w http.ResponseWriter, r *http.Request) {
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	logger.Info("ac(ip:%s) user(ip:%s mac:%s) : register success page request", acIP, userIP, userMac)
	path := filepath.FromSlash(*config.RegisterSucessPage)
	a.disableBrowerCache(w)
	http.ServeFile(w, r, path)
}

//处理用户注册提交信息
func (a *AuthServer) HandleRegister(w http.ResponseWriter, r *http.Request) {
	_, port, _ := net.SplitHostPort(r.Host)
	if port == HTTP_PHONEAUTH_PORT {
		a.registerUsePhone(w, r)
	} else {
		a.registerUsePhoneAndName(w, r)
	}
}

//处理用户注册(手机号绑定方式)提交信息
func (a *AuthServer) registerUsePhone(w http.ResponseWriter, r *http.Request) {
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	phone := r.URL.Query().Get("user_phone_number")
	captcha := r.URL.Query().Get("user_captcha")
	var err error
	err_str := "err_no=1"
	defer func() {
		if err != nil {
			logger.Error("ac(ip:%s) user(ip:%s mac:%s) : register phone fail : phone:%s captcha:%s : %s", acIP, userIP, userMac, phone, captcha, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err_str))
		}
	}()
	//校验用户填写验证码是否正确
	status, code := euht.QueryValidateCode(phone, captcha)
	if status == false {
		err = fmt.Errorf("sms code fail code:%d", code)
		if code == euht.VALIDCODE_NOT_FIND {
			err_str = "err_no=6"
		} else {
			err_str = "err_no=1"
		}
		return
	}

	//查询数据库，是否一个手机号码绑定超过3个设备
	if euht.IsMaxRegisterDevicesByPhone(phone) == true {
		err = fmt.Errorf("the register devices of this phone is full")
		err_str = "err_no=7"
		return
	}

	//验证码正常，把用户acip，mac, phone写入数据库
	if euht.RegisterClientInfo(acIP, userMac, phone) == false {
		err = fmt.Errorf("register client info fail")
		err_str = "err_no=3"
		return
	}
	//删掉该手机号码验证码数据库
	euht.RemoveValidateCode(phone)
	logger.Info("ac(ip:%s) user(ip:%s mac:%s) :  register phone success : phone:%s  captcha:%s", acIP, userIP, userMac, phone, captcha)
	a.disableBrowerCache(w)
	w.WriteHeader(http.StatusOK)
}

//处理用户注册(Router绑定方式)提交信息
func (a *AuthServer) registerUsePhoneAndName(w http.ResponseWriter, r *http.Request) {
	userIP := r.URL.Query().Get("wlanuserip")
	userMac := strings.ToUpper(r.URL.Query().Get("usermac"))
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	wlanacip := r.URL.Query().Get("wlanacip")
	phone := r.URL.Query().Get("user_phone_number")
	name := r.URL.Query().Get("user_true_name")
	captcha := r.URL.Query().Get("user_captcha")
	routerSn := r.URL.Query().Get("routersn")
	routerMac := strings.ToUpper(r.URL.Query().Get("routermac"))
	checksum := r.URL.Query().Get("checksum")
	hideName := utility.HideName(name)
	hidePhone := utility.HidePhone(phone)
	//0:ok 1:sms fail 2:realname fail, 3: save info to db fail
	var err error
	var relation mongodb.Relation
	err_str := "err_no=1"
	defer func() {
		if err != nil {
			logger.Error("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s): register fail : phone:%s name:%s captcha:%s : %s", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac, hidePhone, hideName, captcha, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err_str))
		}
	}()
	//没找到对应关系时，为让用户成功注册，从url中获取注册信息，生成新的校验码，与url中校验码进行对比，
	// 防止用户修改url参数
	relation, err = euht.GetRelationByClientMac(userMac)
	if err != nil {
		user_parameter := "nufront" + wlanacip + userIP + userMac + routerSn + routerMac
		// 实例化一个hash.Hash对象
		hash := md5.New()
		// 将要加密的数据写入到底层数据量
		hash.Write([]byte(user_parameter))
		// Sum 函数是对hash.Hash对象内部存储的内容进行校验和 计算然后将其追加到data的后面形成一个新的byte切片。
		// 因此通常的使用方法就是将data置为nil。
		// 该方法返回一个Size大小的byte数组，对于MD5来说就是一个128bit的16字节byte数组。
		new_checksum := hex.EncodeToString(hash.Sum(nil))

		if new_checksum == checksum {
			relation.RouterSn = routerSn
			relation.RouterMac = routerMac
			relation.AcIp = wlanacip
		} else {
			err = fmt.Errorf("checksum is not match")
			err_str = "err_no=3"
			return
		}

	}

	logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s):  register request : phone:%s name:%s captcha:%s", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac, hidePhone, hideName, captcha)
	//校验用户填写验证码是否正确
	status, code := euht.QueryValidateCode(phone, captcha)
	if status == false {
		err = fmt.Errorf("sms code fail code:%d", code)
		if code == euht.VALIDCODE_NOT_FIND {
			err_str = "err_no=6"
		} else {
			err_str = "err_no=1"
		}
		return
	}

	//查询实名数据库，是否超过一个手机号码绑定3个Router
	if euht.IsMaxRealRoutersByPhone(phone) == false {
		err = fmt.Errorf("real name full")
		err_str = "err_no=5"
		return
	}

	//每天限制实名认证次数，已节省费用
	if euht.IsMaxRealnameCountsByPhone(name, phone, relation) == false {
		err = fmt.Errorf("real name auth: today count full")
		err_str = "err_no=4"
		return
	}
	//手机二元素验证-实名验证
	ret, _ := euht.RealnameAuth(name, phone)
	if ret != true {
		err = fmt.Errorf("real name auth fail")
		err_str = "err_no=2"
		return
	}
	//用户通过验证码、实名验证通过，把用户信息填写到数据库
	if euht.RegisterUserInfo(relation, name, phone) == false {
		err = fmt.Errorf("register userinfo fail")
		err_str = "err_no=3"
		return
	}
	//删掉该手机号码验证码数据库
	euht.RemoveValidateCode(phone)

	logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s):  register success : phone:%s name:%s captcha:%s", acIP, userIP, userMac, relation.RouterSn, relation.RouterMac, hidePhone, hideName, captcha)
	a.disableBrowerCache(w)
	w.WriteHeader(http.StatusOK)
}

//portal通知用户下线处理函数
func (a *AuthServer) NotifyLogout(userip, acip net.IP) error {
	callBackOffline(*config.CallBackUrl, userip, acip)
	//根据用户IP地址，查找对应关系
	relation, err := euht.GetRelationByClientIP(userip.String())
	if err != nil {
		relation.AcIp = acip.String()
		relation.ClientIp = userip.String()
	}
	logger.Info("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac%s) : notify logout", relation.AcIp, relation.ClientIp, relation.ClientMac, relation.RouterSn, relation.RouterMac)
	//记录portal认证过上下线数据
	//euht.RecordClientOnlineStatus(relation, false)
	return nil
}

//处理Logout请求
func (a *AuthServer) HandleLogout(w http.ResponseWriter, r *http.Request) {
	var err error
	var relation mongodb.Relation
	acIP, _, _ := net.SplitHostPort(r.RemoteAddr)
	userip_str := r.FormValue("userip")
	defer func() {
		if err != nil {
			logger.Error("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s): logout fail : %s", acIP, userip_str, relation.ClientMac, relation.RouterSn, relation.RouterMac, err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
	}()

	userip := net.ParseIP(userip_str)
	if userip == nil {
		err = fmt.Errorf("Parse Ip err from %s", userip_str)
		return
	}

	acip := net.ParseIP(acIP)
	if acip == nil {
		err = fmt.Errorf("Parse Ip err from %s", acIP)
		return
	}
	//根据ip地址查找对应关系
	relation, err = euht.GetRelationByClientIP(userip_str)
	if err != nil {
		err = nil
		relation.AcIp = acIP
		relation.ClientIp = userip_str
	}
	_, err = Logout(userip, *config.CmccSecret, acip)
	if err != nil {
		return
	}
	logger.Error("ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s): logout success ", acIP, userip_str, relation.ClientMac, relation.RouterSn, relation.RouterMac)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("logout ok"))
	//euht.RecordClientOnlineStatus(relation, false)

}

//处理AC向radius-server发起计费请求
// Accounting-Request (4)
func (a *AuthServer) AcctStart(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error {
	return nil
}
func (a *AuthServer) AcctStop(username []byte, userip net.IP, nasip net.IP, usermac net.HardwareAddr, sessionid string) error {
	callBackOffline(*config.CallBackUrl, userip, nasip)
	return nil
}
func (a *AuthServer) AcctUpdate(username []byte, userip net.IP, nasip net.IP, inputoctets, outputoctets, acctsessiontime uint32) error {
	//logger.Debug(string(username), inputoctets, outputoctets, acctsessiontime)
	//_ = mongodb.Db_handler.UpdateClientAcctInfo(string(username), inputoctets, outputoctets, acctsessiontime)
	return nil
}

//更新数据库RSA加密-密钥处理
func (a *AuthServer) HandleUpdateEncryptKey(w http.ResponseWriter, r *http.Request) {
	//logger.Debug(r.Method)
	if r.Method == "PUT" {
		r.ParseForm()
		publicKey := r.FormValue("public_key")
		privateKey := r.FormValue("private_key")
		//to do
		// md5 checkout
		// close register server
		go euht.UpdateRSAKey(publicKey, privateKey)
		w.WriteHeader(http.StatusOK)
	} else {
		w.Header().Add("update_status", fmt.Sprintf("%d", euht.UpdateRSAKeyStatus))
		w.WriteHeader(http.StatusOK)
		if euht.UpdateRSAKeyStatus != euht.UpdateRSAKeyStart ||
			euht.UpdateRSAKeyStatus != euht.UpdateRSAKeySuccess {
			euht.UpdateRSAKeyStatus = euht.UpdateRSAKeyInitStatus
		}
	}

}

//县级设备、网管监控portal系统http服务
func (a *AuthServer) HandleMonitor(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}
