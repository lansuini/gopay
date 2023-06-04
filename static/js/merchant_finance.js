$(function() {
    common.getAjax(apiPath + "basedata?requireItems=financeType,operateSource", function(a) {
        $("#selOperateSource").initSelect(a.result.operateSource, "key", "value", "操作来源");
        $("#selFinanceType").initSelect(a.result.financeType, "key", "value", "收支类型");
        $("#btnSearch").initSearch(apiPath + "financesearch", getColumns())
        $("#btnExport").bind('click',function () {
            $("#btnExport").initExport(apiPath + "financesearch", getColumns(), {})
        })
    });
    common.initSection()
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
        field: "sourceDesc",
        title: "订单类型",
        align: "center"
    }, {
        field: "merchantOrderNo",
        title: "商户订单号",
        align: "center"
    }, {
        field: "platformOrderNo",
        title: "平台订单号",
        align: "center"
    }, {
        field: "accountDate",
        title: "账务日期",
        align: "center"
    }, {
        field: "financeTypeDesc",
        title: "收支类型",
        align: "center"
    }, {
        field: "operateSource",
        title: "操作来源",
        align: "center"
    }, {
        field: "amount",
        title: "交易金额",
        align: "center",
        formatter: function(b, c, a) {
            return common.fixAmount(b)
        }
    }, {
        field: "balance",
        title: "余额",
        align: "center",
        formatter: function(b, c, a) {
            return common.fixAmount(b)
        }
    }, {
        field: "summary",
        title: "交易摘要",
        align: "center"
    }, {
        field: "insTime",
        title: "交易日期",
        align: "center"
    }]
}
;