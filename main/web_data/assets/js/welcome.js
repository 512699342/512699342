var location_url = location.origin
var redirection_login = "/login?"

var user_mac;
var user_ip;
var ac_name;
var ac_ip;
var now_date;
//获取URL中的字段
function GetQueryString(name)
{
    var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
    var r = window.location.search.substr(1).match(reg);    //search,查询?后面的参数，并匹配正则
    if(r!=null)return  unescape(r[2]);
    return null;
}

//产生一个随机数，防止浏览器判断为统一网址而使用页面缓存
function GetNonDuplicateID()
{
    var tmp_randomID = Number(Math.random().toString().substr(3) + Date.now()).toString(36);
    return tmp_randomID;
}

user_mac = GetQueryString('usermac');
user_ip = GetQueryString('wlanuserip');
ac_name = GetQueryString('wlanacname');
ac_ip = GetQueryString('wlanacip');
randomID = GetNonDuplicateID();

function show_result_message(text) {
    var msg_text = document.getElementById('status_text');
    msg_text.innerText = text;
}

function user_connect() {
    var connect_url = "/connect?" + "usermac=" + user_mac + "&wlanuserip=" + user_ip + "&wlanacname=" + ac_name + "&wlanacip=" + ac_ip + "&randomID=" + randomID;
    /**
     * 通过ajax向服务器请求连接结果
     * "connect_err=1" -> 进行CMCC Portal认证失败
     */

    $.ajax({
        type: "GET",
        url: connect_url,
        success: function (result_data) {
            show_result_message("连接成功，请关闭此页面");
            clearInterval(get_connect_msg);
        },
        error: function (jqXHR) {
            switch (jqXHR.responseText) {
                case "connect_err=1":
                    show_result_message("正在连接中，请稍后");
                    break;
                default:
                    clearInterval(get_connect_msg);
                    //show_result_message("连接中，请稍后.....");
                    break;
            }
        }
    })
}

var get_connect_msg = setInterval( function() {
    user_connect();
}, 1000);