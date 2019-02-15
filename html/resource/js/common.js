function message(msg) {
    if (IsError(msg)) {
        var m = "error : " + msg.Message
    } else {
        var m = "message : " + msg
    }
    // console.log(m)

}

function printInfo(x, y) {
    var $l = $("#info");
    var tile = gGame.tiles[x + y * gConfig.Size]
    tile.UpdateInfo()

    if (tile.obj.BuildProcessing == true) {
        $l.html("x : " + x + " y : "+y + " lv" + (tile.obj.level+1) + " " + tile.TypeName() + " construction ")
    } else {
        $l.html("x : " + x + " y : "+y + " lv : " + tile.obj.level + " type : " + tile.TypeName())
    }
}

function printLog(msg) {
    var $l = $("#log");
    $l.append($("<p>").html(msg))
    $l.scrollTop($l[0].scrollHeight)

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

function getNum (x, y) {
    return (parseInt(Math.log2((x+1)*73)*100 + Math.log10((y+1)*4321)*100)%10+1);
}

function getXYFromIndex(i) {
    if (i>=0 && i<=gConfig.Size*gConfig.Size) {
        return {x : i%gConfig.Size, y : parseInt(i/gConfig.Size)}
    }
    throw "getXYFromIndex i is out of index"
}

document.addEventListener('keydown', function(event) {
    console.log(event.keyCode)
    if (event.keyCode >= 37 && event.keyCode <= 40) {// arrow
        direction = event.keyCode-37

        var t = $("#menu")[0].target
        if (typeof t === "undefined") {
            t = gGame.tiles[0]
        }
        var o = {x:t.x,y:t.y}
        directByNum(o, direction)
        menuClose()
        if (gGame.tiles[o.x+o.y*gConfig.Size]) {
            menuOpen(gGame.tiles[o.x+o.y*gConfig.Size].Hover())
        }
    }
    switch (event.keyCode) {
        case 27: //esc
            menuClose()
            break;
        case 73: //i 
            $("button#Industrial").click()
            break;
        case 82: //r
            $("button#Residential").click()
            break;
        case 67: //c
            $("button#Commercial").click()
            break;
        case 68: //d
            $("button#Demolition").click()
        case 85: //u
            $("button#Upgrade").click()
            break;
        case 72: //h
            $("button#hideBuilding").click()
            break;
    
        default:
            break;
    }
});

var hideBuilding = "Hide Building"
var viewBuilding = "View Building"
function ViewChanger() {
    var $btn = $("#hideBuilding")

    if ($btn.html() == hideBuilding) {
        $btn.html(viewBuilding)
        $("#touchpad").addClass("hideBuilding")
        $("#screen").addClass("hideBuilding")
    } else {
        $btn.html(hideBuilding)
        $("#touchpad").removeClass("hideBuilding")
        $("#screen").removeClass("hideBuilding")
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
    var $startField = $("#startField")
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
            $startField.prepend($(getStar()))
        }
        sendLeft($startField.find("img"), $body.width())
    }
    var h = [];
    var k = 0;
    var bH = $body.height();
    var bW = $body.width();
    for (var j = 0 ; j < (speed*bH*bW)/1000000 ; j++) {
        var right = Math.floor(Math.random() * (bW - 1)) - 100;
        h[k++] = getStar(right)
    }
    $startField.html(h.join())
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
    return (Math.pow(Math.pow(end.x, 2) + Math.pow(end.y, 2), 0.5) - Math.pow(Math.pow(start.x, 2) + Math.pow(start.y, 2), 0.5))/10
}