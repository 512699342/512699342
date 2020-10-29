package euht

import (
	"component/mongodb"
	"config"
	"github.com/wonderivan/logger"
	"sync/atomic"
	"time"
	"utility"
	//"time"
)

const DB_ROUTER_TMPCOLLECT = "router_tmp"

const (
	UpdateRSAKeyInitStatus  = iota
	UpdateRSAKeyStart       = iota
	SetRSAPubKeyFail        = iota
	SetRSAPriKeyFail        = iota
	IterRouterFail          = iota
	DropRouterCollectFail   = iota
	RenameRouterCollectFail = iota
	UpdateRSAKeySuccess     = iota
)

var UpdateRSAKeyStatus int = UpdateRSAKeyInitStatus

//统计对应关系查找失败计数器
var queryRelationAllCount uint64 = 1
var queryRelationFailCount uint64 = 1

//客户端portal认证过上下线记录，写入数据库
func RecordClientOnlineStatus(relation mongodb.Relation, onlineStatus bool) {
	status := mongodb.ClientStatus{
		Relation:     relation,
		Date:         time.Now(),
		OnlineStatus: onlineStatus,
	}
	err := mongodb.Db_handler.InsertClientStatus(status)
	if err != nil {
		logger.Error("ac(ip:%s) client(ip:%s mac:%s) router(sn:%s mac:%s) : record status(%t) fail: ", relation.AcIp, relation.ClientIp, relation.ClientMac, relation.RouterSn, relation.RouterMac, onlineStatus, err.Error())
	} else {
		logger.Debug("ac(ip:%s) client(ip:%s mac:%s) router(sn:%s mac:%s) : record status(%t) success", relation.AcIp, relation.ClientIp, relation.ClientMac, relation.RouterSn, relation.RouterMac, onlineStatus)
	}
}

//Router绑定方式，注册成功用户信息写入数据库
func RegisterUserInfo(relation mongodb.Relation, name string, phone string) bool {
	hideName := utility.HideName(name)
	hidePhone := utility.HidePhone(phone)
	logger.Debug("RegisterUserInfo: router(sn:%s mac:%s) phone:%s name:%s ", relation.RouterSn, relation.RouterMac, hidePhone, hideName)

	router := mongodb.Router{
		Account: mongodb.Account{
			Name:  hideName,
			Phone: hidePhone,
		},
		Mac:      relation.RouterMac,
		Sn:       relation.RouterSn,
		AcIp:     relation.AcIp,
		BindTime: time.Now(),
	}

	encryptName, err := utility.RSAEncrypt([]byte(name))
	//encryptName, err := utility.RSAEncryptAndBase64([]byte(name))
	if err != nil {
		logger.Error("RegisterUserInfo: router(sn:%s mac:%s) phone:%s name:%s : name encrypt fail", relation.RouterSn, relation.RouterMac, hidePhone, hideName, err.Error())
		return false
	}
	encryptPhone, err := utility.RSAEncrypt([]byte(phone))
	//encryptPhone, err := utility.RSAEncryptAndBase64([]byte(phone))
	if err != nil {
		logger.Error("RegisterUserInfo: router(sn:%s mac:%s) phone:%s name:%s : phone encrypt fail", relation.RouterSn, relation.RouterMac, hidePhone, hideName, err.Error())
		return false
	}
	router.Account.EncryptName = encryptName
	router.Account.EncryptPhone = encryptPhone

	err = mongodb.Db_handler.UpsertRouter(router)
	if err != nil {
		logger.Error("RegisterUserInfo : upsert router(sn:%s mac:%s)  fail: %s", relation.RouterSn, relation.RouterMac, err.Error())
		return false
	} else {
		logger.Debug("RegisterUserInfo : upsert router(sn:%s mac:%s) success", relation.RouterSn, relation.RouterMac)
		return true
	}
}

//手机号绑定方式，注册成功用户信息写入数据库
func RegisterClientInfo(acip string, usermac string, phone string) bool {

	logger.Debug("RegisterClientInfo: phone:%s usermac:%s ", phone, usermac)

	clientInfo := mongodb.ClientInfo{
		Phone:     phone,
		ClientMac: usermac,
		AcIp:      acip,
		BindTime:  time.Now(),
	}

	err := mongodb.Db_handler.UpsertClientinfo(clientInfo)
	if err != nil {
		logger.Error("RegisterClientInfo : upsert clientinfo(mac:%s phone:%s) fail: %s", usermac, phone, err.Error())
		return false
	} else {
		logger.Debug("RegisterClientInfo : upsert clientinfo(mac:%s phone:%s) success", usermac, phone)
		return true
	}
}

//判断连接Router是否有实名验证过
func IsRealName(routerMac string) bool {
	router, err := mongodb.Db_handler.GetRouterByMac(routerMac)
	if err != nil {
		logger.Error("realname : find router(mac:%s) fail: %s", routerMac, err.Error())
		return false
	}
	//
	if router.Account.Phone == "" {
		logger.Error("realname : router(mac:%s) have no accout: %s", routerMac, err.Error())
		return false
	}
	return true
}

//判断设备MAC是否有认证过
func IsClientRegistered(usermac string) bool {
	clientinfo, err := mongodb.Db_handler.GetClientinfoByMac(usermac)
	if err != nil {
		logger.Error("IsClientRegistered : find client(mac:%s) fail: %s", usermac, err.Error())
		return false
	}
	//
	if clientinfo.Phone == "" {
		logger.Error("IsClientRegistered : router(mac:%s) have no accout: %s", usermac, err.Error())
		return false
	}
	return true
}

//根据设备MAC，查询对应关系信息
func GetRelationByClientMac(usermac string) (mongodb.Relation, error) {
	atomic.AddUint64(&queryRelationAllCount, 1)
	relation, err := mongodb.Db_handler.GetRelationByClientMac(usermac)
	if err != nil {
		//统计对应关系无法找到
		failcount := atomic.AddUint64(&queryRelationFailCount, 1)
		allcount := atomic.LoadUint64(&queryRelationAllCount)
		logger.Error("find relation for client(mac:%s) fail: %s", usermac, err.Error())
		logger.Info("get relation  statistics: failcount: %d  allcount: %d  fail per: %d%s ", failcount, allcount, failcount*100/allcount, "%")

	}
	return relation, err
}

//根据设备IP，查询对应关系信息
func GetRelationByClientIP(userip string) (mongodb.Relation, error) {
	//atomic.AddUint64(&queryRelationAllCount, 1)
	relation, err := mongodb.Db_handler.GetRelationByClientIp(userip)
	if err != nil {
		//统计对应关系无法找到
		//failcount := atomic.AddUint64(&queryRelationFailCount, 1)
		//allcount := atomic.LoadUint64(&queryRelationAllCount)
		//logger.Error("get relation for client(ip:%s) fail: %s", userip, err.Error())
		//logger.Info("get relation  statistics: failcount: %d  allcount: %d  fail per: %d%s ", failcount, allcount, failcount*100/allcount, "%")

	}
	return relation, err
}

//检查新密钥,如正确并新建立集合
func updateRouterUseNewRSAKey(router mongodb.Router) error {

	name, err := utility.RSADecrypt(router.Account.EncryptName)
	if err != nil {
		return err
	}
	encryptName, err := utility.RSAEncrypt(name)
	if err != nil {
		return err
	}
	router.Account.EncryptName = encryptName

	phone, err := utility.RSADecrypt(router.Account.EncryptPhone)
	if err != nil {
		return err
	}
	encryptPhone, err := utility.RSAEncrypt(phone)
	if err != nil {
		return err
	}
	router.Account.EncryptPhone = encryptPhone
	return mongodb.Db_handler.InsertRouterInCollection(router, DB_ROUTER_TMPCOLLECT)

}

//根据最新RSA密钥更新相应数据库
func UpdateRSAKey(publicKey string, privateKey string) {
	//logger.Debug(publicKey)
	//logger.Debug(privateKey)
	UpdateRSAKeyStatus = UpdateRSAKeyStart
	err := utility.SetNewPublicKey([]byte(publicKey))
	if err != nil {
		UpdateRSAKeyStatus = SetRSAPubKeyFail
		logger.Error("SetNewPublicKey fail: %s", err.Error())
		return
	} else {
		logger.Info("SetNewPublicKey ok")
	}
	err = utility.SetNewPrivateKey([]byte(privateKey))
	if err != nil {
		UpdateRSAKeyStatus = SetRSAPriKeyFail
		logger.Error("SetNewPrivateKey fail: %s", err.Error())
		return
	} else {
		logger.Info("SetNewPrivateKey ok")
	}

	err = mongodb.Db_handler.DropCollection(*config.DbName + "." + DB_ROUTER_TMPCOLLECT)

	err = mongodb.Db_handler.IterRouter(updateRouterUseNewRSAKey)
	if err != nil {
		UpdateRSAKeyStatus = IterRouterFail
		logger.Error("IterRouter fail: %s", err.Error())
		return
	} else {
		logger.Info("IterRouter ok")
	}

	err = mongodb.Db_handler.DropCollection(*config.DbRouterCollection)
	if err != nil {
		UpdateRSAKeyStatus = DropRouterCollectFail
		logger.Error("DropCollection fail: %s", err.Error())
		return
	} else {
		logger.Info("DropCollection ok")
	}

	err = mongodb.Db_handler.RenameCollection(*config.DbName+"."+DB_ROUTER_TMPCOLLECT, *config.DbName+"."+*config.DbRouterCollection, true)
	if err != nil {
		UpdateRSAKeyStatus = RenameRouterCollectFail
		logger.Error("RenameCollection fail: %s", err.Error())
		return
	} else {
		logger.Info("RenameCollection ok")
	}
	UpdateRSAKeyStatus = UpdateRSAKeySuccess

	//to do

	// modify rsa key file

	// open register server

}

//使用手机号码，查询绑定个数,如已绑定超过3个，返回true
func IsMaxRegisterDevicesByPhone(phone string) bool {

	devices_num, err := mongodb.Db_handler.GetDevicesNumByPhoneAndTime(phone)

	if err != nil {
		logger.Error("GetDevicesNumByPhone err, phone: %s, err: %s", phone, err.Error())
		//如果获取异常，则返回错误
		return false
	} else {
		//logger.Debug("GetDevicesNumByPhone, phone: %s device number[%d]", phone, devices_num)
		if devices_num >= *config.EuhtRegisterDeviceMaxNumber {
			return true
		} else {
			return false
		}
	}
}
