$(document).on('click', 'body', function(e) {
	if (!islandMoved) {
		var point = getPoint(e);
		var tile = getTileFromPoint(point)
		if (typeof tile !== "undefined") {
			var t = $("#menu")[0].target
			if (typeof t === "undefined" || t.index != tile.index ) {
				menuOpen(tile)
				return
			}
		}
		menuClose()
	}
});

$(document).on('mousemove', '#touchpad', function(e) {
	var point = getPoint(e);
	var tile = getTileFromPoint(point)
	if (typeof tile !== "undefined") {
		tile.Hover();
	}
});

var islandMove = undefined;
var islandMoved = false;
function mousedown (e) {
	console.log("mousedown")
	var o = getPoint(e);
	islandMove = {x:o.x,y:o.y};
	islandMoved = false;
}
function mouseup (e) {
	islandMove = undefined;
	setTimeout(function () {
		islandMoved = false;
	})
}
function mousemove (e) {
	if (islandMove && !startTouchPitch) {
		islandMoved = true;
		var o = getPoint(e);
		islandMoveFunc(o)
	}
}

function islandMoveFunc(o, wheelSign) {
	var $case = $(".islandCase");
	var $island = $(".island");

	var cw = $case.width();
	var ch = $case.height();
	var iw = $island.width();
	var ih = $island.height();

	if (wheelSign) {
		var tileindex = $("[tileindex].hover").attr("tileindex")
		if (typeof tileindex !== "undefined") {
			var x = $island.offset().left+(iw/2)
			var y = $island.offset().top+(ih/2)

			var dx = wheelSign*(x-gGame.tiles[tileindex].obj.offset().left)/(gConfig.Unit/10)
			var dy = wheelSign*(y-gGame.tiles[tileindex].obj.offset().top)/(gConfig.Unit/10)
		} else {
			var dx = 0
			var dy = 0
		}

		console.log(dx + " : " + dy + "("+o.x + " : " + o.y+")")
	} else {
		var dx = o.x-islandMove.x
		var dy = o.y-islandMove.y
	}

	var tox = parseInt($island.css("left")) + dx;
	var toy = parseInt($island.css("top")) + dy;

	var topMax = (ih-ch<0)?0:(ih-ch)/2;
	var leftMax = (iw-cw<0)?0:(iw-cw)/2;
	toy = lockUpValueRange(toy, -topMax, topMax);
	tox = lockUpValueRange(tox, -leftMax, leftMax);

	$island.css("left", tox);
	$island.css("top", toy);
	islandMove = {x:o.x,y:o.y};
}

function mousewheel (e) {
	islandMove = {x:0,y:0};
	
	if (e.deltaY < 0 || e.originalEvent.deltaY < 0) {
		var unit = lockUpValueRange(gConfig.Unit+10, 10, 300)
		var sign = 1
	}
	if (e.deltaY > 0 || e.originalEvent.deltaY > 0) {
		var unit = lockUpValueRange(gConfig.Unit-10, 10, 300)
		var sign = -1
	}
	if (gConfig.Unit != unit) {
		ChangeUnit(unit)
		islandMoveFunc(islandMove, sign)
	}
	islandMove = undefined;

}

var tpCache = []
var startDiff = {}
function touchstart (e) {
	var ev = e.originalEvent
	for (var i = 0 ; i < ev.targetTouches.length ; i++ ) {
		tpCache.push(ev.targetTouches[i])
	}
	if (tpCache.length == 2) {
		startDiff = {}
		startDiff.x = Math.abs(tpCache[0].clientX - tpCache[1].clientX)
		startDiff.y = Math.abs(tpCache[0].clientY - tpCache[1].clientY)
	}
}

function touchend (e) {
	tpCache = []
	startDiff = {}
	startTouchPitch = false
}

var startTouchPitch = false
var timeIntervalFleg = true
function touchmove (e) {
	var ev = e.originalEvent
	if (tpCache.length == 2) {
		startTouchPitch = true
		for (var i = 0 ; i < ev.targetTouches.length ; i++ ) {
			if (tpCache[0].identifier == ev.targetTouches[i].identifier) tpCache[0] = ev.targetTouches[i]
			if (tpCache[1].identifier == ev.targetTouches[i].identifier) tpCache[1] = ev.targetTouches[i]
		}

		if (timeIntervalFleg) {
			timeIntervalFleg = false

			var endDiff = {
				x : Math.abs(tpCache[0].clientX-tpCache[1].clientX),
				y : Math.abs(tpCache[0].clientY-tpCache[1].clientY)
			}
		
			var dist = calcDistance(startDiff, endDiff)
		
			var unit = lockUpValueRange(gConfig.Unit+dist, 10, 300)
			if (gConfig.Unit !== unit) {
				ChangeUnit(unit)
				islandMoveFunc(islandMove)
			}
			setTimeout(function () {
				timeIntervalFleg = true
			}, 100)
		}
	
	
	}
}
$(document)
	.on('mousedown touchstart', 'body', mousedown)
	.on('mouseup touchend touchcancel', 'body', mouseup)
	.on('mousemove touchmove', 'body', mousemove)
	.on('mousewheel', 'body', mousewheel)
	.on('touchstart', 'body', touchstart)
	.on('touchend', 'body', touchend)
	.on('touchcancel', 'body', touchend)
	.on('touchmove', 'body', touchmove)
;

function getPoint(e) {
	var point = {x:0, y:0};
	if(e.type == 'touchstart' || e.type == 'touchmove' || e.type == 'touchend' || e.type == 'touchcancel'){
		var touch = e.originalEvent.touches[0] || e.originalEvent.changedTouches[0];
		point.x = touch.pageX;
		point.y = touch.pageY;
	} else if (e.type=='click' || e.type == 'mousedown' || e.type == 'mouseup' || e.type == 'mousemove' || e.type == 'mouseover'|| e.type=='mouseout' || e.type=='mouseenter' || e.type=='mouseleave' || e.type=='mousewheel') {
		point.x = e.pageX;
		point.y = e.pageY;
	}

	return point
}
function getTileFromPoint(point) {
	var jScreen = $("#touchpad");
	var top = jScreen.offset().top;
	var left = jScreen.offset().left;

	var a = (point.x-left)*2/gConfig.Unit - gConfig.Size;
	var b = (point.y-top)/(gConfig.Unit/4) + 2;

	var x = Math.floor((a+b)/2) - 1;
	var y = Math.floor((b-a)/2) - 1;

	if(0 <= x && x < gConfig.Size && 0 <= y && y < gConfig.Size) {
		var tile = gGame.tiles[x + y *gConfig.Size];
		return tile;
	}
}

function connectToServer (addr) {
	var wsUri = "ws://220.76.209.238:8080//websocket/"+addr;
	function connect() {
		var ws = new WebSocket(wsUri)
		ws._init = false;
		ws.onopen = function(e) { onOpen(ws, e) };
		ws.onclose = function(e) { onClose(ws, e) };
		ws.onerror = function(e) { onError(ws, e) };
		ws.onmessage = function(e) { onMessage(ws, e) };
		return ws;
	}

	var ws = connect();
	function onOpen(ws,  e)
	{
		console.log("CONNECTED");
	}

	function onClose(ws,  e)
	{
		console.log("DISCONNECTED");
		// ws = connect();
	}

	function onError(ws,  e)
	{
		console.log("ERROR", e);
	}

}

function onMessage(ws,  e) {
	console.log("MSG", e.data);
	if(!ws._init) {
		ws._init = true;

		var msg = new Buffer(e.data, "hex");
		var sig = loginInfo.Key.sign(msg);
		ws.send(buf2hex(sig.r.toArrayLike(Buffer, "be", 32)) + buf2hex(sig.s.toArrayLike(Buffer, "be", 32)) + "0" + sig.recoveryParam);
	} else {
		if (typeof e.data === "string") {
			var noti = JSON.parse(e.data);
		} else {
			var noti = e.data;
		}
		loginInfo.pushUTXO(noti.utxo)
		switch(noti.type) {
		case 0://Demolition
			console.log("Demolition applied", noti.x, noti.y);
			gGame.tiles[+noti.x + +noti.y * gConfig.Size].Remove()
			gGame.height = noti.height;
			gGame.point_height = noti.point_height;
			gGame.point_balance = noti.point_balance;
			updateResource(gGame.Update());
			break;
		case 1://Upgrade
			var tile = gGame.tiles[noti.x + +noti.y * gConfig.Size]
			if (typeof noti.error == "undefined" || noti.error == "") {
				console.log("Upgrade applied", noti.x, noti.y, noti.area_type, noti.level);
				//gGame.tiles[+noti.x + +noti.y * gConfig.Size].UI.completBuilding(noti.level)
				gGame.height = noti.height;
				gGame.point_height = noti.point_height;
				gGame.point_balance = noti.point_balance;
				if (noti.level == 6) {
					var headTile = tile.obj.headTile
					var o = {x:headTile.x,y:headTile.y}
					for (var i = 0 ; i < 3 ; i++) {
						directByNum(o, i)
						var t = gGame.tiles[o.x + o.y * gConfig.Size];
						t.build_height = noti.height;
						t.level = noti.level;
					}
				}
				tile.build_height = noti.height;
				tile.level = noti.level;
				updateResource(gGame.Update());
			} else { //false
				alert(language[noti.error]||noti.error)
				if(noti.level == 1) {
					tile.Remove()
				} else {
					tile.level = noti.level-1;
				}
				tile.build_height = tile.build_height_old;
				if (noti.level == 5) {
					var headTile = tile.obj.headTile
					var o = {x:headTile.x,y:headTile.y}
					for (var i = 0 ; i < 3 ; i++) {
						directByNum(o, i)
						var t = gGame.tiles[o.x + o.y * gConfig.Size];
						t.build_height = t.build_height_old;
					}
				}
			}
			break;
		}
	}
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
            break;
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
