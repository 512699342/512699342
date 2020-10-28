package config

import (
	//"component"
	"flag"

	toml "github.com/extrame/go-toml-config"
	"github.com/wonderivan/logger"

	//"net"
	"path/filepath"
	//"strings"
)

var (
	NasIp = toml.String("basic.nas_ip", "192.168.2.3")

	//日志文件
	LogFile = toml.String("basic.logfile", "")
	//日志配置
	LogConfig = toml.String("basic.logconfig", "./log.json")
	//运维配置(true：运维相关的代码，false:认证相关的代码)
	//如果是运维，则需配置http.ops_port  cmcc.port
	OpsEnable = toml.Bool("basic.ops_enable", false)

	//Router绑定方式端口是否开启(true:开启,false:关闭)
	HttpNameAuthEnable = toml.Bool("http.name_auth_port_enable", true)
	//Router绑定方式端口
	HttpNameAuthPort = toml.Int("http.name_auth_port", 8080)
	//手机号绑定方式是否开启(true:开启,false:关闭)
	HttpPhoneAuthEnable = toml.Bool("http.phone_auth_port_enable", true)
	//手机号绑定方式端口
	HttpPhoneAuthPort = toml.Int("http.phone_auth_port", 8090)
	//运维http端口
	HttpOpsPort = toml.Int("http.ops_port", 8070)
	//portal用户下线通知url
	CallBackUrl = toml.String("http.callback_logout", "")
	//portal认证成功页面路径
	LoginSuccessPage = toml.String("http.login_success_page", "./web/welcome_success.html")
	//portal登录失败页面
	LoginFailPage = toml.String("http.login_fail_page", "./web/welcome_fail.html")
	//portal注册Router绑定方式页面
	RegisterNamePage = toml.String("http.register_name_page", "./web/register_name.html")
	//portal注册手机号绑定方式页面
	RegisterPhonePage = toml.String("http.register_phone_page", "./web/register_phone.html")
	//portal注册成功页面
	RegisterSucessPage = toml.String("http.register_page_success", "./web/register_success.html")
	//ops 首页
	OpsIndexPage = toml.String("http.ops_index_page", "ops/www/web/index.html")
	//ops 登录页面
	OpsLoginPage = toml.String("http.ops_login_page", "ops/www/web/login.html")
	//ops 用户列表页面
	OpsUserListPage = toml.String("http.ops_user_list_page", "ops/www/web/user_list.html")
	//ops 修改用户信息页面
	OpsAlterUserPage = toml.String("http.ops_alter_user_page", "ops/www/web/alter_user.html")
	//ops 添加用户页面
	OpsAddUserPage = toml.String("http.ops_add_user_page", "ops/www/web/add_user.html")

	//认证服务器端口
	CmccPort = toml.Int("cmcc.port", 50100)
	//版本号
	CmccVersion = toml.Int("cmcc.version", 1)
	//cmcc密码
	CmccSecret = toml.String("cmcc.secret", "123456")
	//nas端口
	CmccNasPort     = toml.Int("cmcc.nas_port", 2000)
	CmccPortalTimes = toml.Int("cmcc.portal_times", 3)
	//radius计费服务开关(true：开启   false：关闭)
	RadiusEnable = toml.Bool("radius.enabled", true)
	//认证端口
	RadiusAuthPort = toml.Int("radius.port", 1812)
	//计费端口
	RadiusAccPort = toml.Int("radius.acc_port", 1813)
	//radius密码
	RadiusSecret = toml.String("radius.secret", "123456")

	NufrontDomain = toml.String("nufront.domain", "nufront.com")

	//数据库访问地址
	DbConnection = toml.String("database.connection", "rw_radius:inqw13912301z@192.168.254.10:27017,192.168.254.11:27017,192.168.254.12:27017/radius")
	//数据库名字
	DbName = toml.String("database.name", "radius")
	//对应关系集合
	DbRelationCollection = toml.String("database.relation_collection", "relation")
	//Router实名绑定-注册成功信息集合
	DbRouterCollection = toml.String("database.router_collection", "router")
	//用户上下线集合
	DbClientStatusCollection = toml.String("database.client_status_collection", "clientstatus")
	//手机验证码集合
	DbPhoneValidateCodeCollection = toml.String("database.phone_validatecode_collection", "phonevalidatecode")
	//Router实名绑定-每天限制实名注册集合
	DbPhoneRealNameAuthCollection = toml.String("database.phone_realname_auth_collection", "phonerealname")
	//手机号绑定-手机号注册
	DbClientInfoCollection = toml.String("database.client_info_collection", "chs_clientinfo")
	//手机验证码有效期开启(true:开启,false:关闭)
	DbPhoneValidateCodeGCEnable = toml.Bool("database.phone_validatecode_gc_enable", true)
	//rsa数据库加密公钥路径
	EncryptRSAPublicKey = toml.String("encrypt.rsa_public_key", "./rsa_public_key.pem")
	//rsa数据库加密私钥路径
	EncryptRSAPrivateKey = toml.String("encrypt.rsa_private_key", "./rsa_private_key.pem")

	//ops用户信息集合
	DbOpsUserInfoCollection = toml.String("database.ops_user_info_collection", "ops_userinfo")
	//ops admin管理员初始密码
	DbOpsAdminPassword = toml.String("database.ops_admin_initial_password", "admin")

	//统计数据,与网管组接口定义集合
	DbOpsAreaBindDataCollection = toml.String("database.ops_area_binddata_collection", "area_binddata")
	//统计数据,与网管组接口定义集合,appkey
	DbOpsAppKeyCollection = toml.String("database.ops_app_key_collection", "appkey")

	//ACIP与服务热线对应关系集合
	ServicePhoneCollection = toml.String("database.service_phone_collection", "service_phone")

	//euht
	//姓名和手机号码加密(false : 不加密, true ：加密)
	EuhtDataEncryptStategy = toml.Bool("euht.data_encrypt_strategy", false)
	//姓名隐藏(false ：不隐藏, true ： 隐藏)
	EuhtHideNameStategy = toml.Bool("euht.hide_name_strategy", false)
	//手机号码隐藏(false ：不隐藏, true ： 隐藏)
	EuhtHidePhoneStategy = toml.Bool("euht.hide_phone_strategy", false)
	//是否开启接受对应关系(true:开启, false:关闭)
	EuhtRelationEnable = toml.Bool("euht.relation_enable", true)
	//对应关系UDP端口-接收村级服务器上下线数据
	EuhtRelationPort = toml.Int("euht.relation_port", 7998)

	//云之讯短信接口信息
	//云之讯-短信请求url
	EuhtSmsYzxUrl = toml.String("euht.sms_yzx_url", "https://open.ucpaas.com/ol/sms/sendsms")
	//云之讯-账号相关信息-应用ID
	EuhtSmsYzxAppid = toml.String("euht.sms_yzx_appid", "93649dcfefc24dc29afd8f6b2b0a1728")
	//云之讯-账号相关信息-用户sid
	EuhtSmsYzxAccountSid = toml.String("euht.sms_yzx_account_Sid", "eb6f0a1b4f3e59ce23903ffb8b2a37c3")
	//云之讯-账号相关信息-密钥
	EuhtSmsYzxAuthToken = toml.String("euht.sms_yzx_auth_token", "0278286f8ebecc70d2c0f891cf521892")
	//云之讯-账号相关信息-短信模板
	EuhtSmsYzxTemplateid = toml.String("euht.sms_yzx_templateid", "470311")

	//聚合数据接口信息
	//聚合数据-短信平台-请求地址
	EuhtSmsJuheUrl = toml.String("euht.sms_juhe_url", "hhttp://v.juhe.cn/sms/send")
	//聚合数据-短信平台-短信模板
	EuhtSmsJuheTemplateid = toml.String("euht.sms_juhe_templateid", "170690")
	//聚合数据-短信平台-密钥
	EuhtSmsJuheAuthToken = toml.String("euht.sms_juhe_auth_token", "403a8fa94b1f8a0a7e71f0f2a13fa410")
	//聚合数据-实名验证-请求地址
	EuhtRealnameJuheUrl = toml.String("euht.realname_juhe_url", "http://v.juhe.cn/telecom2/query")
	//聚合数据-实名验证-密钥
	EuhtRealnameJuheApikey = toml.String("euht.realname_juhe_apikey", "6bc90a99ef8a38b224d701402bc89aaa")

	//接口之家-实名验证-请求地址
	EuhtRealnameUrl = toml.String("euht.realname_url", "https://v.apihome.org/phonetwo")
	//接口之家-实名验证-密钥
	EuhtRealnameApikey = toml.String("euht.realname_apikey", "ad053bb4dc5b498c9868b22e1f603207")

	//短信验证码个数
	EuhtSmsValidatecodeLen = toml.Int("euht.sms_validatecode_len", 4)
	//0：云之讯短信    1: 聚合短信
	EuhtSmsServiceChoice = toml.Int("euht.sms_service_choice", 0)
	//0：接口之家实名  1：聚合实名
	EuhtRealnameServiceChoice = toml.Int("euht.realname_service_choice", 0)

	//Router绑定方式-一个手机号码只能绑定N个Router
	EuhtRealnameRouterNumber = toml.Int("euht.realname_router_number", 3)
	//Router绑定方式-一个手机号码每天只能验证5次，超过N次只能隔天验证
	EuhtRealnameAuthDayCount = toml.Int("euht.realname_auth_daycount", 5)

	//手机绑定方式-一个手机号码只能绑定N个设备
	EuhtRegisterDeviceMaxNumber = toml.Int("euht.register_device_max_number", 3)

	//与网管组统计数据接口hmac sha1 key值，暂时不用
	EuhtHttpgetBindInfoSha1key = toml.String("euht.httpget_bindinfo_sha1key", "d434d828056049e3baba6aab76e2b1bc")

	//服务总线号码
	EuhtDefaultServicePhone = toml.String("euht.default_service_phone", "19926114977")
)

func init() {
	//配置文件：radius.conf
	//当不存在的时候，会读取config里面的默认配置
	path := flag.String("config", "./server.conf", "设置配置文件的路径")
	flag.Parse()
	logger.Info("use config file: ", *path)
	*path = filepath.FromSlash(*path)
	if err := toml.Parse(*path); err != nil {
		logger.Error("config file not find, use default config")
	}

	if *OpsEnable == true {
		*DbPhoneValidateCodeGCEnable = false
	}
}

func IsValid() bool {
	return true
}
