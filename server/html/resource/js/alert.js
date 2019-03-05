function Alert (msg, func) {
	var a = $("#AlertBackground")
	a.show()
	a.find(".alertText").html(msg)
	Alert.func = func
}

function AlertClose(){
	var a = $("#AlertBackground")
	a.hide()
	a.find(".alertText").html("")
	if (typeof Alert.func === "function") {
		Alert.func()
	}
	Alert.func = undefined
}

var UIAlert = {}

UIAlert.Alert = function(btn, okFunc, cancelFunc) {
	UIAlert.btn = $("#"+btn);
	UIAlert.okFunc = okFunc;
	UIAlert.cancelFunc = cancelFunc;
	UIAlert.show()
}

UIAlert.show = function () {
	var $touch = $("#menu").parent()
	if (typeof UIAlert.alertUI === "undefined") {
		UIAlert.alertUI = $("#alertUI")
	}
	UIAlert.alertUI.attr("class", UIAlert.btn.attr("id"))
	$touch.append(UIAlert.alertUI)
	UIAlert.alertUI.show()
}
UIAlert.hide = function () {
	UIAlert.okFunc = null;
	UIAlert.cancelFunc = null;
	if (typeof UIAlert.alertUI === "undefined") {
		UIAlert.alertUI = $("#alertUI")
	}
	UIAlert.alertUI.hide()
}

UIAlert.okOnclick = function () {
	if (UIAlert.okFunc) {
		UIAlert.okFunc()
	}
	UIAlert.hide()
	menuClose()
};

UIAlert.cancelOnclick = function () {
	if (typeof UIAlert.cancelFunc === "function") {
		UIAlert.cancelFunc()
	}
	UIAlert.hide()
};
