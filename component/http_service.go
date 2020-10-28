package component

import (
	"component/auth"
	"config"
	"fmt"
	"github.com/wonderivan/logger"
	"net/http"
	"time"
)

//http route
func SetHttpHandler() {
	// 开启一个文件服务器
	http.Handle("/web_data/", http.StripPrefix("/web_data/", http.FileServer(http.Dir("web_data"))))

	//发送验证码
	http.HandleFunc("/send_validate_code", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		if handler, ok := auth.ExtraAuth.(auth.HttpSendValidateCodeHandler); ok {
			handler.HandleSendValidateCode(w, r)
		} else {
			BASIC_SERVICE.HandleSendValidateCode(w, r)
		}
	})
	//处理用户注册提交信息
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		if handler, ok := auth.ExtraAuth.(auth.HttpRegisterHandler); ok {
			handler.HandleRegister(w, r)
		} else {
			BASIC_SERVICE.HandleRegister(w, r)
		}
	})
	//推送用户注册页面
	http.HandleFunc("/registerpage", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		if handler, ok := auth.ExtraAuth.(auth.HttpRegisterPageHandler); ok {
			handler.HandleRegisterPage(w, r)
		} else {
			BASIC_SERVICE.HandleRegisterPage(w, r)
		}
	})
	//推送用户注册成功页面
	http.HandleFunc("/register_success", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		BASIC_SERVICE.HandleRegisterSuccess(w, r)
	})
	//退出
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		if handler, ok := auth.ExtraAuth.(auth.HttpLogoutHandler); ok {
			handler.HandleLogout(w, r)
		} else {
			BASIC_SERVICE.HandleLogout(w, r)
		}
	})
	if extrahttp, ok := auth.ExtraAuth.(auth.ExtraHttpHandler); ok {
		extrahttp.AddExtraHttp()
	}
	//主页处理，portal服务器路径
	// test
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		if handler, ok := auth.ExtraAuth.(auth.HttpRootHandler); ok {
			handler.HandleRoot(w, r)
		} else {
			BASIC_SERVICE.HandleRoot(w, r)
		}
	})
	//随机数处理
	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		BASIC_SERVICE.HandleRandom(w, r)
	})
	//更新RSA加密-密钥处理
	http.HandleFunc("/update_encrypt_key", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		BASIC_SERVICE.HandleUpdateEncryptKey(w, r)
	})
	//监控http服务，网管心跳接口
	http.HandleFunc("/monitor", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			auth.ErrorWrap(w)
		}()
		BASIC_SERVICE.HandleMonitor(w, r)
	})
}

// 自定义http server 处理Http请求函数
func StartHttpServer(addr string) {
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}
	logger.Info("start http server on %s\n", addr)
	err := s.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

//处理Http请求函数
func StartHttp() {
	logger.Info("listen http on %d\n", *config.HttpNameAuthPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.HttpNameAuthPort), nil)
	if err != nil {
		logger.Error(err)
	}
}
