package ops

import (
	"github.com/wonderivan/logger"
	"net/http"

	"runtime/debug"
)

// define http interface
type OpsHttpInterface interface {
	HandleRoot(w http.ResponseWriter, r *http.Request)
	HandleQueryPhoneNum(w http.ResponseWriter, r *http.Request)
	HandleQueryPhoneRegisterInfo(w http.ResponseWriter, r *http.Request)
	HandleQueryAcStatus(w http.ResponseWriter, r *http.Request)
	HandleQueryRegisterCount(w http.ResponseWriter, r *http.Request)
	HandleLogin(w http.ResponseWriter, r *http.Request)
	HandleLoginOut(w http.ResponseWriter, r *http.Request)
	HandleAddUser(w http.ResponseWriter, r *http.Request)
	HandleDeleteUser(w http.ResponseWriter, r *http.Request)
	HandleGetUsersList(w http.ResponseWriter, r *http.Request)
	HandleGetUserInfo(w http.ResponseWriter, r *http.Request)
	HandleAlterUserInfo(w http.ResponseWriter, r *http.Request)

	HandlePhoneBindNumInfo(w http.ResponseWriter, r *http.Request)
}

var OPS_Server interface{}

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
