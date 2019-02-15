function Language(){}

Language.prototype.add = function (key, msg, location) {
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
