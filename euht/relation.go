package euht

import (
	"component/mongodb"
	"config"
	"github.com/wonderivan/logger"
	"net"
	"strconv"
	"strings"
	"time"
)

func GetRelation() {
	// UDP服务
	// config.EuhtRelationPort,对应关系UDP端口-接收村级服务器上下线数据
	// strconv.Itoa()函数的参数是一个整型数字，它可以将数字转换成对应的字符串类型的数字。
	address := ":" + strconv.Itoa(*config.EuhtRelationPort)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	// defer将一个方法延迟到包裹该方法的方法返回时执行
	defer conn.Close()

	for {
		data := make([]byte, 128)
		// 接收侦听到的数据，接收到的数据存储在data里面，n为接受到的数据的长度
		n, rAddr, err := conn.ReadFromUDP(data) //rAddr
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		strData := string(data[:n])
		acip := rAddr.IP.String()

		//logger.Debug("ac:%s , Received: %s", acip, strData)
		// 上线数据： &E83511713X00253@B8:BB:23:14:94:99!192.168.2.150#20:3C:AE:A0:DB:59$192.168.2.152
		// 下线数据： %E83511713X00253@B8:BB:23:14:94:99!192.168.2.150#20:3C:AE:A0:DB:59$192.168.2.152
		// %路由sn@路由mac!路由ip#客户端（手机）mac$客户端（手机）ip\n
		if strings.Contains(strData, "@") {
			// 判断strData中是否包含@，$，判断接收数据是否为上下行数据
			if strings.Contains(strData, "$") {
				//将获取到有效的的字符串数据进行分割处理
				sn_comma := strings.Index(strData, "@")
				rtMac_comma := strings.Index(strData, "!")
				rtIP_comma := strings.Index(strData, "#")
				clMac_comma := strings.Index(strData, "$")
				clIP_comma := strings.Index(strData, "\n")
				curtimestamp := time.Now()
				//写入数据库需要查询数据库，在这里将接收到的数据写入数据库
				temp := mongodb.Relation{
					RouterSn:  strData[1:sn_comma],
					RouterMac: strData[(sn_comma + 1):rtMac_comma],
					RouterIp:  strData[(rtMac_comma + 1):rtIP_comma],
					ClientMac: strData[(rtIP_comma + 1):clMac_comma],
					ClientIp:  strData[(clMac_comma + 1):clIP_comma],
					AcIp:      acip,
					Timestamp: curtimestamp,
				}
				//用户上线记录处理
				if strData[0] == '&' {
					//logger.Debug("Online: ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s ip:%s)", temp.AcIp, temp.ClientIp, temp.ClientMac, temp.RouterSn, temp.RouterMac, temp.RouterIp)
					upateRelation(temp)
				} else if strData[0] == '*' { //用户下线记录处理
					//logger.Debug("Offline: ac(ip:%s) user(ip:%s mac:%s) router(sn:%s mac:%s ip:%s)", temp.AcIp, temp.ClientIp, temp.ClientMac, temp.RouterSn, temp.RouterMac, temp.RouterIp)
					removeRelation(temp)
				} else {
					logger.Debug("fail!!, udp frame errors....")
				}

			} else {
				logger.Debug("Please check msg ‘$’")
			}
		} else {
			logger.Debug("Please check msg ‘@’")
		}
	}
}

//从数据表中查询对应关系，更新数据库
func upateRelation(rtinfo mongodb.Relation) error {

	err := mongodb.Db_handler.UpsertRelation(rtinfo)
	if err != nil {
		logger.Error("insert relation ac:%s, router(ip:%s sn:%s) client(ip:%s mac:%s)  fail: %s", rtinfo.AcIp, rtinfo.RouterIp, rtinfo.RouterSn, rtinfo.ClientIp, rtinfo.ClientMac, err.Error())
	}
	return err
}

//用户下线记录，需要查看数据库如果之前有上线记录，需要该设备上线记录删掉
func removeRelation(rtinfo mongodb.Relation) error {

	err := mongodb.Db_handler.RemoveRelationByUserInfo(rtinfo.ClientMac)
	return err
}
