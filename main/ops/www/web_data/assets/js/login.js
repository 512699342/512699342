//登录
function login() {
    var name = $("#username").val()
    var password = $("#password").val();

    // console.log("login", name, password);

    if (name == "") {
        alert("请输入用户名");
        return;
    }
    if (password == "") {
        alert("请输入密码");
        return;
    }

    clearCookie("login_username")
    setCookie("login_username", name, 3600*24)

    url = "/login?" + "name=" + name + "&" + "password=" + hex_md5(password);
    $.ajax({
        type: "POST", //方法类型
        url: url,
        success: function (result) {
            //alert(result)
            console.log(result);
            console.log(window.location.href);
            if (result == "UserPassword correct") {
                h_url = window.location.href
                console.log(h_url);
                h_url = h_url.substring(0, h_url.indexOf("/index")) + "/index"
                href = h_url + "?code=" + ((new Date()).getTime());
                console.log(href);
                //alert(href)
                window.location.href = href
                //window.location.reload()
            } else if (result == "UserPassword error") {
                alert("用户名或密码错误！")
            } else {
                //alert(result)
                alert("用户名不存在，请联系管理员！")
            }

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