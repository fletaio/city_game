$(document).on('click', '#touchpad', function(e) {
	if (!islandMoved) {
		var tile = getTileFromPoint(e)
		if (typeof tile !== "undefined") {
			menuOpen(tile)
		}
	}
});

$(document).on('click', '#cover', function(e) {
	if (!islandMoved) {
		menuClose();
	}
});

$(document).on('mousemove', '#touchpad', function(e) {
	var tile = getTileFromPoint(e)
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

		var $case = $(".islandCase");
		var $island = $(".island");
		var tox = parseInt($island.css("left")) + o.x-islandMove.x;
		var toy = parseInt($island.css("top")) + o.y-islandMove.y;

		var cw = $case.width();
		var ch = $case.height();
		var iw = $island.width();
		var ih = $island.height();

		var topMax = (((ch-ih)/2)+ih)-100;
		var leftMax = (((cw-iw)/2)+iw)-100;

		toy = lockUpValueRange(toy, -topMax, topMax);
		tox = lockUpValueRange(tox, -leftMax, leftMax);

		$island.css("left", tox);
		$island.css("top", toy);
		islandMove = {x:o.x,y:o.y};
	}
}

$(document)
	.on('mousedown touchstart', 'body', mousedown)
	.on('mouseup touchend', 'body', mouseup)
	.on('mousemove touchmove', 'body', mousemove)
;

function getPoint(e) {
	var point = {x:0, y:0};
	if(e.type == 'touchstart' || e.type == 'touchmove' || e.type == 'touchend' || e.type == 'touchcancel'){
		var touch = e.originalEvent.touches[0] || e.originalEvent.changedTouches[0];
		point.x = touch.pageX;
		point.y = touch.pageY;
	} else if (e.type=='click' || e.type == 'mousedown' || e.type == 'mouseup' || e.type == 'mousemove' || e.type == 'mouseover'|| e.type=='mouseout' || e.type=='mouseenter' || e.type=='mouseleave') {
		point.x = e.pageX;
		point.y = e.pageY;
	}
	return point
}
function getTileFromPoint(e) {
	var point = getPoint(e);
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
