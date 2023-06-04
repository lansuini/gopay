$(function() {
    getimgcode()
    $("#txtLoginName").focus();
    $("#btnSubmit").click(function() {
        // $('#audioPlay').play();
        document.getElementById('audioPlay').play()
        submit()
    });
    $("#txtLoginName,#txtLoginPwd").keyup(function(a) {
        if (a.keyCode == 13) {
            submit()
        }
    })
});
function getimgcode(){
    common.getAjax("api/captcha/request",function(d) {
        if (d.success == 1) {
            $("#captcha_img").attr("src",d.data.captchaUrl)
            $("#captchaId").val(d.data.captchaId)
        } else {
            myAlert.error(d.message)
        }
    })
}
function submit() {
    if ($("#txtLoginName").val() == "") {
        myAlert.warning($("#txtLoginName").attr("placeholder"));
        return
    }
    if ($("#txtLoginPwd").val() == "") {
        myAlert.warning($("#txtLoginName").attr("txtLoginPwd"));
        return
    }
    if ($("#captchaId").val() == "") {
        myAlert.warning("请输入验证码");
        return
    }
    if ($("#captchaCode").val() == "") {
        myAlert.warning("请输入验证码");
        return
    }
    common.postJson("login", "divContainer", function(d) {
        if (d.success == 1) {
            location.href = contextPath + "head"
        } else {
            getimgcode()
            // myAlert.error(d.result)
        }
    })
}
;