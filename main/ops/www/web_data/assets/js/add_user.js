function validate_email(value) {
	var patn = /^\w+((-\w+)|(\.\w+))*\@[A-Za-z0-9]+((\.|-)[A-Za-z0-9]+)*\.[A-Za-z0-9]+$/;
	if (!patn.test(value))
		return false; //判断Email是否合法
	return true;
}

function validate_phone(phone) {
	var myreg = /^[1][3,4,5,7,8][0-9]{9}$/;
	if (!myreg.test(phone)) {
		return false;
	} else {
		return true;
	}
}

//添加用户
function add_user() {
	var user_name = $("#userName").val();
	var user_password = $("#userPassword").val();
	var user_re_password = $("#userRePassword").val();
	var full_name = $("#fullName").val();
	var user_phone = $("#userPhone").val();
	var user_email = $("#userEmail").val();

	//验证是否含有`~!#$%^&*()+-/=;\':\"?,.<>[]{}| 等字符
	var pattern = /[\`~!#\$%\^&\*\(\)\+-/=;\\':\?,\.<>\[\]{}\|\x22]+/;
	if (user_name.match(pattern)) {
		alert("用户名包含非法字符 `~!#$%^&*()+-/=;\':\"?,.<>[]{}| ");
		return;
	}

	if (full_name.match(pattern)) {
		alert("姓名包含非法字符 `~!#$%^&*()+-/=;\':\"?,.<>[]{}| ");
		return;
	}

	//验证汉字的正则表达式："[\u4e00-\u9fa5]"
	var Chinese_pattern = /[\u4e00-\u9fa5]+/;
	if (user_name.match(Chinese_pattern)) {
		alert("用户名包含中文字符");
		return;
	}

	if (user_name == "" || user_password == "" ||
		user_re_password == "" || full_name == "") {
		alert("请输入信息，星号*处不能为空！");
		return;
	}
	if (user_password != user_re_password) {
		alert("两次输入密码不一致！");
		return;
	}

	if (user_name.length < 5 || user_name.length > 12) {
		alert("用户名长度5-12个字符！");
		return;
	}

	if (user_password.length < 5 || user_password.length > 12) {
		alert("密码长度5-12个字符！");
		return;
	}

	if (full_name.length > 6) {
		alert("姓名长度大于6个字符！");
		return;
	}

	if (user_email.length > 20) {
		alert("电子邮箱长度大于20个字符！");
		return;
	}

	if (user_phone != "") {
		if (validate_phone(user_phone) == false) {
			alert("请输入正确的手机号码！");
			return;
		}
	}

	if (user_phone != "") {
		if (validate_phone(user_phone) == false) {
			alert("请输入正确的手机号码！");
			return;
		}
	}

	if (user_email != "") {
		if (validate_email(user_email) == false) {
			alert("请输入正确的电子邮箱！");
			return;
		}
	}

	console.log("add_user", user_name, user_password, full_name);

	url = "/add_user?" + "username=" + user_name + "&" + "password=" + user_password +
		"&" + "fullname=" + full_name + "&" + "userphone=" + user_phone +
		"&" + "useremail=" + user_email;
	$.ajax({
		type: "POST",
		url: url,
		success: function (result) {

			if (result == "add_user success") {
				alert("成功添加 " + user_name);
				h_url = window.location.href;
				href = h_url.substring(0, h_url.indexOf("/add_user.html")) + "/user_list.html";
                href = href + "?code=" + ((new Date()).getTime());
				window.location.href = href;
			} else if (result == "user already exists") {
				alert(user_name + " 用户名已存在！");
			} else {
				alert("添加失败！");
			}
		},
		error: function () {
			alert("异常！");
		}
	});

}

//返回
function return_index() {
	username = getCookie("login_username")
	console.log("login_username:", username);
	h_url = window.location.href;
	if (username == "admin") {
		href = h_url.substring(0, h_url.indexOf("/add_user.html")) + "/user_list.html";
	} else {
		href = h_url.substring(0, h_url.indexOf("/add_user.html")) + "/index";
	}
    href = href + "?code=" + ((new Date()).getTime());
	window.location.href = href;
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