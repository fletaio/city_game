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
		tile.UI.Hover();
	}
});

var islandMove = undefined;
var downPosition;
var islandMoved = false;
function mousedown (e) {
	var o = getPoint(e);
	downPosition = {x:o.x,y:o.y};
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
		var o = getPoint(e);
		if (islandMoved == false) {
			var d = calcDistance(downPosition, o);
			if (d >= 5 || d <= -5) {
				islandMoved = true;
			}
		}
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
			var x = $island.offset().left+(iw/2) - gConfig.Unit/2
			var y = $island.offset().top+(ih/2) - gConfig.Unit/4

			var dx = wheelSign*(x-gGame.tiles[tileindex].UI.obj.offset().left)/(gConfig.Unit/10)
			var dy = wheelSign*(y-gGame.tiles[tileindex].UI.obj.offset().top)/(gConfig.Unit/10)
		} else {
			var dx = 0
			var dy = 0
		}

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
	try {
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
	
		if(!e){ e = window.event; } /* IE7, IE8, Chrome, Safari */
		if(e.preventDefault) { e.preventDefault(); } /* Chrome, Safari, Firefox */
		e.returnValue = false; /* IE7, IE8 */
	} catch (e) {
		islandMove = undefined;
	}
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
		
			var dist = calcDistance(startDiff, endDiff)/10
		
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

$(function () {
	$(".islandCase:last")
		.on('mousedown touchstart', mousedown)
		.on('mouseup touchend touchcancel', mouseup)
		.on('mousemove touchmove', mousemove)
		.on('mousewheel', mousewheel)
		.on('touchstart', touchstart)
		.on('touchend', touchend)
		.on('touchcancel', touchend)
		.on('touchmove', touchmove)
	;
})

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

var disconnectedCount = 1
function connectToServer (addr) {
	if (location.protocol != 'https:')	{
		var wsUri = "ws://"+window.location.host+"/websocket/"+addr;
	} else {
		var wsUri = "wss://"+window.location.host+"/websocket/"+addr;
	}
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
		disconnectedCount = 1
		console.log("CONNECTED");
	}

	function onClose(ws,  e)
	{
		disconnectedCount = (disconnectedCount+1) * 2
		console.log("DISCONNECTED");
		(function () {
			setTimeout(function () {
				ws = connect();
			}, 1000*disconnectedCount)
		})()
	}

	function onError(ws,  e)
	{
		console.log("ERROR", e);
	}

}

function onMessage(ws,  e) {
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

		gGame.height = noti.height
		gGame.point_balance = noti.point_balance
		if (typeof noti.fleta_city_coins !== "undefined") {
			for (var k in noti.fleta_city_coins) {
				if (typeof gGame.coin_list[k] === "undefined") {
					gGame.coin_list[k] = new FletaCityCoin(noti.fleta_city_coins[k])
				}
			}
			for (var k in gGame.coin_list) {
				if (typeof noti.fleta_city_coins[k] === "undefined") {
					gGame.coin_list[k].Remove()
				}
			}
		}
		if (noti.coin_count) {
			$("[key='coin_count']").html(noti.coin_count)
		}

		SendQueue.NewUTXO(noti.utxo)
		addTx(noti.tx)
		

		switch(noti.type) {
		case 0://Demolition
			var tile = gGame.tiles[noti.x + +noti.y * gConfig.Size]
			if (typeof noti.error == "undefined" || noti.error == "") {
				tile.UI.distructionEffect(function (tileUI) {
					tileUI.Tile.Remove()
				})
				gGame.height = noti.height;
				gGame.point_height = noti.point_height;
				gGame.point_balance = noti.point_balance;
				updateResource(gGame.Update());
			}

			break;
		case 1://Upgrade
			var tile = gGame.tiles[noti.x + +noti.y * gConfig.Size]
			if (typeof noti.error == "undefined" || noti.error == "") {
				gGame.height = noti.height;
				gGame.point_height = noti.point_height;
				gGame.point_balance = noti.point_balance;
				tile.Build(noti.area_type)
				if (noti.level == 6) {
					var checker = tile.CheckLvRound(5)
					
					checker.CheckLvF(function (t, headT) {
						t.build_height = noti.height;
						t.nextLevel = noti.level;
					})
				}
				tile.build_height = noti.height;
				tile.nextLevel = noti.level;
				updateResource(gGame.Update());
			} else { //false
				Alert(language[noti.error]||noti.error)
				if(noti.level == 1) {
					tile.Remove()
				}
				tile.build_height = tile.build_height_old;
				var checker = tile.CheckLvRound(5)
				checker.CheckLvF(function (t, headT) {
					t.build_height = t.build_height_old;
					delete t.headTile
				})
			}
			break;
		case 99://duplicated connection
			alert(language["duplicated connection"]);
			location.reload();
			break;
		}
	}
}

function addKeyShotcut () {
	document.addEventListener('keydown', function(event) {
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
				menuOpen(gGame.tiles[o.x+o.y*gConfig.Size].UI.Hover())
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
			case 79: //o
				$("#alertUI .alertUI_ok").click()
				break;
			case 88: //x
				$("#alertUI .alertUI_cancel").click()
				break;
		
			default:
				break;
		}
	});
}
