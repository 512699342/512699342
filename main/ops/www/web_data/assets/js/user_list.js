function delete_user(param) {
    var user_name = param;

    if (confirm("确定要删除 " + user_name + " 吗？")) {
        //alert("确认删除！");
    } else {
        //alert("取消删除！");
        return;
    }

    url = "/delete_user?username=" + user_name;
    console.log("delete_user", user_name, url);

    $.ajax({
        type: "POST", //方法类型
        url: url, //url  "/users/login"
        success: function (result) {
            // var junior = JSON.parse(result);
            console.log(result);
            console.log(window.location.href);
            if (result != "delete user ok") {
                alert(result);
            }
            href = window.location.href + "";
            console.log(href);
            window.location.href = href;

        },
        error: function () {
            alert("异常！");
        }
    });

}

//写入到Cookie
function SetCookie(name, value, expires) {
    var argv = SetCookie.arguments;

    var argc = SetCookie.arguments.length;
    var expires = (argc > 2) ? argv[2] : null;
    var path = (argc > 3) ? argv[3] : null;
    var domain = (argc > 4) ? argv[4] : null;
    var secure = (argc > 5) ? argv[5] : false;
    document.cookie = name + "=" + escape(value) + ((expires == null) ? "" : ("; expires=" + expires
        .toGMTString())) + ((path == null) ? "" : ("; path=" + path)) + ((domain == null) ? "" : ("; domain=" +
        domain)) + ((secure == true) ? "; secure" : "");
}
//获取一个字符串值在指定字符串第n次出现的位置
function find(str, cha, num) {
    var x = str.indexOf(cha);
    for (var i = 0; i < num; i++) {
        x = str.indexOf(cha, x + 1);
    }
    return x;
}

function alter_user(param) {
    var user_name = param;

    if (confirm("确定要修改 " + user_name + " 吗？")) {
        //alert("确认修改！");
    } else {
        //alert("取消修改！");
        return;
    }

    var trs = document.getElementsByTagName("tr");
    var title = "";
    var content = "";

    for (var j = 0; j < trs.length; j++) {
        var tds = trs[j].getElementsByTagName("td");

        for (var i = 0; i < tds.length; i++) {
            if (j == 0) {
                title += tds[i].innerText + ",";
            } else if (i == 5) {
                content += "operate" + ",";
            } else {
                content += tds[i].innerText + ",";
            }
        }
        if (j > 0) {
            content += "<br/>";
        }
    }
    //alert(title);
    //alert(content);

    select_row = content.substring(content.indexOf(user_name), content.length)
    fullname = select_row.substring(find(select_row, ',', 0) + 1, find(select_row, ',', 1))
    userphone = select_row.substring(find(select_row, ',', 1) + 1, find(select_row, ',', 2))
    useremail = select_row.substring(find(select_row, ',', 2) + 1, find(select_row, ',', 3))

    var expdate = new Date();
    expdate.setTime(expdate.getTime() + 1 * (6 * 60 * 60 * 1000));
    SetCookie("alter_username", user_name, expdate)
    SetCookie("alter_fullname", fullname, expdate)
    SetCookie("alter_userphone", userphone, expdate)
    SetCookie("alter_useremail", useremail, expdate)

    h_url = window.location.href;
    href = h_url.substring(0, h_url.indexOf("/user_list.html")) + "/alter_user.html";
    href = href + "?code=" + ((new Date()).getTime());
    window.location.href = href;

}


$(document).ready(function () {

    $("#span_user_list").html("");

    url = "/user_list?";
    $.ajax({
        type: "GET",
        url: url,
        success: function (result) {

            var json = result;

            var operate_delete_button_head = "<input type=\"button\" onclick=\"delete_user(\'"
            var operate_delete_button_rear = "\')\"  style=\"cursor:pointer\" value=\"删除\">&nbsp;&nbsp;"
            var operate_alter_button_head = "<input type=\"button\" onclick=\"alter_user(\'"
            var operate_alter_button_rear = "\')\"  style=\"cursor:pointer\" value=\"修改\">"

            // console.log(result);
            // console.log(json.result.length)
            // var html =
            //     "  <table width=\"700\" border=\"1\" align=\"center\" cellpadding=\"0\" cellspacing=\"0\"  id=\"tbodydata\"> <thead><tr><th>序号</th> <th>用户名</th> <th>姓名</th> <th>手机号</th> <th>电子邮箱</th> <th>操作</th> </tr> ";
            var html = ""
            for (var i = 0; i < json.result.length; i++) {
                var UserName = json.result[i].UserName;
                var FullName = json.result[i].FullName;
                var UserPhone = json.result[i].UserPhone;
                var UserEmail = json.result[i].UserEmail;

                console.log(UserName, FullName, UserPhone, UserEmail);
                html += "<tr>";
                html += "<td>" + (i + 1) + "</td>";
                html += "<td>" + UserName + "</td>";
                html += "<td>" + FullName + "</td>";
                html += "<td>" + UserPhone + "</td>";
                html += "<td>" + UserEmail + "</td>";

                if (UserName == "admin") {
                    html += "<td>" + operate_alter_button_head + UserName +
                        operate_alter_button_rear + "</td>";
                } else {
                    html += "<td>" + operate_delete_button_head + UserName +
                        operate_delete_button_rear +
                        operate_alter_button_head + UserName + operate_alter_button_rear +
                        "</td>";
                }

                html += "</tr>";

                if (i < g_page_size) {
                    //第一页显示的数量
                    $("#tbody_user_list").append(html);
                    $("#next_page").attr("disabled", "disabled");
                    $("#pre_page").attr("disabled", "disabled");
                    html = ""
                } else {
                    $("#next_page").removeAttr("disabled");
                    $("#pre_page").attr("disabled", "disabled");
                }

                g_count = (i + 1);
                //保存到数组中
                var annex = new Object();
                annex.number = g_count;
                annex.UserName = UserName;
                annex.FullName = FullName;
                annex.UserPhone = UserPhone;
                annex.UserEmail = UserEmail;
                data_list.push(annex);

            }

            //用户总数
            $("#sum_user").text(json.result.length);
            //页总数
            $("#sum_page").text(Math.ceil(json.result.length/g_page_size));
            //显示当前页码
            $("#current_page").text(1);
            g_count = json.result.length;
            g_page = Math.ceil(json.result.length/g_page_size);
            g_current_page = 1;

        },
        error: function () {
            alert("异常！");
        }
    });

});

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



var g_count = 0; //总共有多少条数据
var g_page = 0; //总共有多少页
var g_page_size = 25; //多少条数据为一页
var g_current_page; //当前页码
var data_list = new Array();

//上一页
function pre_page() {
    g_current_page--;
    paging(g_current_page);
}

//下一页
function next_page() {
    g_current_page++;
    paging(g_current_page);
}

//跳转
function jump_page() {
    var jump_index = $("#jump_page_num").val();
    //对输入的数据进行校验，是否合法
    if (jump_index < 1 || jump_index > g_page) {
        alert("非法页码！");
        return
    }
    console.log(jump_index)
    paging(jump_index);
}

//分页
function paging(index) {
    //index页码索引，从1开始
    console.log("g_count, g_page,g_page_size, g_current_page index:", g_count, g_page,g_page_size, g_current_page, index );

    var operate_delete_button_head = "<input type=\"button\" onclick=\"delete_user(\'"
    var operate_delete_button_rear = "\')\"  style=\"cursor:pointer\"  value=\"删除\">&nbsp;&nbsp;"
    var operate_alter_button_head = "<input type=\"button\" onclick=\"alter_user(\'"
    var operate_alter_button_rear = "\')\"  style=\"cursor:pointer\"  value=\"修改\">"

    var html = "";
    var curentNumber = (index-1) * g_page_size;
    var length = curentNumber + g_page_size;
    //当前页数
    g_current_page = index;// + 1;
    for (var i = curentNumber; i < length; i++) {
        // console.log(data_list[i]);
        if (typeof (data_list[i]) == "undefined") {
            break;
        }
        html += "<tr>";
        html += "<td>" + data_list[i].number + "</td>";
        html += "<td>" + data_list[i].UserName + "</td>";
        html += "<td>" + data_list[i].FullName + "</td>";
        html += "<td>" + data_list[i].UserPhone + "</td>";
        html += "<td>" + data_list[i].UserEmail + "</td>";

        if (data_list[i].UserName == "admin") {
            html += "<td>" + operate_alter_button_head + data_list[i].UserName +
                operate_alter_button_rear + "</td>";
        } else {
            html += "<td>" + operate_delete_button_head + data_list[i].UserName +
                operate_delete_button_rear +
                operate_alter_button_head + data_list[i].UserName + operate_alter_button_rear +
                "</td>";
        }
    }

    if (typeof (data_list[length]) == "undefined") {
        //到了最后一页不可以点击
        $("#next_page").attr("disabled", "disabled");
    } else {
        //恢复点击
        $("#next_page").removeAttr("disabled");
    }

    if (index == 1) {
        //到了第一页不可以点击
        $("#pre_page").attr("disabled", "disabled");
    } else {
        $("#pre_page").removeAttr("disabled");
    }
    //填充到表格
    $("#tbody_user_list").html(html);
    //显示当前页数
    $("#current_page").text(g_current_page);
}

function add_user_func() {
    h_url = window.location.href

    if ("admin" == getCookie("login_username")) {
        h_url = h_url.substring(0, h_url.indexOf("/user_list.html")) + "/add_user.html";
        h_url = h_url + "?code=" + ((new Date()).getTime());
    } else {
        h_url = h_url
    }

    window.location.href = h_url;
}

function return_index() {
    h_url = window.location.href

    h_url = h_url.substring(0, h_url.indexOf("/add_user.html")) + "/index";
    h_url = h_url + "?code=" + ((new Date()).getTime());

    window.location.href = h_url;
}
