function initStep () {
    nextStep ("init")
}

function nextStep (step) {
    $("[step]").hide()
    $("[step='"+step+"']").show()
}

function join () {
    $("#ethAddr").val()
    $("#joinId").val()
    $("#joinPw").val()
    alert("join")
}

function login () {
    $("#loginId").val()
    $("#loginPw").val()
    alert("login")
    loginSuccess()
}

function loginSuccess() {
    $("#login").hide()
    $("#game").show()
}