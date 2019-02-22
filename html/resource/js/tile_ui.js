function newTouchDiv(index) {
	return $("<div tileindex="+index+"/>")
		.append($("<div class='scaleArea'><div class='hoverArea'/></div>"))
}

function newObjDiv(x,y,num) {
	return $("<div/>")
		.css("z-index", (+x+ +y)*10)
		.append($("<img class='floor' src='/images/tile/base_floor/groundtiles_tile"+num+".png'>"))
}

function TileUI(jScreen, touchpad, num) {
	this.num = num
	this.objScreen = jScreen
	this.touchpad = touchpad
}

TileUI.prototype.init = function(tile) {
	this.Tile = tile
	this.x = this.Tile.x
	this.y = this.Tile.y
    this.index = this.x+this.y*gConfig.Size;
	this.obj = newObjDiv(this.x, this.y, this.num);
	this.touch = newTouchDiv(this.index)
	this.objScreen.append(this.obj)
	this.touchpad.append(this.touch)
	this.Resize();
}

TileUI.prototype.Resize = function() {
	this.touch.css("left", ((gConfig.Size+this.x-this.y-1)/2) + "rem");
	this.touch.css("bottom", gConfig.Size/2 - ((this.x+this.y+2)/2)/2 + "rem");

	this.obj.css("left", ((gConfig.Size+this.x-this.y-1)/2) + "rem");
	this.obj.css("bottom", gConfig.Size/2 - ((this.x+this.y+2)/2)/2 + "rem");
	return this
}

TileUI.prototype.Hover = function () {
	$(".hover").removeClass("hover");
	this.touch.addClass("hover");
	printInfo(this.x, this.y);
}

TileUI.prototype.SelectTile = function () {
	$(".selected").removeClass("selected");

	var checker = this.Tile.CheckLvRound(6);
	if (this.Tile.headTile) {
		var c = this.Tile.headTile.candidate;
		for (var i = 0 ; i < c.length ; i++) {
			c[i].UI.touch.addClass("selected");
		}
	} else {
		this.touch.addClass("selected");
	}
	if (checker.CheckLvF(function (t, headT) {
	}) == false) {
	}
}
TileUI.prototype.SelectUpperLvTile = function(lv) {
	$(".selected").removeClass("selected");
	this.touch.addClass("selected");

	var checker = this.Tile.CheckLvRound(lv);
	if (checker.CheckLvF(function (t, headT) {
		message(t);
		t.UI.touch.addClass("selected").addClass("hover");
	}) == false) {
		this.touch.addClass("selected");
	}
}
TileUI.prototype.startBuild = function() {
	this.Tile.BuildProcessing = true
}
TileUI.prototype.BuildUp = function() {
	this.startBuild()
	var lv = this.Tile.level + 1;
	var targetTile = this.Tile;
	this.touch.find(".hoverArea").addClass(this.Tile.TypeName());
	this.obj.find("img.floor").attr("src", "/images/tile/building_floor.png");
	if (lv < 5) {
		for (var i = 0 ; i < lv ; i++) {
			if (i == lv-1) {
				var $img = $("<img class='building lv"+(lv)+"' src='/images/building/construction.png'/>")
			} else {
				var $img = $("<img class='building lv"+(i+1)+"' src='/images/building/"+this.Tile.TypeName()+"_Lv1.png'/>")
			}
			this.obj.append($img);
		}
	} else if (lv == 5) {
		this.obj.find(".building").detach();
		var $img = $("<img class='building lv5' src='/images/building/construction.png'/>")
		this.obj.append($img);
	} else if (lv == 6) {
		var checker = this.Tile.CheckLvRound(5);
		if (checker.CheckLvF(function (t, headT) {
			t.UI.obj.find(".building").detach();
			t.UI.startBuild()
		}) == true) {// buildable lvF
			var $img = $("<img class='building lv6' src='/images/building/construction.png'/>")
			this.Tile.headTile.UI.obj.append($img);
			targetTile = this.Tile.headTile
		}
	}

	return targetTile;
}

TileUI.prototype.completBuilding = function (effect) {
	var lv = this.Tile.level;
	if (lv == 6) {
		var checker = this.Tile.headTile.CheckLvRound(6)
		if (checker.CheckLvF(function (t, headT) {
			t.BuildProcessing = false;
			if (t.index == headT.index) {
				t.level = 6;
			} else {
				t.level = 0;
			}

			t.UI.obj.css("z-index", t.UI.obj.css("z-index")-1)
		}))
		this.obj.find("img.floor").attr("src", "/images/tile/"+this.Tile.TypeName()+"_LvFLETA-Tile.png").addClass("lv6");
		var fileTail = "_LvFLETA"
	} else  {
		this.Tile.BuildProcessing = false
		if (this.Tile.level < 5) {
			var fileTail = "_Lv1"
		} else {
			var fileTail = "_Lv5"
		}
	}
	this.obj.find(".lv"+lv+".building").attr("src", "/images/building/"+this.Tile.TypeName()+""+fileTail+".png")
	if (effect != "noEffect") {
		this.constructEffect()
	}
	this.touch.find(".underconstruction").remove()


	var $menu = $("#menu");
	if ($menu[0].target == this.Tile) {
		menuOpen(this.Tile)
	}
}

TileUI.prototype.ShowBuildProcessingTime = function(sec) {
	var lv = this.Tile.level
	var uc = this.touch.find(".underconstruction")
	if (uc.length == 0) {
		uc = $("<div class='lv"+lv+" underconstruction'>")
		this.touch.append(uc)
	}
	uc.html(secondToDate(sec))
}

TileUI.prototype.distructionEffect = function(callback) {
	this.buildEffect("distructionEffect", callback)
}

TileUI.prototype.constructEffect = function(callback) {
	this.buildEffect("constructEffect", callback)
}

TileUI.prototype.buildEffect = function(type, callback) {
	if (this.Tile.headTile) {
		var tileUI = this.Tile.headTile.UI
	} else {
		var tileUI = this
	}

	var effect = $("<div class='lv"+tileUI.Tile.level+" buildEffect "+type+"'/>")
	tileUI.touch.append(effect);
	(function (effect, tileUI, callback) {
		setTimeout(function () {
			if (typeof callback === "function") {
				callback(tileUI)
			}
			effect.remove()
		}, 1500)
	})(effect, tileUI, callback)

	tileUI.completEffect()
}

TileUI.prototype.fletaEffect = function() {
	var effect = $("<div class='FLETAAnimation lv"+this.Tile.level+"'/>")
	this.touch.append(effect);
	(function (effect) {
		setTimeout(function () {
			effect.remove()
		}, 3000)
	})(effect)
}

TileUI.prototype.completEffect = function() {
	var effect = $("<div class='completAnimation lv"+this.Tile.level+"'/>")
	this.touch.append(effect);
	(function (effect, tileUI) {
		setTimeout(function () {
			effect.remove()
			if (tileUI.Tile.level == 6) {
				tileUI.fletaEffect()
			}
		}, 3000)
	})(effect, this)
}

TileUI.prototype.Remove = function() {
	this.obj.find(".building").detach();
	this.touch.find(".hoverArea").attr("class", "hoverArea");
	this.obj.find(".floor").attr("src", "/images/tile/base_floor/groundtiles_tile"+this.num+".png").attr("class", "floor");
	this.obj.css("z-index", (+this.x + +this.y)*10)
}
