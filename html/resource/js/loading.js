$(window).on("load", function() {
    setTimeout(function () {
        $(".loading").css("display", "none")
        $(".after-loading-content").css("opacity", "initial")
        initStep ()
    }, 500)
});
