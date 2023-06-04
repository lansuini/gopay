var channels
$(function() {
    common.getAjax(apiPath + "basedata?requireItems=channel", function(a) {
        apiPath += "merchant/settlementchannel/";
        channels = a.result.channel
        $("#selChannel").initSelect(a.result.channel, "key", "value", "代付渠道");
        $("#btnSearch").initSearch(apiPath + "search", getColumns())
    });
    $("#tabExport").initExportTable(getExportColumns(), exportTable);
    $("#btnImport").click(function() {
        showFileModal()
    });
    $("#btnSubmit").click(uploadFile)
    $("#batchUdate").click(submit)
});
function getColumns() {
    return [
   /* {
        "checkbox":true
    },*/
    /*{
        field: "setId",
        title: "ID",
        align: "center",
    },*/
    {
        field: "-",
        title: "#",
        align: "center",
        formatter: function(b, c, a) {
            return a + 1
        }
    }, {
        field: "channelDesc",
        title: "代付渠道",
        align: "center",
        formatter: function (b,c,a){
            channels.forEach(function(item){
                if(item.key === c.channel){
                    b = item.value
                    return item.value
                }
            });
            return c.channel
        }
    }, {
        field: "channelMerchantNo",
        title: "渠道商户号",
        align: "center"
    }, {
        field: "merchantNo",
        title: "商户号",
        align: "center"
    },{
            field: "status",
            title: "配置状态",
            align: "center"
    },
        /*{
            field: "settlementChannelStatus",
            title: "上游通道状态",
            align: "center"
        },*/
        /*{
        field: "shortName",
        title: "商户简称",
        align: "center"
    }, */
    //     {
    //     field: "settlementAccountTypeDesc",
    //     title: "代付账户",
    //     align: "center"
    // },
        /*{
        field: "accountBalance",
        title: "账户余额",
        align: "center",
        formatter: function(b, c, a) {
            return common.fixAmount(b)
        }
    }, */
        {
        field: "-",
        title: "操作",
        align: "center",
        formatter: function(b, c, a) {
            return "<a onclick='exportTableInvoke(\"" + c.merchantNo + "\")'>下载配置</a><a onclick='showFileModal(\"" + c.merchantNo + "\")'>更改配置</a><a onclick='deleFileModel(\"" + c.setId + "\")'>删除</a>"
        }
    }]
}
function getExportColumns() {
    return [{
        field: "merchantNo",
        // title: "下游商户号"
        title: "merchantNo"
    }, {
        field: "channel",
        // title: "渠道名称"
        title: "channel"
    }, {
        field: "channelMerchantNo",
        // title: "渠道商户号"
        title: "channelMerchantNo"
    }, {
        field: "settlementChannelStatus",
        // title: "代付渠道状态"
        title: "settlementChannelStatus"
    }, {
        field: "openOneAmountLimit",
        // title: "是否开启单笔金额控制",
        title: "openOneAmountLimit",
        formatter: function(b, c, a) {
            return b ? 1 : 0
        }
    }, {
        field: "oneMinAmount",
        // title: "单笔最小金额",
        title: "oneMinAmount",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(2)
        }
    }, {
        field: "oneMaxAmount",
        // title: "单笔最大金额",
        title: "oneMaxAmount",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(2)
        }
    }, {
        field: "openDayAmountLimit",
        // title: "是否开启单日累计金额控制",
        title: "openDayAmountLimit",
        formatter: function(b, c, a) {
            return b ? 1 : 0
        }
    }, {
        field: "dayAmountLimit",
        // title: "单日累计金额上限"
        title: "dayAmountLimit"
    }, {
        field: "openDayNumLimit",
        // title: "是否开启单日累计笔数控制",
        title: "openDayNumLimit",
        formatter: function(b, c, a) {
            return b ? 1 : 0
        }
    }, {
        field: "dayNumLimit",
        // title: "单日累计笔数上限"
        title: "dayNumLimit"
    }, {
        field: "openTimeLimit",
        // title: "是否开启交易时间控制",
        title: "openTimeLimit",
        formatter: function(b, c, a) {
            return b ? 1 : 0
        }
    }, {
        field: "beginTime",
        // title: "开始时间"
        title: "beginTime"
    }, {
        field: "endTime",
        // title: "结束时间"
        title: "endTime"
    }, {
        field: "status",
        // title: "配置状态"
        title: "status"
    }]
}
function exportTableInvoke(a) {
    $("#tabExport").bootstrapTable("refresh", {
        url: apiPath + "export?merchantNo=" + a
    })
}
function exportTable(c) {
    var b = $("#tabExport");
    var a = $("#divExport");
    if (b.find(">tbody >tr.no-records-found").length > 0) {
        myAlert.warning("没有记录可以导出");
        return
    }
    if (a) {
        a.show();
        b.tableExport({
            type: "csv",
            csvUseBOM: false,
            fileName: "商户代付渠道配置" + b.find(">tbody >tr:first >td:first").html()
        });
        a.hide()
    }
}

function deleFileModel(a) {
    myConfirm.show({
        title: "您确定要删除当前配置？",
        sure_callback: function() {
            common.getAjax(apiPath + "del?setId=" + a, function(e) {
                if (e && e.success) {
                    myAlert.success(e.result, undefined, function() {
                        location.href = location.href
                    })
                }else {
                    myAlert.error(e.result, undefined, function () {
                        location.href = location.href
                    })
                }
            })
        }
    })
}

function showFileModal(a) {
    if (a != undefined) {
        $("#txtMerchantNo").val(a).attr("disabled", "disabled")
    } else {
        $("#txtMerchantNo").val("").removeAttr("disabled")
    }
    $("#btnFile").val("");
    $("#fileModal").modal()
}
function uploadFile() {
    if ($("#txtMerchantNo").val() == "") {
        myAlert.warning("请输入商户号");
        return
    }
    var a = $("#btnFile")[0].files;
    if (a.length == 0) {
        myAlert.warning("请选择文件");
        return
    }
    var b = new FormData();
    b.append("file", a[0]);
    b.append("merchantNo", $("#txtMerchantNo").val());
    common.uploadFile(apiPath + "import", b, function(c) {
        if (c.success == 1) {
            myAlert.success("操作成功");
            $("#fileModal").modal("hide");
            $("#btnSearch").click()
        } else {
            myAlert.error(c.result.length > 0 ? c.result : "操作失败")
        }
    })
}

function showEditModal() {
    $("#batchFile").val("");
    var rows = $("#tabMain").bootstrapTable('getSelections');
    if(rows.length == 0){
        myAlert.warning("请选择要修改的配置");
        return false;
    }
    // console.log(rows);
    var ids = new Array();
    $(rows).each(function() {
        ids.push(this.setId);
    });
    $("#setIds").val(ids)
    console.log(ids);
    $("#editModal").modal()


}

function submit() {
    var rows = $("#tabMain").bootstrapTable('getSelections');
    if(rows.length == 0){
        myAlert.warning("请选择要修改的配置");
        return false;
    }
    var a = $("#batchFile")[0].files;
    if (a.length == 0) {
        myAlert.warning("请上传文件");
        return
    }
    var ids = new Array();
    $(rows).each(function() {
        ids.push(this.setId);
    });
    var b = new FormData();
    b.append("file", a[0]);
    b.append("setIds", ids);
    common.uploadFile(apiPath + "batchUpdate", b, function(c) {
        if (c.success == 1) {
            myAlert.success("批量更改成功");
            $("#editModal").modal("hide");
            $("#btnSearch").click()
        } else {
            myAlert.error(c.result.length > 0 ? c.result : "批量更改失败")
        }
    })
}

;