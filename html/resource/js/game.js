$(document).on('click', '#touchpad', function(e) {
	var point = {x:0, y:0};
	if(e.type == 'touchstart' || e.type == 'touchmove' || e.type == 'touchend' || e.type == 'touchcancel'){
		var touch = e.originalEvent.touches[0] || e.originalEvent.changedTouches[0];
		point.x = touch.pageX;
		point.y = touch.pageY;
	} else if (e.type=='click' || e.type == 'mousedown' || e.type == 'mouseup' || e.type == 'mousemove' || e.type == 'mouseover'|| e.type=='mouseout' || e.type=='mouseenter' || e.type=='mouseleave') {
		point.x = e.pageX;
		point.y = e.pageY;
	}

	var jScreen = $("#touchpad");
	var top = parseInt(jScreen.css("top"));

	var a = point.x*2/gConfig.Unit - gConfig.Size;
	var b = (point.y - top/2)*4/gConfig.Unit - 2;

	var x = Math.floor((a+b)/2) - 1;
	var y = Math.floor((b-a)/2) - 1;

	if(0 <= x && x < gConfig.Size && 0 <= y && y < gConfig.Size) {
		var tile = Tiles[x + y *gConfig.Size];
		tile.Menu();
	}
});

$(document).on('mousemove', '#touchpad', function(e) {
	var point = {x:0, y:0};
	if(e.type == 'touchstart' || e.type == 'touchmove' || e.type == 'touchend' || e.type == 'touchcancel'){
		var touch = e.originalEvent.touches[0] || e.originalEvent.changedTouches[0];
		point.x = touch.pageX;
		point.y = touch.pageY;
	} else if (e.type=='click' || e.type == 'mousedown' || e.type == 'mouseup' || e.type == 'mousemove' || e.type == 'mouseover'|| e.type=='mouseout' || e.type=='mouseenter' || e.type=='mouseleave') {
		point.x = e.pageX;
		point.y = e.pageY;
	}

	var jScreen = $("#touchpad");
	var top = parseInt(jScreen.css("top"));

	var a = point.x*2/gConfig.Unit - gConfig.Size;
	var b = (point.y - top/2)*4/gConfig.Unit - 2;

	var x = Math.floor((a+b)/2) - 1;
	var y = Math.floor((b-a)/2) - 1;


	if(0 <= x && x < gConfig.Size && 0 <= y && y < gConfig.Size) {
		var tile = Tiles[x + y *gConfig.Size];
		tile.Hover()
	}
});

function menuClose () {
	$(".selected").removeClass("selected");
	$menu = $("#menu");
	delete $menu[0].target;
	$menu.hide();
	deleteMenu()
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
	this.obj.append($("<image src='tile/base_floor/groundtiles_tile"+num+".png' style='position:absolute; width:"+gConfig.Unit+"px; bottom:0px;'/>"));
	this.obj.level = 0;
	this.Resize();
}

Tile.prototype.Type = "empty"

Tile.prototype.Menu = function() {
	deleteMenu()
	addMenu(MENU["lv"+this.obj.level])
	message("menu open x : " + this.x + " y : " + this.y )
	$("#cover").show();
	var $menu = $("#menu");
	$menu[0].target = this;
	this.SelectTile()
	$menu.show();
	return this
}

Tile.prototype.RunCommand = function(func) {
	if (typeof this[func] === "function") {
		message("command : "+ func + " x : " + this.x + " y : " + this.y )
		this[func]()
		this.Menu()
	}
	return this
}

Tile.prototype.Hover = function() {
	$(".hover").removeClass("hover")
	this.touch.addClass("hover")
	printInfo(this.x, this.y)
	return this
}

Tile.prototype.SelectUpperLvTile = function(lv) {
	$(".selected").removeClass("selected")
	this.touch.addClass("selected")
	
	var candidate = this.CheckLvRound(lv)
	if (candidate !== false) {
		for ( var i = 0 ; i < candidate.length ; i++ ) {
			message(candidate[i])
			var tile = Tiles[candidate[i]];
			tile.touch.addClass("selected").addClass("hover")
		}
	} else {
		this.touch.addClass("selected")
	}
	return this
}
Tile.prototype.SelectTile = function() {
	$(".selected").removeClass("selected")
	if (this.obj.level == 6) {
		var head = Tiles[this.obj.headTile]
		head.SelectUpperLvTile(6)
	} else {
		this.touch.addClass("selected")
	}
	return this
}
Tile.prototype.CheckLvRound = function(checkLv) {
	if (typeof checkLv === "undefined") {
		checkLv = 5
	}
	if (checkLv == 6) {
		var o = getXYFromIndex(this.obj.headTile)
	} else {
		var o = {x : this.x, y : this.y}
	}

	var tile = Tiles[o.x + o.y *gConfig.Size];
	var type = tile.Type
	if (tile.obj.level != checkLv) {
		return false
	}

	for ( var i = 0 ; i < 4 ; i++ ) {
		var tile = Tiles[o.x + o.y * gConfig.Size];
		var candidate = []
		if (tile.obj.level == checkLv && type == tile.Type) {
			for ( var j = i ; j < i+4 ; j++ ) {
				directByNum(o, j%4);
				if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
					var tile = Tiles[o.x + o.y * gConfig.Size];
					if (typeof tile !== "undefined") {
						if (tile.obj.level == checkLv && type == tile.Type) {
							candidate.push(o.x + o.y * gConfig.Size)
						}
					}
				}
			}
		}
		if (candidate.length == 4) {
			return candidate
		}
	}
	return false

}
Tile.prototype.Demolition = function() {
	if (this.obj.level == 6) {
		var list = this.CheckLvRound(6)
		for ( var i = 0 ; i < list.length ; i++ ) {
			Tiles[list[i]].Remove().UpdateInfo()
		}
	} else {
		this.Remove()
	}
	menuClose()
	return this
}

Tile.prototype.UpdateInfo = function() {
	if (this.obj.level == 0) {
		this.touch.find("span").html("")
	} else {
		this.touch.find("span").html(this.Type + "<br>lv" + this.obj.level)
	}
	return this
}

Tile.prototype.Remove = function() {
	this.obj.find(".building").detach();
	this.obj.level = 0;
	this.Type = "empty";
	this.touch.find(".hoverArea").attr("class", "hoverArea")

	delete this.obj.headTile
	return this
}

Tile.prototype.Industrial = function() {
	this.Build("Industrial")
	return this
}
Tile.prototype.Residential = function() {
	this.Build("Residential")
	return this
}
Tile.prototype.Commercial = function() {
	this.Build("Commercial")
	return this
}
Tile.prototype.Upgrade = function() {
	this.Build()
	return this
}

Tile.prototype.Build = function(type) {
	switch(this.obj.level) {
	case 0:
		this.obj.level = 1;
		this.touch.find(".hoverArea").addClass(type)
		this.obj.level = 1;
		this.Type = type;
		this.obj.find("img").attr("src", "tile/building_floor.png");
		//var jImg = $("<image class='building' src='building/construction.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		var jImg = $("<image class='building' src='building/"+this.Type+"_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit/4)+"px");
		jImg.css("bottom", (gConfig.Unit/4)+"px");
		jImg.css("z-index", 1);
		break;
	case 1:
		this.obj.level = 2;
		var jImg = $("<image class='building' src='building/"+this.Type+"_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit*2/4)+"px");
		jImg.css("bottom", (gConfig.Unit/2/4)+"px");
		jImg.css("z-index", 2);
		break;
	case 2:
		this.obj.level = 3;
		var jImg = $("<image class='building' src='building/"+this.Type+"_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit/4)+"px");
		jImg.css("z-index", 4);
		break;
	case 3:
		this.obj.level = 4;
		var jImg = $("<image class='building' src='building/"+this.Type+"_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("bottom", (gConfig.Unit/2/4)+"px");
		jImg.css("z-index", 3);
		break;
	case 4:
		this.obj.level = 5;
		this.obj.find(".building").detach();
		var jImg = $("<image class='building' src='building/"+this.Type+"_Lv5.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit)+"px");
		break;
	case 5:
		var candidate = this.CheckLvRound()
		if (candidate !== false) {
			var maxCand = 0;
			for ( var i = 0 ; i < candidate.length ; i++ ) {
				var tile = Tiles[candidate[i]]
				tile.obj.level = 6;
				tile.UpdateInfo()
				message("fleta!! " + tile.x + " : " + tile.y)
				tile.obj.find(".building").detach();
				if (maxCand < candidate[i]) {
					maxCand = candidate[i]
				}
			}
			for ( var i = 0 ; i < candidate.length ; i++ ) {
				var tile = Tiles[candidate[i]]
				tile.obj.headTile = maxCand
			}

			var tile = Tiles[maxCand];
			var jImg = $("<image class='building' src='building/"+this.Type+"_LvFLETA.png' style='position:absolute; bottom:0px;'/>").appendTo(tile.obj)
			jImg.css("width", (gConfig.Unit*2)+"px");
			jImg.css("left", -(gConfig.Unit/2)+"px");
		}
		break;
	}
	printInfo(this.x, this.y)
	return this
};

Tile.prototype.Resize = function() {
	$("#cssControll").html("#touchpad > div,#screen > div {width:"+gConfig.Unit+"px; height:"+ gConfig.Unit/2+"px;}")
	this.touch.css("left", (gConfig.Unit*(gConfig.Size+this.x-this.y-1)/2) + "px");
	this.touch.css("bottom", gConfig.Unit*gConfig.Size/2 - (gConfig.Unit*(this.x+this.y+2)/2)/2 + "px");

	this.obj.css("left", (gConfig.Unit*(gConfig.Size+this.x-this.y-1)/2) + "px");
	this.obj.css("bottom", gConfig.Unit*gConfig.Size/2 - (gConfig.Unit*(this.x+this.y+2)/2)/2 + "px");
	this.obj.find("img").css("width", gConfig.Unit+"px");

	switch(this.obj.level) {
	case 6:
		//TODO
	case 5:
		var jBuilding = this.obj.find(".building").eq(0);
		jBuilding.css("width", (gConfig.Unit)+"px");
		break;
	case 4:
		var jBuilding = this.obj.find(".building").eq(3);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("bottom", (gConfig.Unit/2/4)+"px");
	case 3:
		var jBuilding = this.obj.find(".building").eq(2);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit/4)+"px");
	case 2:
		var jBuilding = this.obj.find(".building").eq(1);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit*2/4)+"px");
		jBuilding.css("bottom", (gConfig.Unit/2/4)+"px");
	case 1:
		var jBuilding = this.obj.find(".building").eq(0);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit/4)+"px");
		jBuilding.css("bottom", (gConfig.Unit/4)+"px");
		break;
	}
	return this
}



function ChangeUnit(unit) {
	gConfig.Unit = unit;

	var jScreen = $("#screen");
	jScreen.css("width", (gConfig.Unit*gConfig.Size)+"px");
	jScreen.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
	jScreen.css("top", (gConfig.Unit/2*4)+"px");
	var jTouchPad = $("#touchpad");
	jTouchPad.css("width", (gConfig.Unit*gConfig.Size)+"px");
	jTouchPad.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
	jTouchPad.css("top", (gConfig.Unit/2*4)+"px");

	for(var i=0; i<Tiles.length; i++) {
		Tiles[i].Resize();
	}
}