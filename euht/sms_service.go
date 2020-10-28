package euht

import (
	"bytes"
	"component/mongodb"
	"config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/wonderivan/logger"
)

const (
	VALIDCODE_IS_MATCH  = iota
	VALIDCODE_NOT_FIND  = iota
	VALIDCODE_NOT_MATCH = iota
)

//HTTPS POST 云之讯返回json格式
type respSmsHttpsJson struct {
	Code        interface{} `json:"code"`
	Count       interface{} `json:"count"`
	Create_date interface{} `json:"create_date"`
	Mobile      interface{} `json:"mobile"`
	Smsid       interface{} `json:"smsid"`
	Msg         interface{} `json:"msg"`
	Uid         interface{} `json:"uid"`
}

//发送短信返回状态，包括短信状态、校验码、UUID、手机号码
type respSmsStatus struct {
	SmsStatus    bool
	ValidateCode string
	UuidCode     string
	MobileNumber string
}

//生产四位数验证码
func genValidateCode(width int) string {
	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	numberstr := ""
	num := ""

	for i := 0; i < width; i++ {
		num = fmt.Sprintf("%d", numeric[rand.Intn(r)])
		numberstr = numberstr + num
	}

	//fmt.Println("验证码11：", numberstr)

	return numberstr
}

/* 云之讯短信接口 ------------------------------ start */
//云之讯接口
//创建唯一uuid
func get_uuid() (string, error) {

	uuidcode := ""

	// Creating UUID Version 4
	u1, err := uuid.NewV4()
	if err != nil {
		logger.Error("sms(yzx), get uuid fail: %v", err)
	} else {
		uuidcode = fmt.Sprintf("%s", u1)
	}

	str := strings.Replace(uuidcode, "-", "", -1)

	return str, err
}

//云之讯接口
//https post json
func postJson(url string, body []byte) (string, error) {
	// 新建一个http post请求到对应链接
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	// 设置http头
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json;charset=utf-8")
	// 新建http客户端
	client := &http.Client{}
	// 发送一个http请求并且接收返回的响应数据
	resp, err := client.Do(req)
	// 判断请求响应是否成功
	if err != nil {
		logger.Error("sms(yzx), http post fail: %v", err)
		return "", err
	}
	// 关闭http连接
	defer resp.Body.Close()
	// 获取相应数据的响应体
	respbody, _ := ioutil.ReadAll(resp.Body)
	return string(respbody), err
}

//云之讯接口
//发送短信函数
func sendSms(validatecode string, mobile string, uid string, service_phone_number string) (string, error) {

	result := ""

	instance_1 := map[string]interface{}{"sid": "",
		"token":      "",
		"appid":      "",
		"templateid": "",
		"param":      "",
		"mobile":     "",
		"uid":        ""}

	instance_1["sid"] = *config.EuhtSmsYzxAccountSid
	instance_1["token"] = *config.EuhtSmsYzxAuthToken
	instance_1["appid"] = *config.EuhtSmsYzxAppid
	instance_1["templateid"] = *config.EuhtSmsYzxTemplateid

	params := validatecode + "," + service_phone_number

	//赋值验证码，手机号码，uuid值
	instance_1["param"] = params
	instance_1["mobile"] = mobile
	instance_1["uid"] = uid
	// 将已经初始化和赋值后的字典数据转换为json数据
	jsonStr, err := json.Marshal(instance_1)

	if err != nil {
		logger.Error("sms(yzx), json marshal fail: %v", err)
		return result, err
	}
	// 将封装好的json数据发送到云之讯-短信请求url接口
	return postJson(*config.EuhtSmsYzxUrl, jsonStr)

}

//云之讯接口
//输入手机号码，返回发送短信状态， uuid, 手机号码
func yzx_SendSms(mobile string, validatecode string, service_phone_number string) (respSmsStatus, error) {

	uuidcode, err := get_uuid()
	// 新建一个respSmsStatus结构体用于存储调用短信接口返回的响应数据
	respsms := respSmsStatus{}
	respsms.MobileNumber = mobile
	respsms.ValidateCode = validatecode

	respbody, err := sendSms(validatecode, mobile, uuidcode, service_phone_number)

	if err != nil {
		logger.Error("sms(yzx), send fail: %v", err)
		respsms.SmsStatus = false
	} else {

		//logger.Debug(respbody)

		//解析json
		respbody_byte := []byte(respbody)
		stu := respSmsHttpsJson{}
		err = json.Unmarshal(respbody_byte, &stu)
		if err == nil {
			logger.Debug("sms(yzx), mobile: %s validatecode: %s send sms status: %v", mobile, validatecode, stu.Code)
			//判断唯一性
			if stu.Msg == "OK" && stu.Mobile == mobile && stu.Uid == uuidcode {
				respsms.SmsStatus = true
				respsms.UuidCode = uuidcode
			} else {
				respsms.SmsStatus = false
			}
		} else {
			logger.Debug("sms(yzx),  mobile: %s validatecode: %s json unmarshal fail: %v", mobile, validatecode, err)
			respsms.SmsStatus = false
		}
	}

	return respsms, err
}

/* 云之讯短信接口 ------------------------------ end */

/* 聚合数据短信接口 ------------------------------ start */
//聚合数据接口
//发送验证码函数
func juhe_SendSms(mobile string, validatecode string, service_phone_number string) (respSmsStatus, error) {

	respsms := respSmsStatus{}

	respsms.ValidateCode = validatecode
	respsms.MobileNumber = mobile
	codevalue := "#code#=" + validatecode + "&" + "#phone#=" + service_phone_number

	//初始化参数
	param := url.Values{}

	param.Set("mobile", mobile)                        //手机号码
	param.Set("tpl_id", *config.EuhtSmsJuheTemplateid) //短信模板ID
	param.Set("tpl_value", codevalue)                  //验证码
	param.Set("key", *config.EuhtSmsJuheAuthToken)     //应用APPKEY

	//发送请求
	data, err := juhe_httpGet(*config.EuhtSmsJuheUrl, param)
	if err != nil {
		logger.Error("sms(juhe), http get fail: %v", err)
	} else {
		var netReturn map[string]interface{}
		json.Unmarshal(data, &netReturn)
		logger.Debug("sms(juhe), mobile: %s validatecode: %s send sms status: %v", mobile, validatecode, netReturn["reason"])
		if netReturn["error_code"].(float64) == 0 {
			respsms.SmsStatus = true
		} else {
			respsms.SmsStatus = false
		}
	}

	return respsms, err
}

//聚合数据
//http get接口
func juhe_httpGet(apiURL string, params url.Values) (rs []byte, err error) {
	var Url *url.URL
	Url, err = url.Parse(apiURL)
	if err != nil {
		logger.Error("sms(juhe)，analysis url fail:%v", err)
		return nil, err
	}	
	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()
	resp, err := http.Get(Url.String())
	if err != nil {
		//logger.Error("err:",err)
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

/* 聚合数据短信接口 ------------------------------ end */

//发送短信函数
//第三方服务选择，0：接口之家实名&云之讯短信 1：接口之家实名&聚合短信 2：聚合实名&云之讯短信 3：聚合实名&聚合短信
func SendSms(mobile string, service_phone_number string) bool {

	respsms := respSmsStatus{}

	validatecode := genValidateCode(*config.EuhtSmsValidatecodeLen)

	if *config.EuhtSmsServiceChoice == 0 {
		//云之讯短信验证
		respsms, _ = yzx_SendSms(mobile, validatecode, service_phone_number)
	} else {
		//聚合数据短信验证
		respsms, _ = juhe_SendSms(mobile, validatecode, service_phone_number)
	}
	// 如果短信发送成功，那么将数据记录在数据库中
	if respsms.SmsStatus == true {
		pvc := mongodb.PhoneValidateCode{
			Phone:        mobile,
			ValidateCode: validatecode,
			Time:         time.Now(),
		}
		err := mongodb.Db_handler.UpsertPhoneValidateCode(pvc)
		if err != nil {
			logger.Error("phone: %s code: %s : save phone validatecode fail : %s", mobile, validatecode, err.Error())
		} else {
			logger.Debug("phone: %s code: %s : save phone validatecode success", mobile, validatecode)
		}
	}
	return respsms.SmsStatus
}

// mongodb数据库查询验证码
func QueryValidateCode(mobile string, validatecode string) (bool, int) {
	pvc, err := mongodb.Db_handler.GetPhoneValidateCode(mobile)
	if err != nil {
		logger.Error("phone: %s code: %s : get phone validatecode fail : %s", mobile, validatecode, err.Error())
		return false, VALIDCODE_NOT_FIND
	} else {
		if pvc.ValidateCode == validatecode {
			return true, VALIDCODE_IS_MATCH
		} else {
			//logger.Error("check phone validatecode fail, phone: %s code: %s", mobile, validatecode)
			return false, VALIDCODE_NOT_MATCH
		}
	}
}

// mongodb数据库移除验证码
func RemoveValidateCode(mobile string) bool {
	err := mongodb.Db_handler.RemovePhoneValidateCode(mobile)
	if err != nil {
		logger.Error("phone: %s  remove phone validatecode fail : %s", mobile, err.Error())
		return false
	} else {
		return true
	}
}
