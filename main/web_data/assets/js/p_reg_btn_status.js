var regBTN = $("#reg_button");

function reg_btn_status_enable() {
    regBTN.css("background", "rgba(255, 255, 255, 1)")
    .css("border", "1px solid rgba(255, 255, 255, 1)")
    .css("color", "rgba(43, 146, 255, 1)")
    .text("获取上网资格")
    .removeClass("disabled")
    .removeAttr("disabled style");
}

function reg_btn_status_disable() {
    regBTN.css("background", "rgba(255, 255, 255, 0.3)")
    .css("border", "1px solid rgba(255, 255, 255, 0.6)")
    .css("color", "rgba(255, 255, 255, 0.5)")
    .text("获取上网资格")
    .attr('disabled', true)
    .css('cursor', 'not-allowed');
}

//页面加载完成后，禁止认证按钮
window.onload=function() {
    reg_btn_status_disable();
}

$('#phone_num_input').on('input propertychange', function() {
    if(($(this).val().length == 11) && ($('#captcha_input').val().length == 4) && ($('#protocol_input').prop('checked'))) {
        reg_btn_status_enable();
    }else {
        reg_btn_status_disable();
    }
});

$('#captcha_input').on('input propertychange', function() {
    if(($(this).val().length == 4) && ($('#phone_num_input').val().length == 11) && ($('#protocol_input').prop('checked'))) {
        reg_btn_status_enable();
    }else {
        reg_btn_status_disable();
    }
});


$('#protocol_input').bind('click',function(){
    if(($(this).prop('checked')) && ($('#captcha_input').val().length == 4) && ($('#phone_num_input').val().length == 11) ){
        reg_btn_status_enable();
    }else{
        reg_btn_status_disable();
    }
});
