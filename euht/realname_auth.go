package euht

import (
	"config"
	"encoding/json"
	"utility"
	//"fmt"
	"component/mongodb"
	"github.com/wonderivan/logger"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//request:  https://v.apihome.org/phonetwo?realname=************&phone=************&key=************

/*respond
{
    "code": 200,
    "data": {
      "province": "江苏",
      "city": "南京",
      "operator": "中国移动"
    },
     "msg": "一致",
    "ordersign": "20180512210238112683832554313168"
}
*/

//接口之家返回json格式
type respRealnameAuthJkzjJson struct {
	Code      interface{}
	Data      interface{} `json:"data"`
	Msg       interface{} `json:"msg"`
	Ordersign interface{} `json:"ordersign"`
}

type respDataJson struct {
	Province interface{} `json:"province"`
	City     interface{} `json:"city"`
	Operator interface{} `json:"operator"`
}

//聚合数据返回json格式
type respRealnameAuthJuheJson struct {
	Reason     interface{} `json:"reason"`
	Result     juheResult
	Error_code interface{} `json:"error_code"`
}

type juheResult struct {
	Realname interface{} `json:"realname"`
	Mobile   interface{} `json:"mobile"`
	Res      interface{} `json:"res"`
	Resmsg   interface{} `json:"resmsg"`
}

/* 接口之家手机二元素校验接口 ------------------------------ start */
//发送身份验证函数接口
func jkzj_realnameHttpPost(realname string, phone string) (string, error) {

	var clusterinfo = url.Values{}
	clusterinfo.Add("realname", realname)
	clusterinfo.Add("phone", phone)
	clusterinfo.Add("key", *config.EuhtRealnameApikey)
	// 获取byte数据
	var buf io.Reader
	buf = strings.NewReader(clusterinfo.Encode())

	//fmt.Println(buf)
	req, err := http.NewRequest("POST", *config.EuhtRealnameUrl, buf)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return "", err
	}
	defer resp.Body.Close()
	respbody, _ := ioutil.ReadAll(resp.Body)
	return string(respbody), err
}

func jkzj_RealnameAuth(realname string, phone string) (bool, error) {

	hideName := utility.HideName(realname)
	hidePhone := utility.HidePhone(phone)

	realnameStatus := false
	respbody, err := jkzj_realnameHttpPost(realname, phone)

	//fmt.Println(respbody)
	if err != nil {
		logger.Error("realname(jkzj), http get fail: %v", err)
		realnameStatus = false
	} else {
		//解析json
		respbody_byte := []byte(respbody)
		stu := respRealnameAuthJkzjJson{}
		err = json.Unmarshal(respbody_byte, &stu)

		if stu.Msg == "一致" {
			logger.Debug("realname(jkzj), phone[%s]-name[%s] match", hidePhone, hideName)
			realnameStatus = true
		} else {
			logger.Error("realname(jkzj), phone[%s]-name[%s] no match; error msg:%s", hidePhone, hideName, stu.Msg)
			realnameStatus = false
		}
	}

	return realnameStatus, err
}

/* 接口之家手机二元素校验接口 ------------------------------ end */

/* 聚合数据手机二元素校验接口 ------------------------------ start */
//示例
// http://v.juhe.cn/telecom2/query?key=&realname=%E8%91%A3%E5%A5%BD%E5%B8%85&mobile=18912341234
// {
// 	"reason":"成功",
// 	"result":{
// 		"realname":"张汇楼",
// 		"mobile":"15918763096",
// 		"res":1,
// 		"resmsg":"二要素身份验证一致"
// 	},
// 	"error_code":0
// }

//聚合数据接口
//发送验证码函数
func juhe_RealnameAuth(realname string, phone string) (bool, error) {

	hideName := utility.HideName(realname)
	hidePhone := utility.HidePhone(phone)

	realnameStatus := false
	//初始化参数
	param := url.Values{}

	param.Set("mobile", phone)                       //手机号码
	param.Set("realname", realname)                  //短信模板ID
	param.Set("key", *config.EuhtRealnameJuheApikey) //应用APPKEY

	//发送请求
	data, err := juhe_realnameHttpGet(*config.EuhtRealnameJuheUrl, param)
	if err != nil {
		logger.Error("realname(juhe), http get fail: %v", err)
	} else {
		netReturn := respRealnameAuthJuheJson{}
		err = json.Unmarshal(data, &netReturn)
		if err == nil {
			//logger.Error("realname(juhe), phone[%s]-name[%s] no match; error msg:%v", phone, realname, netReturn.Result)
			if netReturn.Result.Res != nil {
				res, ok := netReturn.Result.Res.(float64)
				if ok && res == 1 {
					logger.Debug("realname(juhe), phone[%s]-name[%s] match", hidePhone, hideName)
					realnameStatus = true
				} else {
					logger.Error("realname(juhe), phone[%s]-name[%s] no match; error msg:%v", hidePhone, hideName, netReturn)
					realnameStatus = false
				}
			} else {
				logger.Error("realname(juhe), phone[%s]-name[%s] no match; error msg:%v", hidePhone, hideName, netReturn)
			}
		}
	}

	return realnameStatus, err
}

//聚合数据 手机二元素校验
//http get接口
func juhe_realnameHttpGet(apiURL string, params url.Values) (rs []byte, err error) {
	var Url *url.URL
	Url, err = url.Parse(apiURL)
	if err != nil {
		logger.Error("realname(juhe)，analysis url fail:%v", err)
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

/* 聚合数据手机二元素校验接口 ------------------------------ end */

//调用身份验证接口，返回验证结果
//第三方服务选择，0：接口之家实名&云之讯短信 1：接口之家实名&聚合短信 2：聚合实名&云之讯短信 3：聚合实名&聚合短信
func RealnameAuth(realname string, phone string) (bool, error) {
	var err error
	realnameStatus := false

	hideName := utility.HideName(realname)
	hidePhone := utility.HidePhone(phone)
	logger.Debug("start RealnameAuth: %s %s", hideName, hidePhone)

	if *config.EuhtRealnameServiceChoice == 0 {
		//接口之家手机二元素校验接口
		realnameStatus, err = jkzj_RealnameAuth(realname, phone)
	} else {
		//聚合数据手机二元素校验接口
		realnameStatus, err = juhe_RealnameAuth(realname, phone)
	}

	return realnameStatus, err
}

//使用手机号码，查询实名个数,如已绑定3个，返回false
func IsMaxRealRoutersByPhone(phone string) bool {

	realrouters, err := mongodb.Db_handler.GetRoutersByPhone(phone)

	if err != nil {
		logger.Error("phone: %s not realname...., %s", phone, err.Error())
		return true
	} else {
		realnamenum := len(realrouters)
		logger.Debug("phone: %s realname number[%d] info: %v", phone, realnamenum, realrouters)
		if realnamenum >= *config.EuhtRealnameRouterNumber {
			return false
		} else {
			return true
		}
	}
}

//手机判断实名认证次数
func IsMaxRealnameCountsByPhone(realname string, phone string, relation mongodb.Relation) bool {

	rn, err := mongodb.Db_handler.GetRealNameAuthByPhone(phone)

	rn.AcIp = relation.AcIp
	rn.RouterSn = relation.RouterSn

	if err != nil {
		rn.Phone = phone
		rn.Name = realname
		rn.Date = time.Now()
		rn.Counter = 1
		err := mongodb.Db_handler.InsertRealNameAuth(rn)
		if err != nil {
			logger.Error("relalname(name:%s phone:%s), %s", err.Error())
		}
		return true
	} else {

		if time.Now().Day() != rn.Date.Day() {
			rn.Counter = 0
		} else if rn.Counter >= *config.EuhtRealnameAuthDayCount {
			logger.Error("relalname(name:%s phone:%s),this router can be only try less than 5 times!", realname, phone)
			return false
		}

		rn.Phone = phone
		rn.Name = realname
		rn.Date = time.Now()
		rn.Counter = rn.Counter + 1
		err := mongodb.Db_handler.UpdateRealNameAuth(rn)
		if err != nil {
			logger.Error("relalname(name:%s phone:%s), %s", err.Error())
		}
		return true
	}
}
