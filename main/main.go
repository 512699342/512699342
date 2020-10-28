package main

import (
	"component"
	"config"
	"euht"
	"fmt"
	"github.com/wonderivan/logger"
	"main/ops"
	"time"
)

func main() {
	// 初始化输出格式
	logger.SetLogger(*config.LogConfig)
	if config.IsValid() {
		//监听CmccPortal端口
		go component.StartCmcc()
		//运维配置(true：运维相关的代码，false:认证相关的代码，现在是false)
		if *config.OpsEnable {
			ops.InitBasic()
			ops.SetHttpHandler()
			go ops.StartHttpServer(fmt.Sprintf(":%d", *config.HttpOpsPort))
		} else {
			// 初始化auth info
			component.InitBasic()
			//接受对应关系
			//是否开启接受对应关系(true:开启, false:关闭,现在是true)
			if *config.EuhtRelationEnable {
				// 侦听获取用户上下线数据，并写道数据库中
				go euht.GetRelation()
			}
			if *config.RadiusEnable {
				//监听radius 授权端口
				go component.StartRadiusAuth()
				//监听radius计费端口
				go component.StartRadiusAcc()
			}
			component.SetHttpHandler()
			//监听Http端口
			if *config.HttpPhoneAuthEnable {
				go component.StartHttpServer(fmt.Sprintf(":%d", *config.HttpPhoneAuthPort))
			}
			//监听Http端口
			if *config.HttpNameAuthEnable {
				go component.StartHttp()
			}
		}

		for true {
			time.Sleep(24 * time.Hour)
		}
	}

}
