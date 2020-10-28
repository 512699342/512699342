//获取URL中的字段
function GetQueryString(name) {
    var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
    var r = window.location.search.substr(1).match(reg); //search,查询?后面的参数，并匹配正则
    if (r != null) return unescape(r[2]);
    return null;
}


/*IP正则校验*/
function isValidIP(ip) {
    var reg = /^(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])$/
    return reg.test(ip);
}


// 判断是否为手机号
function isValidPhoneNum(phone) {
    var myreg = /^[1][3,4,5,6,7,8,9][0-9]{9}$/;
    if (!myreg.test(phone)) {
        return false;
    } else {
        return true;
    }
}

function query_ac_status() {
    var village_ac_ip = document.getElementById("village_ac_ip").value;
    document.getElementById("village_ac_status").value = '';
    document.getElementById("ac_search_status").value = '';

    if (village_ac_ip == null || village_ac_ip == "" || (isValidIP(village_ac_ip) == false)) {
        alert("请输入正确的IP！！！");
        return;
    }
    url = "/query_ac_status?" + "acip=" + village_ac_ip;
    var html = "搜索中......";
    document.getElementById("ac_search_status").value = html
    $.ajax({
        type: "GET", //方法类型
        url: url, //url  "/users/login"
        success: function (result) {
            // var junior = JSON.parse(result);
            console.log(result);

            document.getElementById("village_ac_status").value = result;
            var html = "搜索完毕";
            document.getElementById("ac_search_status").value = html
        },
        error: function () {
            alert("异常！");
        }
    });
}


function query_phone_register() {
    var user_phone_number = document.getElementById("user_phone_number").value;
    document.getElementById("user_phone_search_status").value = ""
    $("#span_phone_register").html("");
    if (user_phone_number == null || user_phone_number == "" || (isValidPhoneNum(user_phone_number) == false)) {
        alert("请输入正确的手机号码！！！");
        return;
    }

    url = "/query_phone_register?" + "user_phone_number=" + user_phone_number;

    $.ajax({
        type: "GET", //方法类型
        // dataType: "json",//预期服务器返回的数据类型 json
        url: url,
        success: function (result) {
            // var junior = JSON.parse(result);
            console.log(result);
            var json = result;
            if (json.result == null) {
                // $("#span_phone_register").html("无注册信息");
                document.getElementById("user_phone_search_status").value = "无注册信息"
                $("#span_phone_register").html("");
                return
            }
            document.getElementById("user_phone_search_status").value = ""
            console.log(json.result.length)
            var html = "  <table width=\"700\" border=\"1\" align=\"center\" cellpadding=\"0\" cellspacing=\"0\"  id=\"tbodydata\"> <thead><tr><th>序号</th> <th>村路由器IP</th> <th>手机号</th> <th>MAC地址</th> <th>注册时间</th></tr> ";
            for (var i = 0; i < json.result.length; i++) {
                var phone = json.result[i].phone;
                var clientMac = json.result[i].mac;
                var ac_ip = json.result[i].acip;
                var bind_time = json.result[i].registertime;

                console.log(i, phone, clientMac);
                html += "<tr>";
                html += "<td>" + (i + 1) + "</td>";
                html += "<td>" + ac_ip + "</td>";
                html += "<td>" + phone + "</td>";
                html += "<td>" + clientMac + "</td>";
                html += "<td>" + bind_time + "</td>";
                html += "</tr>";

            }
            html += "</table>";

            $("#span_phone_register").html(html);
        },
        error: function () {
            alert("异常！");
        }
    });

}

// 取得cookie
function getCookie(name) {
    var nameEQ = name + '='
    var ca = document.cookie.split(';') // 把cookie分割成组
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i] // 取得字符串
        while (c.charAt(0) == ' ') { // 判断一下字符串有没有前导空格
            c = c.substring(1, c.length) // 有的话，从第二位开始取
        }
        if (c.indexOf(nameEQ) == 0) { // 如果含有我们要的name
            return unescape(c.substring(nameEQ.length, c.length)) // 解码并截取我们要值
        }
    }
    return false
}

// 清除cookie
function clearCookie(name) {
    setCookie(name, "", -1);
}

// 设置cookie
function setCookie(name, value, seconds) {
    seconds = seconds || 0; //seconds有值就直接赋值，没有为0，这个根php不一样。
    var expires = "";
    if (seconds != 0) { //设置cookie生存时间
        var date = new Date();
        date.setTime(date.getTime() + (seconds * 1000));
        expires = "; expires=" + date.toGMTString();
    }
    document.cookie = name + "=" + escape(value) + expires + "; path=/"; //转码并赋值
}



$(document).ready(function () {
    document.getElementById("user_register_count").value = '';
    name = getCookie("login_username");
    // console.log("ready:", name);
    document.getElementById("user_center_btn").innerText = name;

    url = "/query_register_count?";

    $.ajax({
        type: "GET", //方法类型
        url: url,
        success: function (result) {
            console.log(result);
            console.log("result:", name);
            document.getElementById("user_register_count").value = result;
            document.getElementById("user_center_btn").value = name;
        },
        error: function () {
            alert("异常！");
        }
    });

});

function handleEnter(field, event) {
    var keyCode = event.keyCode ? event.keyCode : event.which ? event.which :
        event.charCode;
    if (keyCode == 13) {
        var i;
        for (i = 0; i < field.form.elements.length; i++)
            if (field == field.form.elements[i])
                break;
        i = (i + 1) % field.form.elements.length;
        field.form.elements[i].focus();
        return false;
    } else
        return true;
}

function user_center() {
    //获取当前登录的用户名，如果用户名是admin，则跳转到user_list.html,
    //否则跳转到alter_user.html
    h_url = window.location.href
    if ("admin" == getCookie("login_username")) {
        h_url = h_url.substring(0, h_url.indexOf("/index")) + "/user_list.html"
    } else if ("" == getCookie("login_username")) {
        h_url = h_url
    } else {
        h_url = h_url.substring(0, h_url.indexOf("/index")) + "/alter_user.html"
    }
    window.location.href = h_url + "?code=" + ((new Date()).getTime());
}

function loginout() {
    //alert("user_center")

    url = "/loginout";
    $.ajax({
        type: "POST", //方法类型
        url: url, //url  "/users/login"
        success: function (result) {
            // var junior = JSON.parse(result);
            console.log(result);
            if (result == "loginout success") {

                alert("退出成功！");
            } else {
                alert("登录会话已失效！请重新登录！");
            }
            //清除cookie
            clearCookie("login_username")

            console.log(window.location.href);
            h_url = window.location.href
            h_url = h_url.substring(0, h_url.indexOf("/index"))
            href = h_url; //window.location.href + "index";
            console.log(href);

            window.location.href = href;

        },
        error: function () {
            alert("异常！");
        }
    });

}

function clear_cookie() {
    var expdate = new Date();
    expdate.setTime(expdate.getTime() - 1);
    SetCookie("login_username", "", -1);
}