function disableAllWidget() {
    var regBTN = $("#reg_button");
    var capBtn = $("#captcha_btn");
    var nameInput = $("#name_input");
    var phoneInput = $("#phone_num_input");
    var captchaInput = $("#captcha_input");

    nameInput
        .attr("disabled", true)
        .css("cursor", "not-allowed");

    phoneInput.attr("disabled", true)
        .css("cursor", "not-allowed");

    captchaInput.attr("disabled", true)
        .css("cursor", "not-allowed");

    capBtn.attr('disabled', true)
        .css('cursor', 'not-allowed');

    regBTN.css("background", "rgba(255, 255, 255, 0.3)")
        .css("border", "1px solid rgba(255, 255, 255, 0.6)")
        .css("color", "rgba(255, 255, 255, 0.5)")
        .text("获取上网资格")
        .attr('disabled', true)
        .css('cursor', 'not-allowed');
}

//获取Router的SN号并显示
var user_router_sn = GetQueryString("routersn");
//user_router_sn = "EXXsyfwurfhw8ffhqwef23fw34g"; //测试用字段
if (user_router_sn != null) {
    $("#routersn_text").css('display', 'inline');
    document.getElementById("routersn_text").innerHTML = "当前路由器序列号：" + user_router_sn;
} else {
    $("#routersn_text").css('display', 'inline');
    document.getElementById("routersn_text").innerHTML = "当前路由器序列号：未知（请重新连接WIFI）";
    //若无法获取SN号，禁用所有输入和按钮功能
    disableAllWidget();
}


//获取URL中的字段
function GetQueryString(name) {
    var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)");
    var r = window.location.search.substr(1).match(reg); //search,查询?后面的参数，并匹配正则
    if (r != null) return unescape(r[2]);
    return null;
}


//产生一个随机数，防止浏览器判断为统一网址而使用页面缓存
function GetNonDuplicateID() {
    var tmp_randomID = Number(Math.random().toString().substr(3) + Date.now()).toString(36);
    return tmp_randomID;
}

function GetInputValue(id) {
    var input = document.getElementById(id).value;
    return input;
}

function set_reg_btn_status() {
    var regBTN = $("#reg_button");
    //var uCheckbox = $("user_checkbox");
    //uCheckbox.attr("disable", true);
    regBTN.css("background", "rgba(255, 255, 255, 0.3)")
        .css("border", "1px solid rgba(255, 255, 255, 0.6)")
        .css("color", "rgba(255, 255, 255, 0.5)")
        .text("验证中，请稍后")
        .attr('disabled', true)
        .css('cursor', 'not-allowed');

    setTimeout(function () {
        //   uCheckbox.removeClass("disable");
        regBTN.css("background", "rgba(255, 255, 255, 1)")
            .css("border", "1px solid rgba(255, 255, 255, 1)")
            .css("color", "rgba(43, 146, 255, 1)")
            .text("获取上网资格")
            .removeClass("disabled")
            .removeAttr("disabled style");
    }, 5000);
}

function jump() {
    var user_mac = GetQueryString('usermac');
    var user_ip = GetQueryString('wlanuserip');
    var ac_name = GetQueryString('wlanacname');
    var ac_ip = GetQueryString('wlanacip');
    var randomID = GetNonDuplicateID();
    var router_sn = GetQueryString('routersn');
    var router_mac = GetQueryString('routermac');
    var checksum = GetQueryString('checksum');

    var s_user_true_name = GetInputValue("name_input");
    var user_phone_number = GetInputValue("phone_num_input");
    var user_captcha = GetInputValue("captcha_input");

    if (s_user_true_name == "") {
        show_err_msg("name_err");
        return;
    }
    if (!(/^1(3|4|5|6|7|8|9)\d{9}$/.test(user_phone_number))) {
        show_err_msg("phone_err");
        return;
    }
    if (!(/^\d{4}$/.test(user_captcha))) {
        show_err_msg("captcha_err");
        return;
    }

    //对汉字进行编码，防止后台接收数据乱码
    user_true_name = encodeURIComponent(s_user_true_name);

    register_url = "/register?" + "usermac=" + user_mac + "&wlanuserip=" + user_ip + "&wlanacname=" + ac_name + "&wlanacip=" + ac_ip + "&randomID=" + randomID + "&user_true_name=" + user_true_name + "&user_phone_number=" + user_phone_number + "&user_captcha=" + user_captcha + "&routersn=" + router_sn + "&routermac=" + router_mac + "&checksum=" + checksum;
    login_url = "/register_success?" + "usermac=" + user_mac + "&wlanuserip=" + user_ip + "&wlanacname=" + ac_name + "&wlanacip=" + ac_ip + "&randomID=" + randomID + "&user_true_name=" + user_true_name + "&user_phone_number=" + user_phone_number + "&user_captcha=" + user_captcha + "&routersn=" + router_sn + "&routermac=" + router_mac + "&checksum=" + checksum;

    //点击注册按钮后，更改按钮样式，并设置为不可用，不可用时间持续5s
    set_reg_btn_status();

    /**
     * 通过ajax向服务器请求认证结果
     * "err_no=1" -> 验证码填写错误，弹出提示框，不进行跳转；
     * "err_no=2" -> 姓名填写错误，弹出提示框，不进行跳转；
     * "err_no=3" -> 未知错误，弹出提示框，不进行跳转
     * "err_no=4" -> 今日可认证次数已用完，不进行跳转
     * "err_no=5" -> 此手机号码已达最大绑定数，不进行跳转
     * "err_no=6" -> 验证码失效，需重新获取，不进行跳转
     */
    $.ajax({
        type: "GET",
        url: register_url,
        success: function (result_data) {
            location.href = login_url;
        },
        error: function (jqXHR) {
            switch (jqXHR.responseText) {
                case "err_no=1":
                    show_err_msg("register_captcha_err");
                    break;
                case "err_no=2":
                    show_err_msg("register_name_err");
                    break;
                case "err_no=3":
                    show_err_msg("register_parameter_err");
                    break;
                case "err_no=4":
                    show_err_msg("register_no_count_err");
                    break;
                case "err_no=5":
                    show_err_msg("register_max_router");
                    break;
                case "err_no=6":
                    show_err_msg("register_captcha_re_get");
                    break;
                default:
                    show_err_msg("register_unknow_err");
                    break;
            }
        }
    })
}

function show_err_msg(text_type) {
    text = "#" + text_type;
    $('#err_bgDiv').css('display', 'block'); //遮挡背景层(半透明)

    //fadeIn淡入显示
    $("#err_fgDiv").fadeIn("slow");

    $(text).css('display', 'block');

    //点击窗口确定按钮，fadeIn淡出显示
    $('#err_back_btn').click(function () {
        //隐藏层隐藏起来。
        $('#err_bgDiv').fadeOut("slow"); //遮挡背景层(半透明)
        $('#err_fgDiv').fadeOut("slow"); //逻辑 添加/修改 界面
        $(text).fadeOut("slow");
    });
    /*
        $('#err_fgDiv').css('display', 'block')
            .css('left', (window.innerWidth - 300) / 2 + "px")
            .css('top', (window.innerHeight - 150) / 2 + "px"); //逻辑业务窗口

        $(text).css('display', 'inline');

        //点击窗口确定按钮，隐藏
        $('#err_back_btn').click(function () {
            //隐藏层隐藏起来。
            $('#err_bgDiv').css('display', 'none'); //遮挡背景层(半透明)
            $('#err_fgDiv').css('display', 'none'); //逻辑 添加/修改 界面
            $(text).css('display', 'none');
        });
        */
}

$(document).ready(function () {
    // 绑定事件(点击按钮，发送验证码)
    var captcha_btn = $('#captcha_btn');
    var countId = null;

    if ($.cookie("captcha")) {
        var count = $.cookie("captcha");

        captcha_btn.text(count + '重新获取').attr('disabled', true).css('cursor', 'not-allowed').css("background", "rgba(143, 143, 143, 0.4)");
        var resend = setInterval(function () {
            count--;
            if (count > 0) {
                captcha_btn.text(count + '重新获取').attr('disabled', true).css('cursor', 'not-allowed').css("background", "rgba(143, 143, 143, 0.4)");
                $.cookie("captcha", count, {
                    path: '/',
                    expires: (1 / 86400) * count
                });
            } else {
                clearInterval(resend);
                captcha_btn.text("获取验证码").removeClass('disabled').removeAttr('disabled style').css("background", "rgba(255, 255, 255, 0)");
            }
        }, 1000);
    }

    captcha_btn.click(function () {
        var user_mac = GetQueryString('usermac');
        var user_ip = GetQueryString('wlanuserip');
        var ac_name = GetQueryString('wlanacname');
        var ac_ip = GetQueryString('wlanacip');
        var randomID = GetNonDuplicateID();
        var user_phone_number = document.getElementById("phone_num_input").value;
        var count = 60;

        if (!(/^1(3|4|5|6|7|8|9)\d{9}$/.test(user_phone_number))) {
            show_err_msg("phone_err");
            return;
        }

        url = "/send_validate_code?" + "user_phone_number=" + user_phone_number + "&usermac=" + user_mac + "&wlanuserip=" + user_ip + "&wlanacname=" + ac_name + "&wlanacip=" + ac_ip + "&randomID=" + randomID;
        captcha_btn.attr('disabled', true);
        clearInterval(countId);
        countId = setInterval(function () {
            count--;
            if (count > 0) {
                captcha_btn.text("(" + count + ")" + "重新获取").css("background", "rgba(143, 143, 143, 0.4)");
                $.cookie("captcha", count, {
                    path: '/',
                    expires: (1 / 86400) * count
                });
            } else {
                clearInterval(countId);
                captcha_btn.text("获取验证码").removeAttr('disabled style').css("background", "rgba(255, 255, 255, 0)");
            }
        }, 1000);

        captcha_btn.attr('disabled', true).css('cursor', 'not-allowed');


        $.ajax({
            type: "GET",
            url: url,
            success: function (result_data) {
                show_err_msg("send_captcha_success");
            },
            error: function (jqXHR) {
                show_err_msg("send_captcha_fail");
            }
        });
    })
})