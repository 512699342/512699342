var phoneBackgroundHeight = 1068;

function IsPC() {
    var userAgentInfo = navigator.userAgent;
    //console.log(userAgentInfo);
    var Agents = ["Android", "iPhone",
        "SymbianOS", "Windows Phone",
        "iPad", "iPod"
    ];
    var flag = true;
    for (var v = 0; v < Agents.length; v++) {
        if (userAgentInfo.indexOf(Agents[v]) > 0) {
            flag = false;
            break;
        }
    }
    return flag;
}

function config_pc_bg_size() {
    var device_width = window.screen.availWidth;
    var device_height = window.screen.availHeight;

    var width = device_width + "px";
    var height = device_height + "px";

    //$('#bg_img').css('min-width', width).css('min-height', height);
    document.getElementById('bg_img').style.minHeight = height;
}

function config_phone_bg_size() {
    //var height = (phoneBackgroundHeight/75) + "rem";
    var height = (document.documentElement.clientHeight);

    //$('#bg_img').css('height', height);
    document.getElementById('bg_img').style.height = height + "px";
}

function config_phone_footer_place() {
    var footer_height = document.documentElement.clientHeight * 0.9;

    //$('#tc_footer').css('top', footer_height);
    document.getElementById('tc_footer').style.top = footer_height + "px";
}

function view_area_config() {
    res = IsPC();
    if (res) {
        config_pc_bg_size();
    } else {
        config_phone_bg_size();
        config_phone_footer_place();
    }
}

view_area_config();