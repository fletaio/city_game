function message(msg) {
    if (IsError(msg)) {
        var m = "error : " + msg.Message
    } else {
        var m = "message : " + msg
    }
    console.log(m)
    printLog(m)

}

function printInfo(x, y) {
    var $l = $("#info");
    var tile = Tiles[x + y * gConfig.Size]
    tile.UpdateInfo()
    
    $l.html("x : " + x + " y : "+y + " lv : " + tile.obj.level + " type : " + tile.Type)
}

function printLog(msg) {
    var $l = $("#log");
    $l.append($("<p>").html(msg))
    $l.scrollTop($l[0].scrollHeight)

}

function directByNum(o, num) {
    switch (num) {
        case 0:
            if (o.x > 0) {
                o.x--
            }
            break;
        case 1:
            if (o.y > 0) {
                o.y--
            }
            break;
        case 2:
            if (o.x < gConfig.Size-1) {
                o.x++
            }
            break;
        case 3:
            if (o.y < gConfig.Size-1) {
                o.y++
            }
            break;
    }
    message("o.x " + o.x + " o.y " + o.y)
}

function getNum (x, y) {
    return (parseInt(Math.log2((x+1)*73)*100 + Math.log10((y+1)*4321)*100)%10+1);
}

function deleteMenu() {
    $("#menu").html("")
}

function addMenu(funcs) {
    //<button id="btn1" onclick="$('#menu')[0].target.Demolition()" value="DEMOLITION">Demolition</button>
    for (var key in funcs) {
        var btn = $("<button id=\""+funcs[key]+"\" onclick=\"$('#menu')[0].target.RunCommand('"+funcs[key]+"')\" value=\""+key+"\">"+key+"</button>")
        $("#menu").append(btn)
    }
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
            t = Tiles[0]
        }
        var o = {x:t.x,y:t.y}
        directByNum(o, direction)
        menuClose()
        Tiles[o.x+o.y*gConfig.Size].Hover().Menu()
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