var loginInfo;
function LoginInfo(key, addr, utxos) {
    this.Key = key;
    this.Addr = addr;
	SendQueue.utxo = utxos;
}

LoginInfo.Simbol = Symbol("LoginInfo");
LoginInfo.prototype.Simbol = LoginInfo.Simbol;

// LoginInfo.prototype.popUTXO = function() {
// 	var utxo = this.utxos[0];
// 	this.utxos.splice(0, 1);
// 	return utxo;
// }

// LoginInfo.prototype.pushUTXO = function(utxo) {
// 	this.utxos.push(utxo);
// }

function initStep () {
    nextStep ("init")
}

function nextStep (step) {
    $("[step]").hide()
    $("[step='"+step+"']").show()
    $("[step='"+step+"'].focus, [step='"+step+"'] .focus").focus()
    // $("[step='"+step+"']").find("input").val("").removeAttr("checked")
    $("[step='"+step+"']").removeAttr("checked")
}

function validate (str) {
    if (str.length < 4) {
        return false;
    }
}

var joinFlag = false
function join () {
    if (joinFlag == false) {
        joinFlag = true
    } else {
        return
    }
    var ethAddr = $("#ethAddr").val()
    var userid = $("#joinId").val()
    var userpw = $("#joinPw").val()

    if (validate(userid)) {
        Alert(language["check id"])
        joinFlag = false
        return
    }
    if (validate(userpw)) {
        Alert(language["check pw"])
        joinFlag = false
        return
    }
    if (validate(ethAddr)) {
        Alert(language["check ethAddr"])
        joinFlag = false
        return
    }

    var key = getPubKey(userid, userpw)
    var pk = key.getPublic().encodeCompressed("hex")

    $.ajax({
        type: "POST",
        dataType : "json",
        url : "/api/accounts",
        data : JSON.stringify({
            "public_key": pk,
            "user_id": userid,
            "reward": ethAddr
        }),
        success : function (d) {
            if (typeof d === "string") {
                d = JSON.parse(d)
            }
            Alert(language["Address Issue Success : "]+ d.address + language[", go to login"])
            nextStep("login")
        },
        error: function(d) {
            joinFlag = false
            Alert(language["Duplicate id or ether addr"])
        }
    })
}

var loginFlag = false
function login () {
    if (loginFlag == false) {
        loginFlag = true
    } else {
        return
    }
    var userid = $("#loginId").val()
    var userpw = $("#loginPw").val()

    if (validate(userid)) {
        Alert(language["check id"])
        loginFlag = false
        return
    }
    if (validate(userpw)) {
        Alert(language["check pw"])
        loginFlag = false
        return
    }

    var key = getPubKey(userid, userpw)
    var pk = key.getPublic().encodeCompressed("hex")

    $.ajax({
        type: "GET",
        url : "/api/accounts?pubkey="+pk,
        success : function (d) {
            if (typeof d === "string") {
                d = JSON.parse(d)
            }
            loginInfo = new LoginInfo(key, d.address, d.utxos)
            Alert(language["login Success"])
            loginSuccess()
        },
        error: function(d) {
            loginFlag = false
			switch(d.responseText) {
			case "not exist account":
                Alert(language["Invalid Id or Password"])
				break;
			default:
				Alert(d.responseText);
			}
        }
    })

}

function loginSuccess() {
    $("#login").hide()
    $("#game").show()
    initGame()
}

function getPubKey (userid, userpw) {
    userid = SHA2(userid);
	userpw = SHA2(userpw);
	var salt = SHA2("this is fleta city game");
	var keyHex = SHA2(userid+"#"+userpw+"#"+salt);
	var key = ec.keyPair({
		priv: keyHex,
		privEnc: "hex",
    });
    return key
}
