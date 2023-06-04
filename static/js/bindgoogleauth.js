$(function() {
    $("#btnSubmit").click(submit)
});
function submit() {
    if ($("#txtLoginPwd").val() == "") {
        myAlert.warning($("#txtLoginPwd").attr("placeholder"));
        return
    }
    myConfirm.show({
        title: "确定修改？",
        sure_callback: function() {
            common.postJson("manager/bindgoogleauth", "divContainer", function(d) {
                if (d.success == 1) {
                    location.href = location.href
                } else {
                    myAlert.error(d.result)
                }
            })
        }
    })
}
;