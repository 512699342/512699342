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

function jump() {
    var welcome_url = "/?" + "usermac=" + user_mac + "&wlanuserip=" + user_ip + "&wlanacname=" + ac_name + "&wlanacip=" + ac_ip + "&randomID=" + randomID;
    location.href = welcome_url;
}

setTimeout(jump, 1000);