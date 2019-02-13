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
	if (islandMove) {
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

			var dx = wheelSign*(x-Tiles[tileindex].obj.offset().left)/(gConfig.Unit/10)
			var dy = wheelSign*(y-Tiles[tileindex].obj.offset().top)/(gConfig.Unit/10)
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
function touchstart (e) {
	var ev = e.originalEvent
	if (ev.targetTouches.length == 2) {
		alert(ev.targetTouches.length)
		for (var i=0; i < ev.targetTouches.length; i++) {
			tpCache.push(ev.targetTouches[i]);
		}
	}
}
function touchend (e) {

}
function touchmove (e) {
	var ev = e.originalEvent
	if (ev.targetTouches.length == 2 && ev.changedTouches.length == 2) {
		// Check if the two target touches are the same ones that started
		// the 2-touch
		var point1=-1, point2=-1;
		for (var i=0; i < tpCache.length; i++) {
		  if (tpCache[i].identifier == ev.targetTouches[0].identifier) point1 = i;
		  if (tpCache[i].identifier == ev.targetTouches[1].identifier) point2 = i;
		}
		if (point1 >=0 && point2 >= 0) {
		  // Calculate the difference between the start and move coordinates
		  var diff1 = Math.abs(tpCache[point1].clientX - ev.targetTouches[0].clientX);
		  var diff2 = Math.abs(tpCache[point2].clientX - ev.targetTouches[1].clientX);
	 
		  // This threshold is device dependent as well as application specific
		  var PINCH_THRESHHOLD = ev.target.clientWidth / 10;
		  if (diff1 >= PINCH_THRESHHOLD && diff2 >= PINCH_THRESHHOLD)
			  ev.target.style.background = "green";
		}
		else {
		  // empty tpCache
		  tpCache = new Array();
		}
	  }
}
$(document)
	.on('touchstart', 'body', touchstart)
	.on('touchend', 'body', touchend)
	.on('touchcancel', 'body', touchend)
	.on('touchmove', 'body', touchmove)
	.on('mousedown touchstart', 'body', mousedown)
	.on('mouseup touchend', 'body', mouseup)
	.on('mousemove touchmove', 'body', mousemove)
	.on('mousewheel', 'body', mousewheel)
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
		var tile = Tiles[x + y *gConfig.Size];
		return tile;
	}
}
