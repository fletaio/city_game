if (!window.numberWithCommas) {
    window.numberWithCommas = function (x) {
        return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
    }
}

function message(msg) {
    if (IsError(msg)) {
        var m = "error : " + msg.Message
    } else {
        var m = "message : " + msg
    }
    // console.log(m)
}

function toShotUnit (num) {
    if (typeof num === "string") {
        num = num.replace(/,/g, "")
    }
    num = parseInt(num)
    if (isNaN(num)) {
        throw "isNaN"
    }
    var unit = ""
    if (num/10000 < 1) {
        return num
    }
    unit = "k"
    num = parseInt(num/1000)
    if (num/1000 < 1) {
        return num+unit
    }

    unit = "m"
    num = parseInt(num/1000)
    if (num/1000 < 1) {
        return num+unit
    }

    unit = "g"
    num = parseInt(num/1000)
    return num+unit
}

function printInfo(x, y) {
    var $l = $("#info");
    var tile = gGame.tiles[x + y * gConfig.Size]
    $l.html("x : " + x + " y : "+y + " lv : " + tile.level + " type : " + tile.TypeName() + ((tile.BuildProcessing == true)?" construction":""))
}

function printLog(msg) {
    var $l = $("#log");
    $l.append($("<p>").html(msg))
    $l.scrollTop($l[0].scrollHeight)
}

function getNum (x, y) {
    return (parseInt(Math.log2((x+1)*73)*100 + Math.log10((y+1)*4321)*100)%10+1);
}

function getXYFromIndex(i) {
    if (i>=0 && i<=gConfig.Size*gConfig.Size) {
        return {x : i%gConfig.Size, y : parseInt(i/gConfig.Size)}
    }
    throw "getXYFromIndex i is out of index"
}

function directByNum(o, num) {
    switch (num) {
        case 0:
            o.x--
            break;
        case 1:
            o.y--
            break;
        case 2:
            o.x++
            break;
        case 3:
            o.y++
            break;
    }
    message("o.x " + o.x + " o.y " + o.y)
}

function ViewChanger() {
    var $btn = $("#hideBuilding")

    if ($btn.hasClass("hideBuilding")) {
        $btn.removeClass("hideBuilding")
        $("#touchpad").removeClass("hideBuilding")
        $("#screen").removeClass("hideBuilding")
    } else {
        $btn.addClass("hideBuilding")
        $("#touchpad").addClass("hideBuilding")
        $("#screen").addClass("hideBuilding")
    }
}

function logClean () {
    $("#log").html("")
}

function lockUpValueRange (v, n1, n2) {
    var min = (n1>n2)?n2:n1;
    var max = (n1<=n2)?n2:n1;

    if (min > v) {
        v = min
    }
    if (max < v) {
        v = max
    }
    return v
}

var Frequency = 0.8
var interval = 900
var speed = 30

$(function () {
    var $starField = $("#starField")
    var $body = $("body")

    var getStar = function (right) {
        if (typeof right === "undefined") {
            right = -100
        }
        var i = Math.floor(Math.random() * (14 - 1)) + 1;
        if(i > 5) {
            i = i % 4 + 2;
        }
        var top =  Math.floor(Math.random() * ($body.height() - 1)) + 1;
        return "<img src='/images/background/stars_"+i+".png' style='top:"+top+"px;right:"+right+"px;' />"
    }
    var make = function () {
        if (Math.random() < Frequency) {
            $starField.prepend($(getStar()))
        }
        sendLeft($starField.find("img"), $body.width())
    }
    var h = [];
    var k = 0;
    var bH = $body.height();
    var bW = $body.width();
    for (var j = 0 ; j < (speed*bH*bW)/1000000 ; j++) {
        var right = Math.floor(Math.random() * (bW - 1)) - 100;
        h[k++] = getStar(right)
    }
    $starField.html(h.join())
    setInterval(make, interval)
})

function sendLeft ($eles, maxWidth) {
    var deleteList = []
    for (var i = 0 ; i < $eles.length ; i++ ) {
        var t = $eles.eq(i);
        var left = parseInt(t.css("right"))
        t.css("right", (left+(speed*(interval/1000)))+"px")
        if (left > maxWidth) {
            deleteList.push(t)
        }
    }

    for (var i = 0 ; i < deleteList.length ; i++ ) {
        deleteList[i].remove()
    }
}


function calcDistance (start, end) {
    return Math.pow(Math.pow(end.x, 2) + Math.pow(end.y, 2), 0.5) - Math.pow(Math.pow(start.x, 2) + Math.pow(start.y, 2), 0.5)
}


function secondToDate(time) {
	time = parseInt(time)
	var ss = time%60
	time = parseInt((time)/60)
	var mm = time%60
	var hh = parseInt(time/60)
	var r = ""
	if (hh > 0) {
		r += hh+"h"
		if (mm != 0) {
			r += mm+"m"
		}
	} else if (mm > 0) {
		r += mm+"m"
	} else if (ss >= 0) {
		r += ss+"s"
	} else {
        return "pending"
    }
	// r += ("0"+mm).substr(-2)+"m"
	return r
}
