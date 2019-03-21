if (!window.numberWithCommas) {
    window.numberWithCommas = function (x) {
        return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
    }
}

function nextStep (step) {
    $("[step]").hide()
    $("[step='"+step+"']").show()
    $("[step='"+step+"'].focus, [step='"+step+"'] .focus").focus()
    $("[step='"+step+"']").removeAttr("checked")
}

function message(msg) {
    if (IsError(msg)) {
        var m = "error : " + msg.Message
    } else {
        var m = "message : " + msg
    }
    // console.log(m)
}

function toShortUnit (num) {
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

function printInfo(tile) {
    var $l = $("#info");
    $l.show();
    $l.html("x : " + tile.x + " y : "+ tile.y + " lv : " + tile.level + " type : " + getAreaTypeName(tile.area_type) + ((tile.is_building == true)?" construction":""))
}
function hideInfo(tile) {
    $("#info").hide();
}

function printLog(msg) {
    var $l = $("#log");
    $l.append($("<p>").html(msg))
    $l.scrollTop($l[0].scrollHeight)
}

function getNum(x, y) {
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
        return "<img src='/public/images/background/stars_"+i+".png' style='top:"+top+"px;right:"+right+"px;' />"
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


function c (str) {
    var strs = str.split(" ")
    var t = (strs[0]-strs[1])/1000
    const input = document.createElement('input');
    input.style.position = 'fixed';
    input.style.opacity = 0;
    input.value = t;
    document.body.appendChild(input);
    input.select();
    document.execCommand('Copy');
    document.body.removeChild(input);
}

function getBuildError(resource, bd) {
	if(resource.balance < bd.cost_usage) {
		return language["not enough balance"];
	}
	if(resource.power_remained < bd.power_usage) {
		return language["not enough power"];
	}
	if(resource.man_remained < bd.man_usage) {
		return language["not enough people"];
	}
	return null;
}

var CommercialAreaType = 1;
var IndustrialAreaType = 2;
var ResidentialAreaType = 3;
var EndOfAreaType = 4;

function getAreaTypeName(area_type) {
	switch(area_type) {
	case CommercialAreaType:
		return "Commercial";
	case IndustrialAreaType:
		return "Industrial";
	case ResidentialAreaType:
		return "Residential";
	default:
		return "Unknown";
	}
}

var gGameExpDefine = [
	{"lv":1	, "exp": 0,		"acc_exp":0,	"class":"lv_bronze"		},
	{"lv":2	, "exp": 10,	"acc_exp":10,	"class":"lv_bronze"		},
	{"lv":3	, "exp": 15,	"acc_exp":25,	"class":"lv_bronze"		},
	{"lv":4	, "exp": 20,	"acc_exp":45,	"class":"lv_bronze"		},
	{"lv":5	, "exp": 25,	"acc_exp":70,	"class":"lv_bronze"		},
	{"lv":6	, "exp": 30,	"acc_exp":100,	"class":"lv_bronze"		},
	{"lv":7	, "exp": 50,	"acc_exp":150,	"class":"lv_bronze"		},
	{"lv":8	, "exp": 70,	"acc_exp":220,	"class":"lv_silver"		},
	{"lv":9	, "exp": 100,	"acc_exp":320,	"class":"lv_silver"		},
	{"lv":10, "exp": 140,	"acc_exp":460,	"class":"lv_silver"		},
	{"lv":11, "exp": 190,	"acc_exp":650,	"class":"lv_silver"		},
	{"lv":12, "exp": 250,	"acc_exp":900,	"class":"lv_silver"		},
	{"lv":13, "exp": 320,	"acc_exp":1220,	"class":"lv_silver"		},
	{"lv":14, "exp": 400,	"acc_exp":1620,	"class":"lv_silver"		},
	{"lv":15, "exp": 500,	"acc_exp":2120,	"class":"lv_gold"		},
	{"lv":16, "exp": 700,	"acc_exp":2820,	"class":"lv_gold"		},
	{"lv":17, "exp": 800,	"acc_exp":3620,	"class":"lv_gold"		},
	{"lv":18, "exp": 1000,	"acc_exp":4620,	"class":"lv_gold"		},
	{"lv":19, "exp": 1500,	"acc_exp":6120,	"class":"lv_gold"		},
	{"lv":20, "exp": 2000,	"acc_exp":8120,	"class":"lv_gold"		},
	{"lv":21, "exp": 2500,	"acc_exp":10620,"class":"lv_gold"		},
	{"lv":22, "exp": 3000,	"acc_exp":13620,"class":"lv_fleta"		},
	{"lv":23, "exp": 3500,	"acc_exp":17120,"class":"lv_fleta"		},
	{"lv":24, "exp": 4000,	"acc_exp":21120,"class":"lv_fleta"		},
	{"lv":25, "exp": 4500,	"acc_exp":25620,"class":"lv_fleta"		},
	{"lv":26, "exp": 5000,	"acc_exp":30620,"class":"lv_fleta"		},
	{"lv":27, "exp": 5500,	"acc_exp":36120,"class":"lv_fleta"		},
	{"lv":28, "exp": 6000,	"acc_exp":42120,"class":"lv_fleta"		},
	{"lv":29, "exp": 6500,	"acc_exp":48620,"class":"lv_fleta"		},
	{"lv":30, "exp": 7000,	"acc_exp":55620,"class":"lv_fleta"		},
	{"lv":31, "exp": 7500,	"acc_exp":63120,"class":"lv_fleta"		},
	{"lv":32, "exp": 8000,	"acc_exp":71120,"class":"lv_fleta"		},
	{"lv":33, "exp": 8500,	"acc_exp":79620,"class":"lv_fleta"		},
	{"lv":34, "exp": 9000,	"acc_exp":88620,"class":"lv_fleta"		},
	{"lv":35, "exp": 9500,	"acc_exp":98120,"class":"lv_fleta"		}
]

function expIndexOf(searchAccExp) {
    var minIndex = 0;
    var maxIndex = gGameExpDefine.length - 1;
    var currentIndex;
    var currentAccExp;

    while (minIndex <= maxIndex) {
        currentIndex = (minIndex + maxIndex) / 2 | 0;
        currentAccExp = gGameExpDefine[currentIndex].acc_exp;

        if (currentAccExp < searchAccExp) {
            minIndex = currentIndex + 1;
        }
        else if (currentAccExp > searchAccExp) {
            maxIndex = currentIndex - 1;
        }
        else {
            return gGameExpDefine[+currentIndex+1];
        }
    }

    if ( gGameExpDefine[currentIndex].acc_exp > searchAccExp ) {
        return gGameExpDefine[+currentIndex];
    } else {
        return gGameExpDefine[+currentIndex+1];
    }
}