function FletaCityCoin (data) {
    this.x = +data.x;
    this.y = +data.y;
    this.index = this.x + this.y*gConfig.Size;
    this.hash = data.hash;
    this.height = +data.height;
    this.coin_type = data.coin_type;
    this.PutOnMapFlag = false;
}

FletaCityCoin.prototype.PutOnMap = function () {
    if (this.PutOnMapFlag == false) {
        this.PutOnMapFlag = true
        var t = gGame.tiles[this.x+this.y*gConfig.Size].UI.touch
        var has = t.find("[hash='"+this.hash+"']").length
        if (!has)  {
            this.UI = $("#fletaCoinTemplet .fletaCoin").clone()
            t.append(this.UI)
            this.UI.attr("x", this.x)
            this.UI.attr("y", this.y)
            this.UI.attr("hash", this.hash)
            this.UI.attr("height", this.height)
            this.UI.attr("coin_type", this.coin_type)
            this.UI.attr("onclick", "event.stopPropagation();gGame.tiles["+(+this.x + +this.y*gConfig.Size)+"].RunCommand('GetCoin', '"+this.coin_type+":"+this.height+":"+this.hash+"')")
        }
    }
}

FletaCityCoin.prototype.HideOnMap = function () {
    this.UI.hide();
    (function (hash) {
        setTimeout(function () {
            if (gGame.coin_list[hash]) {
                gGame.coin_list[hash].Remove()
            }
        }, 3000)
    })(this.hash)
}

FletaCityCoin.prototype.ShowOnMap = function () {
    if (this.PutOnMapFlag == false) {
        this.PutOnMap()
    }
    this.UI.show()
}

FletaCityCoin.prototype.Remove = function () {
    this.UI.remove()
    delete gGame.coin_list[this.hash]
}
