var DashboardTransactionsChart={
    chart : new lineChart(),
    reload : function() {
        DashboardTransactionsChart.chart.draw()
    },
    init : function (recursive) {
        DashboardTransactionsChart.chart.target = "fleta-Transactions";
        DashboardTransactionsChart.chart.color = "#e31062";
        DashboardTransactionsChart.chart.dataUrl = "/data/transactions.data";
        DashboardTransactionsChart.chart.tooltipPrefix = "Txs : ";
        // DashboardTransactionsChart.chart.m = {top: 50, right: 20, bottom: 20, left: 50};
        // DashboardTransactionsChart.chart.height = 280;
        DashboardTransactionsChart.reload();
        if (recursive !== false) {
            setInterval( function () {
                DashboardTransactionsChart.reload();
            }, 3000 );
        }
    }
}
