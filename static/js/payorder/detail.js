$(function() {
    apiPath += "payorder/";
    var a = common.getQuery("orderId");
    if (a == undefined) {
        alert("请求错误");
        location.href = contextPath + "payorder";
        return
    }
    common.getAjax(apiPath + "detail?orderId=" + a, function(b) {
        if (b.success) {
            bindData(b.result);
            // $("#btnSearch").initSearch(apiPath + "getMakeUpList?platformOrderNo="+ b.result.platformOrderNo, getColumns(), {
            //     /* success_callback: buildSummary */
            // })
        } else {
            location.href = contextPath + "payorder"
        }
    });
    common.initDateTime("txtChannelNoticeTime")

    // $("#btnFile").change(function(){
    // var v = $(this).val();
    // var reader = new FileReader();
    // reader.readAsDataURL(this.files[0]);
    // reader.onload = function(e){
    // $('#file_base64').val(e.target.result);
    // };
    // });

    $("#btnSynchronizedState").click(function(){
        common.postAjax(apiPath + "queryorder?orderId=" + a, function(b) {
            console.log(b)
            alert('订单同步成功!')
            location.reload();
        })
    });
});
function bindData(a) {
    $("#divWest").append(common.buildTabNav("订单信息")).append(buildOrder(a)).append(common.buildTabNav("支付信息")).append(buildPay(a));
    $("#divEast").append(common.buildTabNav("商户信息")).append(buildMerchant(a)).append(common.buildTabNav("上游渠道信息")).append(buildChannel(a));
    switch (a.orderStatus) {
    case "WaitPayment":
        if (a.channel != null && a.channel.length > 0) {
            $("#txtPlatformOrderNo").val(a.platformOrderNo);
            $("#txtOrderAmount").val(common.fixAmount(a.orderAmount));
            $("#txtChannel").val(a.channelDesc);
            $("#txtChannelMerchantNo").val(a.channelMerchantNo);
            $("#btnPerfect").show().click(showPerfectModal);
            $("#btnSynchronizedState").show()
            $("#btnSubmit").click(submitPerfect)
        } else {
            $("#btnPerfect").show().click(function() {
                myAlert.warning("该订单未向上游发起，目前不能补单")
            })
        }
        break;
    case "Success":
        $("#btnSendNotify").show().click(function() {
            common.submit(apiPath + "notify?orderId=" + common.getQuery("orderId"))
        });
        break
    }
    initShowModal("showMerchantParam", "回传参数", a.merchantParam);
    initShowModal("showFrontNoticeUrl", "前台通知地址", a.frontNoticeUrl);
    initShowModal("showBackNoticeUrl", "异步通知地址", a.backNoticeUrl);

}
function appendShowLink(b, a) {
    return a != null && a.length > 0 ? "<a id='" + b + "' style='color:#6dc3ea; cursor:pointer'>查看</a>" : "-"
}
function initShowModal(c, b, a) {
    $("#" + c).click(function() {
        $("#showModal").find(".modal-title").html(b).end().find(".modal-body").html(a).end().modal()
    })
}
function buildOrder(b) {
    var a = [];
    a.push({
        key: "平台订单号",
        value: b.platformOrderNo
    });
    a.push({
        key: "平台子订单号",
        value: "-",
        holder: "subOrderNo"
    });
    a.push({
        key: "订单状态",
        value: b.orderStatus
    });
    a.push({
        key: "订单创建时间",
        // value: common.toDateStr("yyyy-MM-dd HH:mm:ss", b.createTime)
        value: common.timeFormatSeconds(b.createTime)
    });
    a.push({
        key: "账务日期",
        value: common.timeFormatDate(b.accountDate)
    });
    // a.push({
    //     key: "交易流水号",
    //     value: b.transactionNo
    // });
    a.push({
        key: "用户IP",
        value: b.userIp
    });
    a.push({
        key: "用户终端",
        value: b.userTerminal
    });
    a.push({
        key: "订单超时时间",
        // value: common.toDateStr("yyyy-MM-dd HH:mm:ss", b.timeoutTime)
        value: common.timeFormatSeconds(b.timeoutTime)
    });
    return common.buildTabPanel(a)
}
function buildPay(b) {
    var a = [];
    a.push({
        key: "支付模式",
        value: b.payModelDesc
    });
    a.push({
        key: "支付方式",
        value: b.payTypeDesc
    });
    a.push({
        key: "支付银行",
        value: b.bankCodeDesc
    });
    a.push({
        key: "银行卡类型",
        value: b.cardTypeDesc
    });
    a.push({
        key: "订单金额",
        value: common.fixAmount(b.orderAmount)
    });
    a.push({
        key: "手续费",
        value: common.fixAmount(b.serviceCharge)
    });
    return common.buildTabPanel(a)
}
function buildMerchant(b) {
    var a = [];
    a.push({
        key: "商户号",
        value: b.merchantNo
    });
    a.push({
        key: "商户订单号",
        value: b.merchantOrderNo
    });
    a.push({
        key: "交易摘要",
        value: b.tradeSummary
    });
    a.push({
        key: "回传参数",
        value: appendShowLink("showMerchantParam", b.merchantParam)
    });
    a.push({
        key: "前台通知地址",
        value: appendShowLink("showFrontNoticeUrl", b.frontNoticeUrl)
    });
    a.push({
        key: "异步通知地址",
        value: appendShowLink("showBackNoticeUrl", b.backNoticeUrl)
    });
    a.push({
        key: "商户请求时间",
        // value: common.toDateStr("yyyy-MM-dd HH:mm:ss", b.merchantReqTime)
        value: common.timeFormatSeconds(b.merchantReqTime)
    });
    a.push({
        key: "通知",
        value: b.callbackSuccess ? '成功' : '失败'
    });
    a.push({
        key: "通知次数",
        value: b.callbackLimit
    });
    return common.buildTabPanel(a)
}
function buildChannel(b) {
    var a = [];
    a.push({
        key: "上游渠道",
        value: b.channelDesc
    });
    a.push({
        key: "上游商户号",
        value: b.channelMerchantNo
    });
    a.push({
        key: "向上游推送时间",
        // value: common.toDateStr("yyyy-MM-dd HH:mm:ss", b.pushChannelTime)
        value: b.pushChannelTime
    });
    a.push({
        key: "上游订单号",
        value: b.channelOrderNo
    });
    a.push({
        key: "上游处理时间",
        // value: common.toDateStr("yyyy-MM-dd HH:mm:ss", b.channelNoticeTime)
        value: b.channelNoticeTime
    });
    a.push({
        key: "上游手续费",
        value: common.fixAmount(b.channelServiceCharge)
    }); 
    a.push({
        key: "处理标识",
        value: b.processTypeDesc
    });
    return common.buildTabPanel(a)
}
function showPerfectModal() {
    $("#txtChannelOrderNo,#txtChannelNoticeTime").val("");
    $("#txtChannelNoticeTime").val(common.toDateStr("yyyy-MM-dd HH:mm:ss", {
        time: new Date().getTime()
    }));
    $("#perfectModal").modal()
}
/* function submitPerfect() {
    if ($("#txtChannelOrderNo").val() == "") {
        myAlert.warning($("#txtChannelOrderNo").attr("placeholder"));
        return
    }
    if ($("#txtChannelNoticeTime").val() == "") {
        myAlert.warning($("#txtChannelNoticeTime").attr("placeholder"));
        return
    }
    common.submit(apiPath + "perfect?orderId=" + common.getQuery("orderId")+'&type=in', "perfectModal", function() {
        location.href = location.href
    })
} */


function submitPerfect() {
    if ($("#txtChannelOrderNo").val() == "") {
        myAlert.warning($("#txtChannelOrderNo").attr("placeholder"));
        return
    }
    if ($("#txtChannelNoticeTime").val() == "") {
        myAlert.warning($("#txtChannelNoticeTime").attr("placeholder"));
        return
    }
    /* common.submit(apiPath + "perfect?orderId=" + common.getQuery("orderId")+'&type=in', "perfectModal", function() {
        location.href = location.href
    }) */


    // var a = $("#btnFile")[0].files;
    // if (a.length == 0) {
    //     myAlert.warning("请选择要上传的截图");
    //     return
    // }
    //
    // if(a[0]['size'] > 2097152){
    //     myAlert.warning("上传图片不能大于2M");
    //     return
    // }
    //
    // if(a[0]['type'] != 'image/png' && a[0]['type'] != 'image/jpeg' && a[0]['type'] != 'image/jpg'){
    //     myAlert.warning("只能上传图片");
    //     return
    // }

    /* console.log(a[0]);
    return; */
    var b = new FormData();
    // b.append("file", a[0]);
    b.append("orderId", common.getQuery("orderId"));
    b.append("type", 'in');
    b.append("channelOrderNo", $("#txtChannelOrderNo").val());
    b.append("channelNoticeTime", $("#txtChannelNoticeTime").val());
    b.append("desc", $("#desc").val());
    // b.append("file_base64",$('#file_base64').val());
    desc = $("#desc").val()
    channelNoticeTime = $("#txtChannelNoticeTime").val()
    channelOrderNo = $("#txtChannelOrderNo").val()
    orderId = common.getQuery("orderId")
    $.ajax({
        url:apiPath + "confirm",
        dataType:'json',
        type:"post",
        data: {"orderId":orderId,"type":"in","channelOrderNo":channelOrderNo,"channelNoticeTime":channelNoticeTime,"desc":desc},
        async:false,
        success:function (d) {
            if (d) {
                if (d.success == -1) {
                    location.href = "/logout";
                    return
                }else if(d.success == 1){
                    myAlert.success(d.result)
                    $("#perfectModal").modal("hide");
                    $("#btnSearch").click()
                    window.location.reload();
                }else if(d.success == 0){
                    myAlert.error(d.result)
                }

            } else {
                myAlert.error(d.result.length > 0 ? d.result : "操作异常")
            }
        }
    });

}


function getColumns() {
    return [{
        field: "channelOrderNo",
        title: "上游订单号",
        align: "center"
    }, {
        field: "channelNoticeTime",
        title: "上游交易时间",
        align: "center",
        // formatter: function(b, c, a) {
        //     return common.toDateStr("yyyy-MM-dd HH:mm:ss", b)
        // }
    }, {
        field: "created_at",
        title: "提审时间",
        align: "center",
        // formatter: function(b, c, a) {
        //     return common.toDateStr("yyyy-MM-dd HH:mm:ss", b)
        // }
    }, {
        field: "check_time",
        title: "审核时间",
        align: "center",
        // formatter: function(b, c, a) {
        //     return common.toDateStr("yyyy-MM-dd HH:mm:ss", b)
        // }
    }, {
        field: "status",
        title: "审核状态",
        align: "center",
        formatter: function(b, c, a) {
            /* console.log(b); */
            if(b == '待审核'){
                $('#btnPerfect').attr('style','display:none');
            }
            return b;
        }
    }, {
        field: "commiter_desc",
        title: "提审备注",
        align: "center"
    }, {
        field: "desc",
        title: "审核备注",
        align: "center"
    }]
}

;