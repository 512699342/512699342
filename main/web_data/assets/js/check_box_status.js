var cb = document.getElementById("user_checkbox");
var regBTN = $("#reg_button");
/*
$(document).ready(function () {
    console.log("1");
    regBTN.css("background", "rgba(255, 255, 255, 0.3)")
        .css("border", "1px solid rgba(255, 255, 255, 0.6)")
        .css("color", "rgba(255, 255, 255, 0.5)")
        .text("获取上网资格")
        .attr('disabled', true)
        .css('cursor', 'not-allowed');
});
*/

document.getElementById("user_checkbox").onchange = function () {
    if (cb.checked) {
        regBTN.css("background", "rgba(255, 255, 255, 1)")
            .css("border", "1px solid rgba(255, 255, 255, 1)")
            .css("color", "rgba(43, 146, 255, 1)")
            .text("获取上网资格")
            .removeClass("disabled")
            .removeAttr("disabled style");
    } else {
        regBTN.css("background", "rgba(255, 255, 255, 0.3)")
            .css("border", "1px solid rgba(255, 255, 255, 0.6)")
            .css("color", "rgba(255, 255, 255, 0.5)")
            .text("获取上网资格")
            .attr('disabled', true)
            .css('cursor', 'not-allowed');
    }
}