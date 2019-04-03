function Language(){}

Language.prototype.add = function (key, msg, location) {
    if (typeof msg === "undefined") {
        msg = key
    }
    if (typeof location === "undefined") {
        this[key] = msg
    }

    if (typeof this["location set " + location] === "undefined") {
        this["location set " + location] = {}
    }
    this["location set " + location] = msg
}

var language = new Language()

language.add("not enough balance", "Insufficient Balance")
language.add("not enough people", "Insufficient Manpower")
language.add("not enough power", "Insufficient power")
language.add("too fast", "too fast")
language.add("BuildProcessing not finished", "The construction was not completed.")
language.add("not enough lv5 building", "not enough lv5 buildings")
language.add("under construction", "It is not possible to build on a tile under construction.")
language.add("Available after agreeing", "First Agree to Terms and Conditions and Privace Policy")
language.add("Duplicate id", "Duplicate entry. Need another ID")
language.add("Failed to execute demolation command", "Failed to execute the demolation command")
language.add("Failed to execute upgrade command", "Failed to execute the upgrade command")
language.add("Failed to execute build command", "Failed to execute the build command")
language.add("expired account", "Your account has expired.")
language.add("commit error", "Commit error")
language.add("load fail", "Failed to load")
language.add("check id", "Please check your ID")
language.add("check pw", "Please check your PASSWORD")
language.add("check ethAddr", "Please check your ethereum wallet address")

language.add("Address Issue Success : ")
language.add(", go to login")
language.add("login Success", "Login Success!")
language.add("Invalid Id or Password", "Invalid ID or PASSWORD")

language.add("duplicated connection", "Duplicated connection.")
language.add("restart page", "Failed to communicate with server, page reload.")
