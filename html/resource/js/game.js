$(document).on('click', '#touchpad', function(e) {
	if (!islandMoved) {
		var tile = getTileFromPoint(e)
		if (typeof tile !== "undefined") {
			tile.Menu();
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

function menuClose () {
	$(".selected").removeClass("selected");
	$(".hover").removeClass("hover");
	$menu = $("#menu");
	delete $menu[0].target;
	$menu.hide();
	deleteMenu();
	$("#cover").hide();
}

function Tile(jScreen, $touchpad, x, y, num) {
	this.x = x;
	this.y = y;
	this.touch = $("<div/>").appendTo($touchpad);
	this.touch.append($("<div class='scaleArea'><div class='hoverArea'/></div>"));
	this.touch.append($("<span/>"));
	this.obj = $("<div/>").appendTo(jScreen);
	this.obj.css("z-index", x*gConfig.Size+y);
	this.obj.append($("<img class='floor' src='/images/tile/base_floor/groundtiles_tile"+num+".png'>"));
	this.obj.level = 0;
	this.Resize();
}

Tile.prototype.Type = "empty";

Tile.prototype.Menu = function() {
	deleteMenu();
	addMenu(MENU["lv"+this.obj.level]);
	message("menu open x : " + this.x + " y : " + this.y );
	$("#cover").show();
	var $menu = $("#menu");
	$menu[0].target = this;
	this.SelectTile();
	$menu.show();
	return this;
}

Tile.prototype.RunCommand = function(func) {
	if (typeof this[func] === "function") {
		message("command : "+ func + " x : " + this.x + " y : " + this.y );
		this[func]();
		this.Menu();
	}
	return this
}

Tile.prototype.Hover = function() {
	$(".hover").removeClass("hover");
	this.touch.addClass("hover");
	printInfo(this.x, this.y);
	return this;
}

Tile.prototype.SelectUpperLvTile = function(lv) {
	$(".selected").removeClass("selected");
	this.touch.addClass("selected");
	
	var candidate = this.CheckLvRound(lv);
	if (candidate !== false) {
		for ( var i = 0 ; i < candidate.length ; i++ ) {
			message(candidate[i]);
			var tile = Tiles[candidate[i]];
			tile.touch.addClass("selected").addClass("hover");
		}
	} else {
		this.touch.addClass("selected");
	}
	return this;
}
Tile.prototype.SelectTile = function() {
	$(".selected").removeClass("selected");
	if (this.obj.level == 6) {
		var head = Tiles[this.obj.headTile];
		head.SelectUpperLvTile(6);
	} else {
		this.touch.addClass("selected");
	}
	return this;
}
Tile.prototype.CheckLvRound = function(checkLv) {
	if (typeof checkLv === "undefined") {
		checkLv = 5;
	}
	if (checkLv == 6) {
		var o = getXYFromIndex(this.obj.headTile);
	} else {
		var o = {x : this.x, y : this.y};
	}

	var tile = Tiles[o.x + o.y *gConfig.Size];
	var type = tile.Type;
	if (tile.obj.level != checkLv) {
		return false;
	}

	for ( var i = 0 ; i < 4 ; i++ ) {
		var tile = Tiles[o.x + o.y * gConfig.Size];
		var candidate = [];
		if (tile.obj.level == checkLv && type == tile.Type) {
			for ( var j = i ; j < i+4 ; j++ ) {
				directByNum(o, j%4);
				if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
					var tile = Tiles[o.x + o.y * gConfig.Size];
					if (typeof tile !== "undefined") {
						if (tile.obj.level == checkLv && type == tile.Type) {
							candidate.push(o.x + o.y * gConfig.Size);
						}
					}
				}
			}
		}
		if (candidate.length == 4) {
			return candidate;
		}
	}
	return false;

}
Tile.prototype.Demolition = function() {
	if (this.obj.level == 6) {
		var list = this.CheckLvRound(6)
		for ( var i = 0 ; i < list.length ; i++ ) {
			Tiles[list[i]].Remove().UpdateInfo();
		}
	} else {
		this.Remove();
	}
	menuClose();
	return this;
}

Tile.prototype.UpdateInfo = function() {
	if (this.obj.level == 0) {
		this.touch.find("span").html("");
	} else {
		this.touch.find("span").html(this.Type + "<br>lv" + this.obj.level);
	}
	return this
}

Tile.prototype.Remove = function() {
	this.obj.find(".building").detach();
	this.obj.level = 0;
	this.Type = "empty";
	this.touch.find(".hoverArea").attr("class", "hoverArea");
	this.obj.find(".floor").attr("src", "/images/tile/building_floor.png").attr("class", "floor");

	delete this.obj.headTile;
	return this;
}

Tile.prototype.Industrial = function() {
	this.Build("Industrial");
	return this;
}
Tile.prototype.Residential = function() {
	this.Build("Residential");
	return this;
}
Tile.prototype.Commercial = function() {
	this.Build("Commercial");
	return this;
}
Tile.prototype.Upgrade = function() {
	this.Build();
	return this;
}

Tile.prototype.Build = function(type) {
	switch(this.obj.level) {
	case 0:
		this.obj.level = 1;
		this.touch.find(".hoverArea").addClass(type);
		this.obj.level = 1;
		this.Type = type;
		this.obj.find("img.floor").attr("src", "/images/tile/building_floor.png");
		this.obj.append($("<img class='building lv1' src='/images/building/"+this.Type+"_Lv1.png'/>"));
		break;
	case 1:
		this.obj.level = 2;
		this.obj.append($("<img class='building lv2' src='/images/building/"+this.Type+"_Lv1.png'/>"));
		break;
	case 2:
		this.obj.level = 3;
		this.obj.append($("<img class='building lv3' src='/images/building/"+this.Type+"_Lv1.png'/>"));
		break;
	case 3:
		this.obj.level = 4;
		this.obj.append($("<img class='building lv4' src='/images/building/"+this.Type+"_Lv1.png'/>"));
		break;
	case 4:
		this.obj.level = 5;
		this.obj.find(".building").detach();
		this.obj.append($("<img class='building lv5' src='/images/building/"+this.Type+"_Lv5.png'/>"));
		break;
	case 5:
		var candidate = this.CheckLvRound();
		if (candidate !== false) {
			var maxCand = 0;
			for ( var i = 0 ; i < candidate.length ; i++ ) {
				var tile = Tiles[candidate[i]];
				tile.obj.level = 6;
				tile.UpdateInfo();
				tile.obj.find(".building").detach();
				if (maxCand < candidate[i]) {
					maxCand = candidate[i];
				}
			}
			for ( var i = 0 ; i < candidate.length ; i++ ) {
				var tile = Tiles[candidate[i]];
				tile.obj.headTile = maxCand;
			}

			var tile = Tiles[maxCand];
			message("fleta!! " + tile.x + " : " + tile.y);

			tile.obj.find("img.floor").attr("src", "/images/tile/"+this.Type+"_LvFLETA-Tile.png").addClass("lvF");
			tile.obj.append($("<img class='building lvF' src='/images/building/"+this.Type+"_LvFLETA.png'/>"));
		}
		break;
	}
	printInfo(this.x, this.y);
	return this;
};

Tile.prototype.Resize = function() {
	this.touch.css("left", ((gConfig.Size+this.x-this.y-1)/2) + "rem");
	this.touch.css("bottom", gConfig.Size/2 - ((this.x+this.y+2)/2)/2 + "rem");

	this.obj.css("left", ((gConfig.Size+this.x-this.y-1)/2) + "rem");
	this.obj.css("bottom", gConfig.Size/2 - ((this.x+this.y+2)/2)/2 + "rem");
	return this
}

function ChangeUnit(unit) {
	gConfig.Unit = unit;


	var h = [], i =0
	h[i++] = ".island{width:"+(gConfig.Size*1.086875)+"rem;height:"+(gConfig.Size*0.805)+"rem}"
	
// {
//     top: 4.03rem;
//     left: 0.7rem;
// }

	h[i++] = "#tileCase{top:"+(gConfig.Size*0.251875)+"rem;left:"+(gConfig.Size*0.04375)+"rem}"

	$("#cssControll").html(h.join("\n"));
	$("html").css("font-size", gConfig.Unit+"px");
}
