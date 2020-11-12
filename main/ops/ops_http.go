package ops

import (
	"component"
	"component/Sessions"
	"component/mongodb"
	"config"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"github.com/wonderivan/logger"
	"math"
	"net"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var sessionMgr *Sessions.SessionMgr = nil

func init() {
	sessionMgr = Sessions.NewSessionMgr("ops_cookie_name", 3600*6)

	//查询admin账号是否已设置
	info, err := mongodb.Db_handler.GetOpsUserInfoByName(*config.DbOpsUserInfoCollection, "admin")
	if err != nil {
		result_string := ""
		logger.Error("ops admin password init(username:%s) %s", "admin", err.Error())

		if info.UserName == "" {
			//将用户数据写入数据库
			userInfo := mongodb.OpsUserInfo{
				UserName:     "admin",
				UserPassword: GetMD5Encode(*config.DbOpsAdminPassword),
				FullName:     "管理员",
				UserPhone:    "",
				UserEmail:    "",
			}

			err = mongodb.Db_handler.UpsertOpsUserInfo(*config.DbOpsUserInfoCollection, userInfo)
			if err != nil {
				result_string = err.Error()
				logger.Info("ops admin password init:", result_string)
			} else {
				//添加成功
				result_string = "initial admin password success！"
				logger.Info("ops admin password init:", result_string)
			}
		}
	}
}

//禁止缓存
func DisableBrowserCache(w http.ResponseWriter) {
	//禁止浏览器缓存
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Cache-Control", "max-age=0")
	w.Header().Add("Cache-Control", "must-revalidate")
	w.Header().Add("Cache-Control", "private")
	w.Header().Add("Pragma", "no-cache")
}

//返回一个32位md5加密后的字符串
func GetMD5Encode(data string) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

type OPSServer struct {
}

type clientInfo struct {
	Mac          string    `json:"mac"`
	Phone        string    `json:"phone"`
	AcIP         string    `json:"acip"`
	RegsiterTime time.Time `json:"registertime"`
}
type clientInfos struct {
	Result []clientInfo `json:"result"`
}

type userInfo struct {
	UserName     string `bson:"userName"`
	UserPassword string `bson:"userPassword"`
	FullName     string `bson:"fullName"`
	UserPhone    string `bson:"userPhone"`
	UserEmail    string `bson:"userEmail"`
}
type userInfos struct {
	Result []userInfo `json:"result"`
}

//网管组http get统计接口数据库
type monthData struct {
	Month       string `json:"month"`
	ClientNum   int    `json:"clientNum"`
	ClientTotal int    `json:"clientTotal"`
}

type clientBindInfo struct {
	AreaParttion string      `json:"area"`
	MonthDatas   []monthData `json:"monthData"`
}

type responseBindInfo struct {
	Error           int              `json:"error"`
	Msg             string           `json:"msg"`
	ClientBindInfos []clientBindInfo `json:"clientBindInfo"`
}

//网管组http get统计接口数据库

func (ops *OPSServer) HandleRoot(w http.ResponseWriter, r *http.Request) {
	//  logger.Info(r.RequestURI)
	DisableBrowserCache(w)
	if r.RequestURI == "/" {
		path := filepath.FromSlash(*config.OpsLoginPage)
		http.ServeFile(w, r, path)
	} else if strings.Index(r.RequestURI, "/index") == 0 {

		var sessionID = sessionMgr.CheckCookieValid(w, r)

		if sessionID == "" {
			path := filepath.FromSlash(*config.OpsLoginPage)
			http.ServeFile(w, r, path)
			return
		}
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		path := filepath.FromSlash(*config.OpsIndexPage)
		http.ServeFile(w, r, path)
	} else if strings.Index(r.RequestURI, "/add_user.html") == 0 {
		var sessionID = sessionMgr.CheckCookieValid(w, r)

		if sessionID == "" {
			path := filepath.FromSlash(*config.OpsLoginPage)
			http.ServeFile(w, r, path)
			return
		}

		val, ok := sessionMgr.GetSessionVal(sessionID, "session_username")
		if ok {
			if val != "admin" {
				return
			}
		} else {
			return
		}
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		path := filepath.FromSlash(*config.OpsAddUserPage)
		http.ServeFile(w, r, path)
	} else if strings.Index(r.RequestURI, "/user_list.html") == 0 {
		var sessionID = sessionMgr.CheckCookieValid(w, r)

		if sessionID == "" {
			path := filepath.FromSlash(*config.OpsLoginPage)
			http.ServeFile(w, r, path)
			return
		}

		val, ok := sessionMgr.GetSessionVal(sessionID, "session_username")
		if ok {
			if val != "admin" {
				return
			}
		} else {
			return
		}
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		path := filepath.FromSlash(*config.OpsUserListPage)
		http.ServeFile(w, r, path)
	} else if strings.Contains(r.RequestURI, "alter_user.html") {
		var sessionID = sessionMgr.CheckCookieValid(w, r)

		if sessionID == "" {
			path := filepath.FromSlash(*config.OpsLoginPage)
			http.ServeFile(w, r, path)
			return
		}
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		path := filepath.FromSlash(*config.OpsAlterUserPage)
		http.ServeFile(w, r, path)
	}

}

//like this  get:  http://192.168.22.123:18070/query_phone_num?mac=aa:bb:cc:dd:ee:00&mac=aa:bb:cc:dd:ee:01&mac=aa:bb:cc:dd:ee:02
//return application/json or  text/plain
func (ops *OPSServer) HandleQueryPhoneNum(w http.ResponseWriter, r *http.Request) {
	//logger.Info(r.RequestURI)
	r.ParseForm()
	macs := r.Form["mac"]
	accept := r.Header.Get("accept")

	if strings.Contains(accept, "json") {
		//return application/json
		var clientInfos clientInfos
		for _, mac := range macs {
			var item clientInfo
			item.Mac = mac
			info, err := mongodb.Db_handler.GetClientinfoByMac(strings.ToUpper(mac))
			if err != nil {
				item.Phone = ""
				item.AcIP = ""
				item.RegsiterTime = time.Unix(0, 0)
			} else {
				item.Phone = info.Phone
				item.AcIP = info.AcIp
				item.RegsiterTime = info.BindTime
			}
			clientInfos.Result = append(clientInfos.Result, item)
		}
		w.Header().Add("content-type", "application/json")
		b, err := json.Marshal(clientInfos)
		if err != nil {
			logger.Error("marshal fail: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(b))
		}
	} else {
		// return text/plain
		var s string
		for _, mac := range macs {
			info, err := mongodb.Db_handler.GetClientinfoByMac(strings.ToUpper(mac))
			if err != nil {
				s = s + mac + "=,"
			} else {
				s = s + mac + "=" + info.Phone + ","
			}
		}
		len := len(s)
		s = s[:len-1]
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(s))
	}

}

//手机号注册信息查询
func (ops *OPSServer) HandleQueryPhoneRegisterInfo(w http.ResponseWriter, r *http.Request) {
	var clientInfos clientInfos
	var err error
	var b []byte
	errString := ""
	user_phone_number := r.URL.Query().Get("user_phone_number")

	defer func() {

		if err != nil {
			logger.Info("[Phone register search] phone(%s) result(%s)", user_phone_number, errString)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errString))
		}
	}()

	infos, err := mongodb.Db_handler.GetClientinfoByPhoneAll(*config.DbClientInfoCollection, user_phone_number)
	if err != nil {
		errString = err.Error()
		return
	}

	//组包json格式数据 ClientInfo
	for _, info := range infos {
		clientInfos.Result = append(clientInfos.Result, clientInfo{
			Mac:          info.ClientMac,
			Phone:        info.Phone,
			AcIP:         info.AcIp,
			RegsiterTime: info.BindTime,
		})
	}

	b, err = json.Marshal(clientInfos)
	if err != nil {
		errString = err.Error()
		return
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))
}

//检查瑞斯康达cmcc和radius状态
func (ops *OPSServer) HandleQueryAcStatus(w http.ResponseWriter, r *http.Request) {
	var err error
	var result_string string
	acIP := r.URL.Query().Get("acip")
	//经测试，只要ip和mac不在瑞斯康达在线列表，就对瑞斯康达没有影响
	userIP := "1.1.1.1"
	userMac := "AA:BB:CC:11:22:33"
	acip := net.ParseIP(acIP)
	userip := net.ParseIP(userIP)
	userPwd := []byte(string(*config.RadiusSecret))

	defer func() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result_string))
		logger.Info("[ACStatus] ac(ip:%s) user(ip:%s mac:%s) result(%s): auth  finish", acIP, userIP, userMac, result_string)
	}()

	logger.Info("[ACStatus] ac(ip:%s) user(ip:%s mac:%s) : auth  start", acIP, userIP, userMac)

	if acip == nil {
		result_string = "ac IP 配置错误"
		return
	}
	_, err = mongodb.Db_handler.GetClientinfoByAcIP(*config.DbClientInfoCollection, acIP)
	if err != nil {
		result_string = "数据库查不到此ac IP"
		return
	}
	for i := 0; i < *config.CmccPortalTimes; i++ {
		err = component.Auth(userip, acip, uint32(0), []byte(userMac), userPwd)
		if err == nil {
			break
		} else {
			logger.Error("[ACStatus] ac(ip:%s) user(ip:%s mac:%s) : auth fail %d times, %s", acIP, userIP, userMac, i, err.Error())
		}
	}
	if err == nil {
		result_string = "正常" //正常
	} else if strings.Contains(err.Error(), "Challenge fail") {
		result_string = "路由器 cmcc Portal异常" //Portal异常
	} else if strings.Contains(err.Error(), "ChapAuth fail") {
		result_string = "路由器 radius 异常" //radius异常
	} else {
		result_string = "timeout"
	}

}

//查询用户注册总数
func (ops *OPSServer) HandleQueryRegisterCount(w http.ResponseWriter, r *http.Request) {
	result_string := "0"
	defer func() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result_string))
		logger.Info("[QueryRegisterCount] result(%s).", result_string)
	}()

	count, err := mongodb.Db_handler.GetCollectionCount(*config.DbClientInfoCollection)
	if err != nil {
		logger.Error("[QueryRegisterCount] fail：%s .", err.Error())
		return
	}

	result_string = strconv.Itoa(count)
}

//用户登录处理
func (ops *OPSServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	result_string := ""
	username := r.URL.Query().Get("name")
	password := r.URL.Query().Get("password")

	defer func() {
		DisableBrowserCache(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result_string))
		logger.Info("[HandleLogin](username:%s) result(%s).", username, result_string)
	}()

	//根据用户名向数据库查询密码
	userinfo, err := mongodb.Db_handler.GetOpsUserInfoByName(*config.DbOpsUserInfoCollection, username)
	if err != nil {
		logger.Error("HandleLogin(username:%s) fail: %s", username, err.Error())
		result_string = err.Error()
		return
	}

	if (password) == userinfo.UserPassword {
		//密码正确
		var sessionID = sessionMgr.StartSession(w, r)
		var loginUserInfo = ""

		loginUserInfo = username

		//踢除重复登录的
		var onlineSessionIDList = sessionMgr.GetSessionIDList()
		for _, onlineSessionID := range onlineSessionIDList {
			if userInfo, ok := sessionMgr.GetSessionVal(onlineSessionID, "session_username"); ok {

				if loginUserInfo == userInfo {
					sessionMgr.EndSessionByID(onlineSessionID)
				}

			}
		}
		onlineSessionIDList = sessionMgr.GetSessionIDList()

		//设置变量值
		sessionMgr.SetSessionVal(sessionID, "session_username", loginUserInfo)

		result_string = "UserPassword correct"
	} else {
		result_string = "UserPassword error"
	}

}

//用户退出登录处理
func (ops *OPSServer) HandleLoginOut(w http.ResponseWriter, r *http.Request) {

	var sessionID = sessionMgr.CheckCookieValid(w, r)
	result_string := ""
	operator, ok := sessionMgr.GetSessionVal(sessionID, "session_username")

	defer func() {
		logger.Info("[HandleLoginOut]{%s} result(%s).", operator, result_string)
	}()

	if !ok {
		result_string = "sessionID invalid"
		return
	}

	if sessionID == "" {
		path := filepath.FromSlash(*config.OpsLoginPage)
		http.ServeFile(w, r, path)
		result_string = "sessionID invalid"
		return
	}

	sessionMgr.EndSessionByID(sessionID)

	result_string = "loginout success"
	DisableBrowserCache(w)
	sessionMgr.UpdateBrowserCookie(w, sessionID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result_string))
}

//添加用户
func (ops *OPSServer) HandleAddUser(w http.ResponseWriter, r *http.Request) {

	user_name := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	full_name := r.URL.Query().Get("fullname")
	user_phone := r.URL.Query().Get("userphone")
	user_email := r.URL.Query().Get("useremail")

	result_string := ""
	// err := ""
	var sessionID = sessionMgr.CheckCookieValid(w, r)
	operator, _ := sessionMgr.GetSessionVal(sessionID, "session_username")

	defer func() {
		DisableBrowserCache(w)
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result_string))
		logger.Info("[HandleAddUser]{%s}(username:%s) result(%s).", operator, user_name, result_string)
	}()

	if operator != "admin" {
		result_string = "AddUser failed, operator not admin"
		return
	}

	//查询数据库user_name账号是否已存在
	_, err := mongodb.Db_handler.GetOpsUserInfoByName(*config.DbOpsUserInfoCollection, user_name)
	if err == nil {
		result_string = "user already exists"
		return
	}

	//将用户数据写入数据库
	userInfo := mongodb.OpsUserInfo{
		UserName:     user_name,
		UserPassword: GetMD5Encode(password),
		FullName:     full_name,
		UserPhone:    user_phone,
		UserEmail:    user_email,
	}

	err = mongodb.Db_handler.UpsertOpsUserInfo(*config.DbOpsUserInfoCollection, userInfo)
	if err != nil {
		logger.Error("[HandleAddUser]{%s} UpsertOpsUserInfo (username:%s) result(%s)", operator, user_name, err.Error())
		result_string = err.Error()
		return
	} else {
		//添加成功
		result_string = "add_user success"
	}

}

//删除用户
func (ops *OPSServer) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {

	user_name := r.URL.Query().Get("username")
	result_string := ""

	var sessionID = sessionMgr.CheckCookieValid(w, r)
	operator, _ := sessionMgr.GetSessionVal(sessionID, "session_username")

	defer func() {
		DisableBrowserCache(w)
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result_string))
		logger.Info("[HandleDeleteUser]{%s}(username:%s) result(%s).", operator, user_name, result_string)
	}()

	if operator != "admin" {
		result_string = "DeleteUser failed, operator not admin"
		return
	}

	if user_name == "" {
		logger.Error("[HandleDeleteUser]{%s} user_name:%s.", operator, user_name)
		result_string = "username is null"
		return
	}

	err := mongodb.Db_handler.RemoveOpsUserInfoByName(*config.DbOpsUserInfoCollection, user_name)
	if err != nil {
		logger.Error("[HandleDeleteUser]{%s}RemoveOpsUserInfoByName (user_name:%s ) fail: %s", operator, user_name, err.Error())
		result_string = err.Error()
		return
	} else {
		result_string = "delete user ok"
	}
}

//用户列表
func (ops *OPSServer) HandleGetUsersList(w http.ResponseWriter, r *http.Request) {

	var userInfos userInfos
	var err error
	var b []byte
	errString := ""

	var sessionID = sessionMgr.CheckCookieValid(w, r)
	operator, _ := sessionMgr.GetSessionVal(sessionID, "session_username")

	defer func() {

		if err != nil {
			logger.Info("[HandleGetUsersList]{%s} result(%s)", operator, errString)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errString))
		}
	}()

	infos, err := mongodb.Db_handler.GetOpsUserInfos(*config.DbOpsUserInfoCollection)
	if err != nil {
		errString = err.Error()
		return
	}

	//组包json格式数据 userInfos
	for _, info := range infos {
		userInfos.Result = append(userInfos.Result, userInfo{
			UserName:     info.UserName,
			UserPassword: info.UserPassword,
			FullName:     info.FullName,
			UserPhone:    info.UserPhone,
			UserEmail:    info.UserEmail,
		})
	}

	b, err = json.Marshal(userInfos)
	if err != nil {
		errString = err.Error()
		return
	}

	// logger.Info("[HandleGetUsersList]:%s.", b)
	DisableBrowserCache(w)
	sessionMgr.UpdateBrowserCookie(w, sessionID)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))

}

//获取某个用户的信息
func (ops *OPSServer) HandleGetUserInfo(w http.ResponseWriter, r *http.Request) {

	var userInfos userInfos
	var err error
	var b []byte
	errString := ""
	user_name := r.URL.Query().Get("username")

	var sessionID = sessionMgr.CheckCookieValid(w, r)
	operator, _ := sessionMgr.GetSessionVal(sessionID, "session_username")

	defer func() {

		if err != nil {
			logger.Info("[HandleGetUserInfo]{%s} result(%s)", operator, errString)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errString))
		}
	}()

	info, err := mongodb.Db_handler.GetOpsUserInfoByName(*config.DbOpsUserInfoCollection, user_name)
	if err != nil {
		errString = err.Error()
		return
	}

	//组包json格式数据 userInfos
	userInfos.Result = append(userInfos.Result, userInfo{
		UserName:     info.UserName,
		UserPassword: info.UserPassword,
		FullName:     info.FullName,
		UserPhone:    info.UserPhone,
		UserEmail:    info.UserEmail,
	})

	b, err = json.Marshal(userInfos)
	if err != nil {
		errString = err.Error()
		return
	}

	// logger.Info("[HandleGetUserInfo]:%s.", b)
	DisableBrowserCache(w)
	sessionMgr.UpdateBrowserCookie(w, sessionID)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(b))
}

//修改某个用户的信息
func (ops *OPSServer) HandleAlterUserInfo(w http.ResponseWriter, r *http.Request) {

	user_name := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")
	oldpassword := r.URL.Query().Get("oldpassword")
	full_name := r.URL.Query().Get("fullname")
	user_phone := r.URL.Query().Get("userphone")
	user_email := r.URL.Query().Get("useremail")
	operator := r.URL.Query().Get("operator")

	var sessionID = sessionMgr.CheckCookieValid(w, r)
	session_operator, _ := sessionMgr.GetSessionVal(sessionID, "session_username")

	logger.Debug("[HandleAlterUserInfo]", user_name, full_name, user_phone, user_email, operator)
	var err error
	errString := ""

	defer func() {

		if err != nil {
			logger.Error("[HandleAlterUserInfo]{%s} result(%s)", session_operator, errString)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errString))
		}
	}()

	//操作者校验
	if !(operator == "admin" || operator == user_name) {
		//非法操作
		logger.Error("[HandleAlterUserInfo]{%s}非法操作 : (user_name:%s operator:%s)", session_operator, user_name, operator)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alter_user fail"))
		return
	}

	info, err := mongodb.Db_handler.GetOpsUserInfoByName(*config.DbOpsUserInfoCollection, operator)
	if err != nil {
		logger.Error("[HandleAlterUserInfo]{%s}GetOpsUserInfoByName (operator:%s) alter fail: %s", session_operator, operator, err.Error())
		errString = "System error!"
		return
	}

	if info.UserPassword != GetMD5Encode(oldpassword) {
		logger.Error("[HandleAlterUserInfo]{%s}: (UserPassword != oldpassword) alter fail", session_operator)
		errString = "oldpassword error"
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(errString))
		return
	}

	//将用户数据写入数据库
	userInfo := mongodb.OpsUserInfo{
		UserName:     user_name,
		UserPassword: GetMD5Encode(password),
		FullName:     full_name,
		UserPhone:    user_phone,
		UserEmail:    user_email,
	}

	err = mongodb.Db_handler.UpsertOpsUserInfo(*config.DbOpsUserInfoCollection, userInfo)
	if err != nil {
		logger.Error("[HandleAlterUserInfo]{%s}(user_name:%s full_name:%s) alter fail: %s", session_operator, user_name, full_name, err.Error())
		errString = "System error!"
	} else {
		logger.Debug("[HandleAlterUserInfo]{%s}(user_name:%s full_name:%s) alter success", session_operator, user_name, full_name)
		//添加成功
		DisableBrowserCache(w)
		sessionMgr.UpdateBrowserCookie(w, sessionID)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alter_user success"))
	}
}

//计算HMAC
func gethmacsha1(sha1key string, content string) string {

	key := []byte(sha1key)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(content))

	hmac_sha1 := mac.Sum(nil)

	string_hmac_sha1 := base64.StdEncoding.EncodeToString(hmac_sha1)

	return string_hmac_sha1
}

/*
//测试案例
http://172.16.34.101:8070/phone_bind_num?begindate=201908&enddate=202001&timestamp=1555581203&signature=11
//计算HMAC SHA1 网站   begindate + enddate  +  timestamp  通过以下网站计算，KEY=d434d828056049e3baba6aab76e2b1bc   配置文件httpget_routerinfo_sha1key
https://1024tools.com/hmac
*/
func (ops *OPSServer) HandlePhoneBindNumInfo(w http.ResponseWriter, r *http.Request) {
	var areaName = [13]string{"meijiang", "meixian", "fengshun", "wuhua", "jiaoling", "pingyuan", "dapu", "xingning", "longchuan", "heping", "lianping", "nanxiong", "raoping"}
	var responseBindInfo responseBindInfo

	begindate := r.URL.Query().Get("begindate")
	enddate := r.URL.Query().Get("enddate")
	appkey := r.URL.Query().Get("appkey")
	timestamp := r.URL.Query().Get("ts")
	signature := r.URL.Query().Get("sign")

	if begindate == "" || enddate == "" || timestamp == "" || signature == "" || appkey == "" {
		responseBindInfo.Error = 1
		responseBindInfo.Msg = "请求参数有误"
		//logger.Error("请求参数有误")
	} else {
		//验证时间戳是否有效性
		cur_timestamp := time.Now().UnixNano() / int64(time.Millisecond)
		http_timestamp, err := strconv.Atoi(timestamp)
		if err == nil {
			dv_time := math.Abs(float64(cur_timestamp) - float64(http_timestamp))
			logger.Debug("cur_time:%d , http_time:%d , dv_time:%d", cur_timestamp, http_timestamp, int64(dv_time))
			if dv_time > 120000 {
				responseBindInfo.Error = 8
				responseBindInfo.Msg = "请求超时,请重试"
				//logger.Error("请求超时,请重试")
			} else {
				//logger.Debug("appkey: " + appkey)
				//logger.Debug("signature: " + signature)
				appKeyinfo, err := mongodb.Db_handler.GetOpsAppKey(*config.DbOpsAppKeyCollection, appkey)
				if err != nil {
					logger.Error("Find appkey err: ", err)
					responseBindInfo.Error = 2
					responseBindInfo.Msg = "非法请求"
				} else {
					//appkey+timestamp内容进行SHA1加密
					content := appkey + timestamp
					string_hmac_sha1 := gethmacsha1(appKeyinfo.Secret, content)
					chiReg := regexp.MustCompile("[`~!@#$%^&*()+=|{}:;\\[\\].<>/?~！@#￥%……&*（）——+|{}【】‘；：”“’。，、？']")
					handle_hmacsha1 := chiReg.ReplaceAllString(string_hmac_sha1, "")
					//判断加密是否一致，否则是非法侵入
					if strings.Compare(handle_hmacsha1, signature) == 0 {
						logger.Debug("查询当月绑定终端数量")
						//根据区域，查询数据库
						for i := 0; i < len(areaName); i++ {
							var item monthData
							var clientBindInfo clientBindInfo

							areaMonthDatas, err := mongodb.Db_handler.GetOpsAreaBindDatas(*config.DbOpsAreaBindDataCollection, areaName[i], begindate, enddate)
							if err != nil {
								logger.Error("Find bindinfo err: ", err)
							} else {
								clientBindInfo.AreaParttion = areaName[i]
								for j := 0; j < len(areaMonthDatas); j++ {
									item.Month = areaMonthDatas[j].Month
									item.ClientNum = areaMonthDatas[j].ClientNum
									item.ClientTotal = areaMonthDatas[j].ClientTotal
									clientBindInfo.MonthDatas = append(clientBindInfo.MonthDatas, item)
								}
							}
							responseBindInfo.ClientBindInfos = append(responseBindInfo.ClientBindInfos, clientBindInfo)
						}
					} else {
						responseBindInfo.Error = 2
						responseBindInfo.Msg = "非法请求"
					}
				}
			}
		}
	}

	response_json, err := json.Marshal(responseBindInfo)
	if err != nil {
		logger.Error("marshal fail: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response_json))
	}
}

// 删除手机号绑定终端接口
/*测试案例 
http://localhost:8070/del_phone_bind_ca?phone=18712341234&caMac=50012
*/
func (ops *OPSServer)HandleDelPhoneBindCa(w http.ResponseWriter, r *http.Request) {
	// 获取url参数
	caMac := r.URL.Query().Get("caMac")
	// acMac := r.URL.Query().Get("acMac")
	phone := r.URL.Query().Get("phone")
	// 定义返回结果处理，失败和成功
	var err error
	err_str := "err_no=1"
	defer func(){
		DisableBrowserCache(w)
		if err != nil{
			logger.Error("DeleteBind failed: ClientInfo(phone:%s clientMac:%s), %s",phone,caMac,err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err_str))
		}else{
			logger.Info("DeleteBind success: ClientInfo(phone:%s clientMac:%s)",phone,caMac)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}()
	// 验证用户权限(获取cookie，获取session，获取操作权限，判断，返回结果)
	if userAuth(w, r) == false{
		err_str = "err_no=2"
		err = fmt.Errorf("DeleteBind failed, operator not admin")
		return
	}
	// 验证今日删除次数（获取今日删除次数，判断时间，不是今日则为今日第一次，否则判断次数，小于加1，返回true）
	userName := getUserName(w, r).(string)
	if IsMaxDelCaByUser(userName) == false{
		err_str = "err_no=3"
		err = fmt.Errorf("DeleteBind failed, Today del is more than %d time",30)
		return
	}
	// 验证数据合法性，待做
	// 
	// 删除用户数据,将用户数据从原数据库删除，转移到另外一个数据库
	err = mongodb.Db_handler.RemoveClientInfoByMacAndPhone(*config.DbClientInfoCollection, caMac, phone)
	if err != nil{
		err_str = "err_no=4"
	}
}

// 删除手机号绑定路由接口
/*测试案例 
http://localhost:8070/del_phone_bind_ac?phone=18712341233&acmac=1
*/
func (ops *OPSServer) HandleDelPhoneBindAc(w http.ResponseWriter, r *http.Request) {
		// 获取url参数
		acMac := r.URL.Query().Get("acmac")
		// acMac := r.URL.Query().Get("acMac")
		phone := r.URL.Query().Get("phone")
		// 定义返回结果处理，失败和成功
		var err error
		err_str := "err_no=1"
		defer func(){
			DisableBrowserCache(w)
			if err != nil{
				logger.Error("DeleteBind failed: Router(phone:%s Mac:%s), %s",phone,acMac,err.Error())
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err_str))
			}else{
				logger.Info("DeleteBind success: Router(phone:%s Mac:%s)",phone,acMac)
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}
		}()
		// 验证用户权限(获取cookie，获取session，获取操作权限，判断，返回结果)
		if userAuth(w, r) == false{
			err_str = "err_no=2"
			err = fmt.Errorf("DeleteBind failed, operator not admin")
			return
		}
		// 验证今日删除次数（获取今日删除次数，判断时间，不是今日则为今日第一次，否则判断次数，小于加1，返回true）
		userName := getUserName(w, r).(string)
		if IsMaxDelAcByUser(userName) == false{
			err_str = "err_no=3"
			err = fmt.Errorf("DeleteBind failed, Today del is more than %d time",30)
			return
		}
		// 验证数据合法性，待做
		// 
		// 删除用户数据,将用户数据从原数据库删除，转移到另外一个数据库
		err = mongodb.Db_handler.RemoveRouterByMacAndPhone(*config.DbClientInfoCollection, acMac, phone)
		if err != nil{
			err_str = "err_no=4"
		}
}

// 删除根据手机号，删除手机号绑定的终端
/*测试案例 
http://localhost:8070/del_cabind_by_phone?phone=18712341233
*/
func (ops *OPSServer) HandleDelCaBindByPhone(w http.ResponseWriter, r *http.Request) {
	// 获取url参数
	phone := r.URL.Query().Get("phone")
	// 定义返回结果处理，失败和成功
	var err error
	err_str := "err_no=1"
	defer func(){
		DisableBrowserCache(w)
		if err != nil{
			logger.Error("DeleteBind failed: phone:%s, %s",phone,err.Error())
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err_str))
		}else{
			logger.Info("DeleteBind success: phone:%s",phone)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		}
	}()
	// 验证用户权限(获取cookie，获取session，获取操作权限，判断，返回结果)
	if userAuth(w, r) == false{
		err_str = "err_no=2"
		err = fmt.Errorf("DeleteBind failed, operator not admin")
		return
	}
	// 验证今日删除次数（获取今日删除次数，判断时间，不是今日则为今日第一次，否则判断次数，小于加1，返回true）
	userName := getUserName(w, r).(string)
	if IsMaxDelPhoneByUser(userName) == false{
		err_str = "err_no=3"
		err = fmt.Errorf("DeleteBind failed, Today del is more than %d time",30)
		return
	}
	// 验证数据合法性，待做
	// 
	// 删除用户数据,将用户数据从原数据库删除，转移到另外一个数据库
	err = mongodb.Db_handler.RemoveCaByPhone(*config.DbClientInfoCollection,phone)
	if err != nil{
		err_str = "err_no=4"
	}
}

// 获取登录的用户名字
func getUserName(w http.ResponseWriter, r *http.Request)(interface{}){
	var sessionID = sessionMgr.CheckCookieValid(w, r)
	name, _ := sessionMgr.GetSessionVal(sessionID, "session_username")
	return name
}
// 登录用户，管理员身份验证
func userAuth(w http.ResponseWriter, r *http.Request)(bool){
	operator := getUserName(w, r)
	if operator != "admin" {
		return false
	}
	return true
}



// 判断当前管理员用户，今日CA删除接口调用是否达到最大次数
func IsMaxDelCaByUser(userName string)(bool){
	// 根据用户名字和时间，查找用户删除次数数据库
	// 没有数据，则新建数据初始化，并插入数据到数据库
	dc, err := mongodb.Db_handler.GetDelBindConutByName(userName)
	// 找不到数据，报错，属于今日第一次调用删除手机号绑定接口
	if err != nil{
		dc.UserName = userName
		dc.Date = time.Now()
		dc.AcCounter = 0
		dc.CaCounter = 1
		err := mongodb.Db_handler.InsertDelBindConut(dc)
		if err != nil {
			logger.Error("del bind fail in InsertDelBindConut, %s", err.Error())
		}
		return true
	}else{
		// 如果找到数据，那么判断用户数据的删除次数，没有超过指定次数则返回成功
		if dc.CaCounter >= *config.OpsDelBindDayCount {
			logger.Error("this user(name:%s) can be del ,because del bind max %d times a day!", userName, *config.OpsDelBindDayCount)
			return false
		}
		oldDate := dc.Date
		dc.Date = time.Now()
		dc.CaCounter = dc.CaCounter + 1
		err := mongodb.Db_handler.UpdateDelBindConut(dc, oldDate)
		if err != nil {
			logger.Error("del bind fail in UpdateDelBindConut, %s", err.Error())
		}
		return true
	}
}
// 判断当前管理员用户，今日AC删除接口调用是否达到最大次数
func IsMaxDelAcByUser(userName string)(bool){
	// 根据用户名字和时间，查找用户删除次数数据库
	// 没有数据，则新建数据初始化，并插入数据到数据库
	dc, err := mongodb.Db_handler.GetDelBindConutByName(userName)
	// 找不到数据，报错，属于今日第一次调用删除手机号绑定接口
	if err != nil{
		dc.UserName = userName
		dc.Date = time.Now()
		dc.AcCounter = 1
		dc.CaCounter = 0
		err := mongodb.Db_handler.InsertDelBindConut(dc)
		if err != nil {
			logger.Error("del bind fail in InsertDelBindConut, %s", err.Error())
		}
		return true
	}else{
		// 如果找到数据，那么判断用户数据的删除次数，没有超过指定次数则返回成功
		if dc.AcCounter >= *config.OpsDelBindDayCount {
			logger.Error("this user(name:%s) can be del ,because del bind max %d times a day!", userName, *config.OpsDelBindDayCount)
			return false
		}
		oldDate := dc.Date
		dc.Date = time.Now()
		dc.AcCounter = dc.AcCounter + 1
		err := mongodb.Db_handler.UpdateDelBindConut(dc, oldDate)
		if err != nil {
			logger.Error("del bind fail in UpdateDelBindConut, %s", err.Error())
		}
		return true
	}
}
//判断当前管理员用户，今日Phone删除接口调用是否达到最大次数
func IsMaxDelPhoneByUser(userName string)(bool){
	// 根据用户名字和时间，查找用户删除次数数据库
	// 没有数据，则新建数据初始化，并插入数据到数据库
	dc, err := mongodb.Db_handler.GetDelBindConutByName(userName)
	// 找不到数据，报错，属于今日第一次调用删除手机号绑定接口
	if err != nil{
		dc.UserName = userName
		dc.Date = time.Now()
		dc.AcCounter = 0
		dc.CaCounter = 0
		dc.PhoneCounter = 1
		err := mongodb.Db_handler.InsertDelBindConut(dc)
		if err != nil {
			logger.Error("del bind fail in InsertDelBindConut, %s", err.Error())
		}
		return true
	}else{
		// 如果找到数据，那么判断用户数据的删除次数，没有超过指定次数则返回成功
		if dc.AcCounter >= *config.OpsDelBindDayCount {
			logger.Error("this user(name:%s) can be del ,because del bind max %d times a day!", userName, *config.OpsDelBindDayCount)
			return false
		}
		oldDate := dc.Date
		dc.Date = time.Now()
		dc.PhoneCounter = dc.PhoneCounter + 1
		err := mongodb.Db_handler.UpdateDelBindConut(dc, oldDate)
		if err != nil {
			logger.Error("del bind fail in UpdateDelBindConut, %s", err.Error())
		}
		return true
	}
}
