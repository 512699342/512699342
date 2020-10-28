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

function deletePhoneCSS() {
    var links = document.getElementsByTagName("link");
    var target = "phone";
    for (var i = 0; i < links.length; i++) {
        if (links[i] && links[i].href && links[i].href.indexOf(target) != -1) {
            links[i].parentNode.removeChild(links[i]);
            i --;
        }
    }
}

function getTrueCss(isPC) {
    var linkNode = document.createElement("link");
    linkNode.setAttribute("rel", "stylesheet");
    linkNode.setAttribute("type", "text/css");

    if (isPC) {
        deletePhoneCSS();
        linkNode.setAttribute("href", "../web_data/assets/css/style_pc.css");
    } else {
        linkNode.setAttribute("href", "../web_data/assets/css/style_phone.css");
    }

    document.head.appendChild(linkNode);
}

res = IsPC();
getTrueCss(res);