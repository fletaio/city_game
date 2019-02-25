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

language.add("not enough balance", "not enough balance")
language.add("not enough people", "not enough people")
language.add("not enough power", "not enough power")
language.add("too fast", "too fast")
language.add("BuildProcessing not finished", "BuildProcessing not finished")
language.add("not enough lv5 building", "not enough lv5 building")
language.add("under construction", "It is not possible to build on a tile under construction.")
language.add("Available after agreeing", "Available after agreeing to the Terms and Conditions and Privacy Policy")
language.add("Duplicate id or ether addr", "Duplicate id or ether addr")
language.add("Failed to execute demolation command")
language.add("Failed to execute upgrade command")
language.add("Failed to execute build command")
language.add("commit error")
language.add("load fail")
language.add("check id")
language.add("check pw")
language.add("check ethAddr")

language.add("Address Issue Success : ")
language.add(", go to login")
language.add("login Success")
language.add("Invalid Id or Password")

language.add("duplicated connection")

