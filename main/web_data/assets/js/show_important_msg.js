/*
$(document).ready(function () {
    res = IsPC();

    $('#popup_bgDiv').css('display', 'block') //遮挡背景层(半透明)
        .width(window.innerWidth + "px")
        .height(window.innerHeight + "px");

    if(res) {
        $('#popup_fgDiv').css('display', 'block')
        .css('left', (window.innerWidth - 430) / 2 + "px")
        .css('top', (window.innerHeight - 540) / 2 + "px"); //逻辑业务窗口
    } else {
        $('#popup_fgDiv').css('display', 'block')
        .css('left', (window.innerWidth - 300) / 2 + "px")
        .css('top', (window.innerHeight - 400) / 2 + "px"); //逻辑业务窗口
    }

    //点击窗口保存按钮，隐藏
    $('#ok_btn').click(function () {
        //隐藏层 隐藏起来。
        $('#popup_bgDiv').css('display', 'none'); //遮挡背景层(半透明)
        $('#popup_fgDiv').css('display', 'none'); //逻辑 添加/修改 界面
    });

});
*/

function show_the_message() {
    res = IsPC();

    $('#popup_bgDiv').css('display', 'block') //遮挡背景层(半透明)
        .width(window.innerWidth + "px")
        .height(window.innerHeight + "px");

    if(res) {
        $('#popup_fgDiv').css('display', 'block')
        .css('left', (window.innerWidth - 430) / 2 + "px")
        .css('top', (window.innerHeight - 540) / 2 + "px"); //逻辑业务窗口
    } else {
        $('#popup_fgDiv').css('display', 'block')
        .css('left', (window.innerWidth - 300) / 2 + "px")
        .css('top', (window.innerHeight - 400) / 2 + "px"); //逻辑业务窗口
    }

    //点击窗口保存按钮，隐藏
    $('#ok_btn').click(function () {
        //隐藏层 隐藏起来。
        $('#popup_bgDiv').css('display', 'none'); //遮挡背景层(半透明)
        $('#popup_fgDiv').css('display', 'none'); //逻辑 添加/修改 界面
    });
}


function IsPC() {
    var userAgentInfo = navigator.userAgent;
    //console.log(userAgentInfo);
    var Agents = ["Android", "iPhone",
                 "SymbianOS", "Windows Phone",
                 "iPad", "iPod"];
    var flag = true;
    for (var v = 0; v < Agents.length; v++) {
        if (userAgentInfo.indexOf(Agents[v]) > 0) {
            flag = false;
            break;
        }
    }
    return flag;
}
