var ERROR_CODE = 1
function Error (message) {
    this.Type = ERROR_CODE++;
    this.Message = message
    this.Symbol
}
function IsError(obj) {
    return obj.Symbol == Error.Symbol
}

Error.Symbol = Symbol("Error")
Error.prototype.Symbol = Error.Symbol
Error.prototype.Type = 0
Error.prototype.Message = "Error"

var NotEnoughBuildingLv = new Error("NotEnoughBuildingLv")