package ops

import (
	"config"
	"fmt"
	"github.com/wonderivan/logger"
	"net/http"
	"time"
)

//http route
func SetHttpHandler() {
	http.Handle("/web_data/", http.StripPrefix("/web_data/", http.FileServer(http.Dir("ops/www/web_data"))))
	//主页处理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleRoot(w, r)
		} else {
			panic("")
		}
	})
	//查询mac对应的手机号
	http.HandleFunc("/query_phone_num", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleQueryPhoneNum(w, r)
		} else {
			panic("")
		}
	})

	//手机号注册信息查询
	http.HandleFunc("/query_phone_register", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleQueryPhoneRegisterInfo(w, r)
		} else {
			panic("")
		}
	})

	//查询AC状态
	http.HandleFunc("/query_ac_status", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleQueryAcStatus(w, r)
		} else {
			panic("")
		}
	})

	//查询全省用户终端注册总数
	http.HandleFunc("/query_register_count", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleQueryRegisterCount(w, r)
		} else {
			panic("")
		}
	})

	//用户登录处理
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleLogin(w, r)
		} else {
			panic("")
		}
	})

	//用户退出登录处理
	http.HandleFunc("/loginout", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleLoginOut(w, r)
		} else {
			panic("")
		}
	})

	//用户添加
	http.HandleFunc("/add_user", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleAddUser(w, r)
		} else {
			panic("")
		}
	})

	//用户删除
	http.HandleFunc("/delete_user", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleDeleteUser(w, r)
		} else {
			panic("")
		}
	})

	//用户列表
	http.HandleFunc("/user_list", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleGetUsersList(w, r)
		} else {
			panic("")
		}
	})

	//获取用户信息
	http.HandleFunc("/get_user_info", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleGetUserInfo(w, r)
		} else {
			panic("")
		}
	})

	//修改用户信息
	http.HandleFunc("/alter_user_info", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandleAlterUserInfo(w, r)
		} else {
			panic("")
		}
	})

	//查询终端绑定数量
	http.HandleFunc("/phone_bind_num", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := OPS_Server.(OpsHttpInterface); ok {
			handler.HandlePhoneBindNumInfo(w, r)
		} else {
			panic("")
		}
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
