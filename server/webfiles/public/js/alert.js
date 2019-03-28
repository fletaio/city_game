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
	if (typeof UIAlert.alertUI === "undefined") {
		UIAlert.alertUI = $("#alertUI")
	}
	if (btn == "Upgrade" || btn == "Demolition") {
		$("#upgrade_menu").append(UIAlert.alertUI)
	} else if (btn == "Industrial" || btn == "Commercial" || btn == "Residential") {
		$("#build_menu").append(UIAlert.alertUI)
	} else {
		return
	}
	var $touch = $("#menu").parent()
	UIAlert.alertUI.attr("class", btn)
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
};

UIAlert.cancelOnclick = function () {
	if (UIAlert.cancelFunc) {
		UIAlert.cancelFunc()
	}
	UIAlert.hide()
};
