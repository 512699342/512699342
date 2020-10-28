package mongodb

import (
	"config"

	"time"

	"github.com/gopkg.in/mgo.v2"
	"github.com/gopkg.in/mgo.v2/bson"
	"github.com/wonderivan/logger"
)

//
type Account struct {
	Name         string `bson:"name"`
	Phone        string `bson:"phone"`
	EncryptName  []byte `bson:"encryptname"`
	EncryptPhone []byte `bson:"encryptphone"`
	//EncryptName  string `bson:"encryptname"`
	//EncryptPhone string `bson:"encryptphone"`
}

// euht router
type Router struct {
	AcIp string `bson:"acIp"`
	Mac  string `bson:"mac"`
	Sn   string `bson:"sn"`
	//router状态
	Status bool `bson:"status"`
	// 用户信息
	Account Account `bson:"account"`
	//用户信息绑定时间
	BindTime time.Time `bson:"bindtime"`
}

//euht phone client
type ClientInfo struct {
	AcIp      string    `bson:"acIp"`
	Phone     string    `bson:"phone"`
	ClientMac string    `bson:"clientMac"`
	BindTime  time.Time `bson:"bindtime"`
}

//router 和下面的手机对应关系
type Relation struct {
	AcIp      string    `bson:"acIp"`
	RouterMac string    `bson:"routerMac"`
	RouterIp  string    `bson:"routerIp"`
	RouterSn  string    `bson:"routerSn"`
	ClientMac string    `bson:"clientMac"`
	ClientIp  string    `bson:"clientIp"`
	Timestamp time.Time `bson:"timestamp"`
}

type ClientStatus struct {
	Relation     Relation
	Date         time.Time `bson:"date"`
	OnlineStatus bool      `bson:"onlinestatus"`
}

//手机验证码
type PhoneValidateCode struct {
	Phone        string    `bson:"phone"`
	ValidateCode string    `bson:"validatecode"`
	Time         time.Time `bson:"time"`
}

//手机实名认证次数
type RealNameAuth struct {
	AcIp     string    `bson:"acIp"`
	RouterSn string    `bson:"routerSn"`
	Phone    string    `bson:"phone"`
	Name     string    `bson:"name"`
	Date     time.Time `bson:"date"`
	Counter  int       `bson:"counter"`
}

//ops user info
type OpsUserInfo struct {
	UserName     string `bson:"userName"`
	UserPassword string `bson:"userPassword"`
	FullName     string `bson:"fullName"`
	UserPhone    string `bson:"userPhone"`
	UserEmail    string `bson:"userEmail"`
}

//ops bindnum info
type AreaMonthData struct {
	AreaParttion string `bson:"area"`
	Month        string `bson:"month"`
	ClientNum    int    `bson:"clientNum"`
	ClientTotal  int    `bson:"clientTotal"`
}

//ops appkey info
type AppKeyinfo struct {
	Name          string `bson:"name"`
	Secret        string `bson:"secret"`
	AppKey        string `bson:"appkey"`
	AreaPartition string `bson:"areapartition"`
}

//service phone info
type ServicePhoneInfo struct {
	AcIp         string `bson:"acIp"`
	ServicePhone string `bson:"servicePhone"`
}

type MongodbHandler struct {
	*mgo.Session
}

var Db_handler *MongodbHandler

const CODE_GC_TIME = 5
const BIND_SP_TIME = time.Hour * 24 * 365 / 2

func init() {
	//连接mongodb
	var err error
	Db_handler, err = NewMongodbHandler(*config.DbConnection)
	if err != nil {
		logger.Error("connect to mongodb error")
	} else {
		if *config.DbPhoneValidateCodeGCEnable == true {
			go phoneValidateCodeGC()
		}
	}
}

func NewMongodbHandler(connection string) (*MongodbHandler, error) {
	s, err := mgo.Dial(connection)
	return &MongodbHandler{
		Session: s,
	}, err
}

//验证码超时回收
func phoneValidateCodeGC() {
	logger.Info("start phoneValidateCodeGC")
	for true {
		now := time.Now()
		s := Db_handler.getFreshSession()
		defer s.Close()
		query := bson.M{"time": bson.M{"$lt": now.Add(-CODE_GC_TIME * time.Minute)}}
		_, err := s.DB(*config.DbName).C(*config.DbPhoneValidateCodeCollection).RemoveAll(query)
		if err != nil {
			logger.Debug(err.Error())
		}
		time.Sleep(CODE_GC_TIME * time.Minute)
	}
}

//数据库基本操作函数

func (handler *MongodbHandler) getFreshSession() *mgo.Session {
	return handler.Session.Copy()
}

func (handler *MongodbHandler) findOne(collect string, query interface{}, result interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Find(query).One(result)
}

func (handler *MongodbHandler) findAll(collect string, query interface{}, result interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Find(query).All(result)
}

func (handler *MongodbHandler) findCount(collect string, query interface{}) (int, error) {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Find(query).Count()
}

func (handler *MongodbHandler) update(collect string, selector interface{}, update interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Update(selector, update)
}

// if not exist ,insert, else update
func (handler *MongodbHandler) upsert(collect string, selector interface{}, update interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	_, err := s.DB(*config.DbName).C(collect).Upsert(selector, update)
	return err
}
func (handler *MongodbHandler) insert(collect string, docs ...interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Insert(docs...)
}

func (handler *MongodbHandler) sort(collect string, query interface{}, sort string, result interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Find(query).Sort(sort).All(result)
}

func (handler *MongodbHandler) remove(collect string, selector interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Remove(selector)
}

func (handler *MongodbHandler) removeAll(collect string, selector interface{}) (*mgo.ChangeInfo, error) {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).RemoveAll(selector)
}

func (handler *MongodbHandler) DropCollection(collect string) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).DropCollection()
}

func (handler *MongodbHandler) runCommand(cmd interface{}, result interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).Run(cmd, result)
}

func (handler *MongodbHandler) runAdminCommand(cmd interface{}, result interface{}) error {
	s := handler.getFreshSession()
	defer s.Close()
	err := s.DB("admin").Run(cmd, result)
	return err
}

func (handler *MongodbHandler) RenameCollection(oldName string, newName string, dropTarget bool) error {
	cmd := bson.M{"renameCollection": oldName, "to": newName, "dropTarget": dropTarget}
	var result interface{}
	return handler.runAdminCommand(cmd, result)
}

func (handler *MongodbHandler) count(collect string) (int, error) {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(collect).Count()
}

func (handler *MongodbHandler) GetCollectionCount(collect string) (int, error) {
	return handler.count(collect)
}

//clientInfo phonenum clientmac
func (handler *MongodbHandler) GetClientinfoByMac(mac string) (ClientInfo, error) {
	query := bson.M{"clientMac": mac}
	clientInfo := ClientInfo{}
	err := handler.findOne(*config.DbClientInfoCollection, query, &clientInfo)
	return clientInfo, err
}

func (handler *MongodbHandler) GetClientinfoByAcIP(collect string, acIp string) (ClientInfo, error) {
	query := bson.M{"acIp": acIp}
	clientInfo := ClientInfo{}
	err := handler.findOne(collect, query, &clientInfo)
	return clientInfo, err
}

func (handler *MongodbHandler) UpdateClientinfo(mac string, clientInfo ClientInfo) error {
	selector := bson.M{"clientMac": mac}
	return handler.update(*config.DbClientInfoCollection, selector, clientInfo)
}

func (handler *MongodbHandler) UpsertClientinfo(clientInfo ClientInfo) error {
	selector := bson.M{"clientMac": clientInfo.ClientMac}
	return handler.upsert(*config.DbClientInfoCollection, selector, clientInfo)
}

func (handler *MongodbHandler) InsertClientinfo(clientInfo ClientInfo) error {
	return handler.insert(*config.DbClientInfoCollection, clientInfo)
}

func (handler *MongodbHandler) GetClientinfoByPhoneAll(collect string, phone string) ([]ClientInfo, error) {
	query := bson.M{"phone": phone}
	var clientInfo []ClientInfo
	err := handler.findAll(collect, query, &clientInfo)
	return clientInfo, err
}

func (handler *MongodbHandler) GetDevicesNumByPhone(phone string) (int, error) {
	query := bson.M{"phone": phone}
	Count, err := handler.findCount(*config.DbClientInfoCollection, query)
	return Count, err
}

func (handler *MongodbHandler) GetDevicesNumByPhoneByTime(phone string) (int, error) {
	now := time.Now()
	query := bson.M{"$and":[]bson.M{bson.M{"phone": phone},bson.M{"bindtime":bson.M{"$gt": now.Add(-BIND_SP_TIME)}}}}
	Count, err := handler.findCount(*config.DbClientInfoCollection, query)
	return Count, err
}

//relation

func (handler *MongodbHandler) GetRelationByClientMac(clientMac string) (Relation, error) {
	query := bson.M{"clientMac": clientMac}
	relation := Relation{}
	err := handler.findOne(*config.DbRelationCollection, query, &relation)
	return relation, err
}

func (handler *MongodbHandler) GetRelationByClientIp(clientIp string) (Relation, error) {
	query := bson.M{"clientIp": clientIp}
	relation := Relation{}
	err := handler.findOne(*config.DbRelationCollection, query, &relation)
	return relation, err
}

func (handler *MongodbHandler) GetCurRelationByClientIp(acIp string, clientIp string) ([]Relation, error) {
	query := bson.M{"clientIp": clientIp, "acIp": acIp}
	relations := []Relation{}
	err := handler.sort(*config.DbRelationCollection, query, "-timestamp", &relations)
	return relations, err
}

func (handler *MongodbHandler) RemoveRelationByUserInfo(clientMac string) error {
	s := handler.getFreshSession()
	defer s.Close()
	query := bson.M{"clientMac": clientMac}
	return s.DB(*config.DbName).C(*config.DbRelationCollection).Remove(query)
}

func (handler *MongodbHandler) UpdateRelation(clientMac string, relation Relation) error {
	selector := bson.M{"clientMac": clientMac}
	return handler.update(*config.DbRelationCollection, selector, relation)
}

func (handler *MongodbHandler) InsertRelation(relation Relation) error {
	return handler.insert(*config.DbRelationCollection, relation)
}

func (handler *MongodbHandler) UpsertRelation(relation Relation) error {
	selector := bson.M{"clientMac": relation.ClientMac}
	return handler.upsert(*config.DbRelationCollection, selector, relation)
}

// router

func (handler *MongodbHandler) IterRouter(fn func(Router) error) error {
	s := handler.getFreshSession()
	defer s.Close()

	var router Router
	iter := s.DB(*config.DbName).C(*config.DbRouterCollection).Find(nil).Iter()
	var err error
	for iter.Next(&router) {
		err = fn(router)
		if err != nil {
			return err
		}
	}
	return iter.Close()

}

func (handler *MongodbHandler) GetRouterByMac(mac string) (Router, error) {
	query := bson.M{"mac": mac}
	router := Router{}
	err := handler.findOne(*config.DbRouterCollection, query, &router)
	return router, err
}

func (handler *MongodbHandler) GetRoutersByPhone(phone string) ([]Router, error) {
	query := bson.M{"account.phone": phone}
	routers := []Router{}
	err := handler.findAll(*config.DbRouterCollection, query, &routers)
	return routers, err
}

func (handler *MongodbHandler) UpdateRouter(mac string, router Router) error {
	selector := bson.M{"mac": mac}
	return handler.update(*config.DbRouterCollection, selector, router)
}

func (handler *MongodbHandler) InsertRouter(router Router) error {
	return handler.insert(*config.DbRouterCollection, router)
}

func (handler *MongodbHandler) InsertRouterInCollection(router Router, collect string) error {
	return handler.insert(collect, router)
}

// if not exist ,insert, else update
func (handler *MongodbHandler) UpsertRouter(router Router) error {
	selector := bson.M{"mac": router.Mac}
	return handler.upsert(*config.DbRouterCollection, selector, router)
}

// RealName
func (handler *MongodbHandler) GetRealNameAuthByPhone(phone string) (RealNameAuth, error) {
	query := bson.M{"phone": phone}
	rn := RealNameAuth{}
	err := handler.findOne(*config.DbPhoneRealNameAuthCollection, query, &rn)
	return rn, err
}

func (handler *MongodbHandler) UpdateRealNameAuth(rn RealNameAuth) error {
	selector := bson.M{"phone": rn.Phone}
	return handler.update(*config.DbPhoneRealNameAuthCollection, selector, rn)
}

func (handler *MongodbHandler) InsertRealNameAuth(rn RealNameAuth) error {
	return handler.insert(*config.DbPhoneRealNameAuthCollection, rn)
}

// client status : online or offline

func (handler *MongodbHandler) InsertClientStatus(status ClientStatus) error {
	return handler.insert(*config.DbClientStatusCollection, status)
}

// PhoneValidateCode
func (handler *MongodbHandler) GetPhoneValidateCode(phone string) (PhoneValidateCode, error) {
	query := bson.M{"phone": phone}
	pvc := PhoneValidateCode{}
	err := handler.findOne(*config.DbPhoneValidateCodeCollection, query, &pvc)
	return pvc, err
}

// if not exist ,insert, else update
func (handler *MongodbHandler) UpsertPhoneValidateCode(pvc PhoneValidateCode) error {
	selector := bson.M{"phone": pvc.Phone}
	return handler.upsert(*config.DbPhoneValidateCodeCollection, selector, pvc)
}

func (handler *MongodbHandler) InsertPhoneValidateCode(pvc PhoneValidateCode) error {
	return handler.insert(*config.DbPhoneValidateCodeCollection, pvc)
}

func (handler *MongodbHandler) RemovePhoneValidateCode(phone string) error {
	query := bson.M{"phone": phone}
	return handler.remove(*config.DbPhoneValidateCodeCollection, query)
}

//Ops user info
func (handler *MongodbHandler) GetOpsUserInfoByName(collect string, name string) (OpsUserInfo, error) {
	query := bson.M{"userName": name}
	var userInfo OpsUserInfo
	err := handler.findOne(collect, query, &userInfo)
	return userInfo, err
}

func (handler *MongodbHandler) UpsertOpsUserInfo(collect string, userInfo OpsUserInfo) error {
	selector := bson.M{"userName": userInfo.UserName}
	return handler.upsert(collect, selector, userInfo)
}

func (handler *MongodbHandler) GetOpsUserInfos(collect string) ([]OpsUserInfo, error) {
	query := bson.M{}
	var userInfo []OpsUserInfo
	err := handler.findAll(collect, query, &userInfo)
	return userInfo, err
}

func (handler *MongodbHandler) RemoveOpsUserInfoByName(collect string, name string) error {
	query := bson.M{"userName": name}
	return handler.remove(collect, query)
}

func (handler *MongodbHandler) GetOpsAreaBindDatas(collect string, area string, beginmonth string, endmonth string) ([]AreaMonthData, error) {
	query := bson.M{"area": area, "month": bson.M{"$gte": beginmonth, "$lte": endmonth}}
	var binddatas []AreaMonthData
	err := handler.findAll(collect, query, &binddatas)
	return binddatas, err
}

func (handler *MongodbHandler) GetOpsAppKey(collect string, appkey string) (AppKeyinfo, error) {
	query := bson.M{"appkey": appkey}
	var appKeyinfo AppKeyinfo
	err := handler.findOne(collect, query, &appKeyinfo)
	return appKeyinfo, err
}

func (handler *MongodbHandler) GetServicePhoneByAcIP(collect string, acIp string) (ServicePhoneInfo, error) {
	query := bson.M{"acIp": acIp}
	servicePhoneInfo := ServicePhoneInfo{}
	err := handler.findOne(collect, query, &servicePhoneInfo)
	return servicePhoneInfo, err
}

/*

//获取所有帐号
func (handler *MongodbHandler) GetAvailableRouters() ([]Router, error) {
	s := handler.getFreshSession()
	defer s.Close()
	routers := []Router{}
	err := s.DB(*config.DbName).C(*config.DbRouterCollection).Find(nil).All(&routers)
	return routers, err
}

func (handler *MongodbHandler) GetRouterByMac(mac string) (Router, error) {
	s := handler.getFreshSession()
	defer s.Close()
	r := Router{}
	err := s.DB(*config.DbName).C(*config.DbRouterCollection).Find(bson.M{"mac": mac}).One(&r)
	return r, err
}

func (handler *MongodbHandler) GetRouterByClientMac(clientMac string) (Router, error) {
	s := handler.getFreshSession()
	defer s.Close()
	r := Router{}
	err := s.DB(*config.DbName).C(*config.DbRouterCollection).Find(bson.M{CLIENTS_MAC: clientMac}).One(&r)
	return r, err
}

func (handler *MongodbHandler) AddRouter(r Router) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(*config.DbRouterCollection).Insert(r)
}

func (handler *MongodbHandler) UpdateRouter(r Router) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(*config.DbRouterCollection).Update(bson.M{"mac": r.Mac}, r)
}

//更新帐号
func (handler *MongodbHandler) UpdateClientOnlineTime(clientMac, clientIp string, time time.Time) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(*config.DbRouterCollection).Update(bson.M{CLIENTS_MAC: clientMac}, bson.M{"$set": bson.M{"clients.$.ip": clientIp}, "$push": bson.M{"clients.$.onlinetime": time}})

}

func (handler *MongodbHandler) UpdateClientOfflineTime(clientMac, clientIp string, time time.Time) error {
	s := handler.getFreshSession()
	defer s.Close()
	return s.DB(*config.DbName).C(*config.DbRouterCollection).Update(bson.M{CLIENTS_MAC: clientMac}, bson.M{"$set": bson.M{"clients.$.ip": clientIp}, "$push": bson.M{"clients.$.offlinetime": time}})
}



func (handler *MongodbHandler) UpdateClientTime(clientMac, clientIp string, isOnlineTime bool) {
	//目前只保存数据
	// 查询对应集合中 euht router 和 手机用户对应关系
	r, err := handler.GetRouterInRelation(clientMac)
	// 有对应关系
	if err == nil {
		// 查询euht集合中是否存在euht router
		logger.Debug("find relation router: %s : client: %s in Collection : %s\n", r.RouterMac, clientMac, *config.DbRelationCollection)
		router, err := handler.GetRouterByMac(r.RouterMac)
		//查到有router
		if err == nil {
			logger.Debug("find  router: %s  in Collection : %s\n", r.RouterMac, *config.DbRouterCollection)
			index := -1
			//查询用户信息
			for i, v := range router.Clients {
				if v.Mac == clientMac {
					index = i
					break
				}
			}
			if index != -1 {
				logger.Debug("find  client : %s  connect  router: %s\n", clientMac, r.RouterMac)
				if isOnlineTime {
					router.Clients[index].OnlineTime = append(router.Clients[index].OnlineTime, time.Now())
				} else {
					router.Clients[index].OfflineTime = append(router.Clients[index].OfflineTime, time.Now())
				}
			} else {
				logger.Debug("can not find  client : %s  connect router: %s\n", clientMac, r.RouterMac)
				client := Client{Mac: clientMac, Ip: clientIp}
				if isOnlineTime {
					client.OnlineTime = append(client.OnlineTime, time.Now())
				} else {
					client.OfflineTime = append(client.OfflineTime, time.Now())
				}
				router.Clients = append(router.Clients, client)
			}
			handler.UpdateRouter(router)
		} else {
			logger.Debug("can not find  router: %s  in Collection: %s\n", r.RouterMac, *config.DbRouterCollection)
			client := Client{Mac: clientMac, Ip: clientIp}
			if isOnlineTime {
				client.OnlineTime = append(client.OnlineTime, time.Now())
			} else {
				client.OfflineTime = append(client.OfflineTime, time.Now())
			}
			router.Clients = append(router.Clients, client)
			router.Mac = r.ClientMac
			handler.AddRouter(router)
		}
	} else {
		// 没有对应关系
		logger.Debug("can not find relation router: %s  client: %s in Collection: %s \n", r.RouterMac, clientMac, *config.DbRelationCollection)
		// 查询有没有client
		router, err := handler.GetRouterByClientMac(clientMac)
		if err == nil {
			logger.Debug("find  client : %s  in Collection: %s\n", clientMac, *config.DbRouterCollection)
			if isOnlineTime {
				handler.UpdateClientOnlineTime(clientMac, clientIp, time.Now())
			} else {
				handler.UpdateClientOfflineTime(clientMac, clientIp, time.Now())
			}

		} else {
			logger.Debug("can not find  client : %s  in Collection: %s\n", clientMac, *config.DbRouterCollection)
			client := Client{Mac: clientMac, Ip: clientIp}
			client.OnlineTime = append(client.OnlineTime, time.Now())
			router.Clients = append(router.Clients, client)
			handler.AddRouter(router)
		}
	}

}

func (handler *MongodbHandler) UpdateClientAcctInfo(clientMac string, inputoctets, outputocts, acctsessiontime uint32) error {
	s := handler.getFreshSession()
	defer s.Close()
	info := ClientAcctInfo{AcctSessionTime: acctsessiontime, InputOctets: inputoctets, OutputOctets: outputocts}
	return s.DB(*config.DbName).C(*config.DbRouterCollection).Update(bson.M{CLIENTS_MAC: clientMac}, bson.M{"$push": bson.M{"clients.$.clientacctinfos": info}})
}
*/
