var productType,rateType,payType,commonStatus,bankCode
$(function() {
    common.getAjax(apiPath + "basedata?requireItems=productType,rateType,payType,commonStatus,bankCode", function(a) {
        productType = a.result.productType
        rateType = a.result.rateType
        payType = a.result.payType
        commonStatus = a.result.commonStatus
        bankCode = a.result.bankCode
        $("#selProductType").initSelect(a.result.productType, "key", "value", "产品类型");
        $("#selRateType").initSelect(a.result.rateType, "key", "value", "费率类型");
        $("#selPayType").initSelect(a.result.payType, "key", "value", "支付方式");
        $("#selStatus").initSelect(a.result.commonStatus, "key", "value", "状态");
        $("#btnSearch").initSearch(option.apiPath + "search", getColumns())
    });
    $("#tabExport").initExportTable(getExportColumns(), exportTable);
    $("#btnImport").click(function() {
        showFileModal()
    });
    $("#btnSubmit").click(uploadFile)
});
function getColumns() {
    return [{
        field: "-",
        title: "#",
        align: "center",
        formatter: function(b, c, a) {
            return a + 1
        }
    }, {
        field: "channelMerchantNo",
        title: "渠道商户号",
        align: "center"
    }, {
        field: "channel",
        title: "渠道商户名称",
        align: "center"
    }, {
        field: "productTypeDesc",
        title: "产品类型",
        align: "center",
        formatter: function(b,c,a) {
            productType.forEach(function(item){
                if(item.key === c.productType){
                    b = item.value
                    return item.value
                }
            });
            return b
        }
    }, {
        field: "payTypeDesc",
        title: "支付方式",
        align: "center",
        formatter: function(b,c,a) {
            payType.forEach(function(item){
                if(item.key === c.payType){
                    b = item.value
                    return item.value
                }
            });
            return b
        }
    }, {
        field: "bankCodeDesc",
        title: "银行",
        align: "center",
        formatter: function(b,c,a) {
            bankCode.forEach(function(item){
                if(item.key === c.bankCode){
                    b = item.value
                    return item.value
                }
            });
            return b
        }
    }, {
        field: "rateTypeDesc",
        title: "费率类型",
        align: "center",
        formatter: function(b,c,a) {
            rateType.forEach(function(item){
                if(item.key === c.rateType){
                    b = item.value
                    return item.value
                }
            });
            return b
        }
    }, {
        field: "rate",
        title: "费率值",
        align: "center",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(6)
        }
    }, {
        field: "fixed",
        title: "固定值",
        align: "center",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(2)
        }
    },{
        field: "minServiceCharge",
        title: "最小手续费",
        align: "center",
        formatter: function(b, c, a) {
            return common.fixAmount(b)
        }
    }, {
        field: "maxServiceCharge",
        title: "最大手续费",
        align: "center",
        formatter: function(b, c, a) {
            return common.fixAmount(b)
        }
    }, {
        field: "beginTime",
        title: "生效时间",
        align: "center",
        formatter: function(b, c, a) {
            if(b === "") return ""
            return common.timeFormatSeconds(b)
        }
    }, {
        field: "endTime",
        title: "失效时间",
        align: "center",
        formatter: function(b, c, a) {
            if(b === "") return ""
            return common.timeFormatSeconds(b)
        }
    }, {
        field: "statusDesc",
        title: "状态",
        align: "center",
        formatter: function(b,c,a) {
            commonStatus.forEach(function(item){
                if(item.key === c.status){
                    b = item.value
                    return item.value
                }
            });
            return b
        }
    }, {
        field: "-",
        title: "操作",
        align: "center",
        formatter: function(b, c, a) {
            return "<a onclick='exportTableInvoke(\"" + c.channelMerchantNo + "\")'>下载配置</a><a onclick='showFileModal(\"" + c.channelMerchantNo + "\")'>更改配置</a>"
        }
    }]
}

function getExportColumns() {
    return [{
        field: "channelMerchantNo",
        // title: "渠道商户号",
        title: "channelMerchantNo"
    }, {
        field: "productType",
        // title: "产品类型"
        title: "productType"
    }, {
        field: "payType",
        // title: "支付方式"
        title: "payType"
    }, {
        field: "bankCode",
        // title: "银行代码",
        title: "bankCode",
        formatter: function(b, c, a) {
            return b != null ? b : ""
        }
    }, {
        field: "cardType",
        // title: "卡种",
        title: "cardType",
        formatter: function(b, c, a) {
            return b != null ? b : ""
        }
    }, {
        field: "minAmount",
        // title: "最小金额",
        title: "minAmount",
            formatter: function(b, c, a) {
                b = parseFloat(b)
                return b.toFixed(2) ? b.toFixed(2) : ""
            }
    }, {
        field: "maxAmount",
        // title: "最大金额",
        title: "maxAmount",
            formatter: function(b, c, a) {
                b = parseFloat(b)
                return b.toFixed(2) ? b.toFixed(2) : ""
            }
    }, {
        field: "rateType",
        // title: "费率类型"
        title: "rateType"
    }, {
        field: "rate",
        // title: "费率值",
        title: "rate",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(6)
        }
    }, {
        field: "fixed",
        // title: "费率固定值",
        title: "fixed",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return b.toFixed(2)
        }
    }, {
        field: "minServiceCharge",
        // title: "最小手续费",
        title: "minServiceCharge",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return c.rateType == "Rate" ? b.toFixed(2) : ""
        }
    }, {
        field: "maxServiceCharge",
        // title: "最大手续费",
        title: "maxServiceCharge",
        formatter: function(b, c, a) {
            b = parseFloat(b)
            return c.rateType == "Rate" ? b.toFixed(2) : ""
        }
    }, {
        field: "beginTime",
        // title: "生效时间",
        title: "beginTime",
        formatter: function(b, c, a) {
            return common.timeFormatDate(b)
        }
    }, {
        field: "endTime",
        // title: "失效时间",
        title: "endTime",
        formatter: function(b, c, a) {
            return common.timeFormatDate(b)
        }
    }, {
        field: "status",
        // title: "状态"
        title: "status"
    }]
}
function exportTableInvoke(a) {
    $("#tabExport").bootstrapTable("refresh", {
        url: option.apiPath + "export?merchantNo=" + a
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
            // csvEnclosure: '',
            fileName: option.merchantType + "费率配置" + b.find(">tbody >tr:first >td:first").html()
        });
        a.hide()
    }
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
        myAlert.warning($("#txtMerchantNo").attr("placeholder"));
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
    // b.append("loginName", $("#txtLoginName").val());
    common.uploadFile(option.apiPath + "import", b, function(c) {
        if (c.success == 1) {
            myAlert.success("操作成功");
            $("#fileModal").modal("hide");
            $("#btnSearch").click()
        } else {
            myAlert.error(c.result.length > 0 ? c.result : "操作失败")
        }
    })
}
;