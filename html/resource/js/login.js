var loginInfo;
function LoginInfo(key, addr, utxos) {
    this.Key = key;
    this.Addr = addr;
	this.utxos = utxos;
}

LoginInfo.Simbol = Symbol("LoginInfo");
LoginInfo.prototype.Simbol = LoginInfo.Simbol;

LoginInfo.prototype.popUTXO = function() {
	var utxo = this.utxos[0];
	this.utxos.splice(0, 1);
	return utxo;
}

LoginInfo.prototype.pushUTXO = function(utxo) {
	this.utxos.push(utxo);
}

function initStep () {
    nextStep ("init")
}

function nextStep (step) {
    $("[step]").hide()
    $("[step='"+step+"']").show()
    $("[step='"+step+"'].focus, [step='"+step+"'] .focus").focus()
    $("[step='"+step+"']").find("input").val("").removeAttr("checked")
}

function validate (str) {
    if (str.length < 4) {
        return false;
    }
}

function join () {
    var ethAddr = $("#ethAddr").val()
    var userid = $("#joinId").val()
    var userpw = $("#joinPw").val()

    if (validate(userid)) {
        alert("check id")
        return
    }
    if (validate(userpw)) {
        alert("check pw")
        return
    }
    if (validate(ethAddr)) {
        alert("check ethAddr")
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
            alert("Address Issue Success : "+ d.address + ", go to login")
            nextStep("login")
        },
        error: function(d) {
            alert("error")
        }
    })
}

function login () {
    var userid = $("#loginId").val()
    var userpw = $("#loginPw").val()

    if (validate(userid)) {
        alert("check id")
        return
    }
    if (validate(userpw)) {
        alert("check pw")
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
            alert("login Success")
            loginSuccess()
        },
        error: function(d) {
			switch(d.responseText) {
			case "not exist account":
				alert("Account or password in correct");
				break;
			default:
				alert(d.responseText);
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
