function menuClose () {
	$("#alertUI").hide()
	$(".tooltip").remove()
	$(".selected").removeClass("selected");
	$(".hover").removeClass("hover");
	$menu = $("#menu");
	delete $menu[0].target;
	$menu.hide();
	deleteMenu();
}

function deleteMenu() {
    $("#menu").html("")
}

function addMenu(tile, funcs) {
    for (var key in funcs) {
		var able = buildableResource(tile, funcs[key])
		var btn = $("<button id=\""+funcs[key]+"\" value=\""+key+"\">"+key+"</button>")

		if (able != true) {
			btn.addClass("disable")
			btn.attr("onclick", "event.stopPropagation();alert('"+able+"');")
		} else {
			btn.attr("onclick", "event.stopPropagation();$('#menu')[0].target.RunCommand('"+funcs[key]+"');")
		}
		if (funcs[key] !== "Demolition") {
			var $tooltip = $("#tooltip").clone();
			var $this = btn

			$tooltip.removeAttr("id")
			$tooltip.attr("class", "tooltip " + funcs[key])
			
			if (tile.obj.level == 0) {
				var type = buildingNum($this.attr("id"))
			} else {
				var type = tile.type
			}
			var r = gBuildingDefine[type][tile.obj.level]
			$tooltip.find("#needDollar").html(r.cost_usage)
			$tooltip.find("#needPower").html(r.power_usage)
			$tooltip.find("#needDemographic").html(r.man_usage)
			$tooltip.find("#resource").html("+"+r.output+"/s").attr("class", buildingType(type))


			var time = secondToDate(r.build_time);
			if (time.length > 8) {
				$tooltip.find("#needTime").html(time).attr("class", "date")
			} else if (time.length > 5) {
				$tooltip.find("#needTime").html(time).attr("class", "hour")
			} else {
				$tooltip.find("#needTime").html(time).attr("class", "")
			}

			$("#menu").append($tooltip)
		}
		$("#menu").append(btn)
    }
}

function secondToDate(sec) {
	var ss = sec%60
	sec = parseInt(sec/60)
	var mm = sec%60
	var hh = parseInt(sec/60)
	if (hh > 0) {
		return [
			hh,
			("0"+mm).substr(-2),
			("0"+ss).substr(-2)
		].join(":")
	}
	return [
		("0"+mm).substr(-2),
		("0"+ss).substr(-2)
	].join(":")
}

Date.prototype.yyyymmdd = function() {
};

function menuOpen(tile) {
	$("#alertUI").hide()
	$(".tooltip").remove()
	if (!IsTile(tile)) {
		return
	}
	deleteMenu();
	if (tile.obj.headTile) {
		addMenu(tile.obj.headTile, MENU["lv"+tile.obj.headTile.obj.level]);
	} else {
		addMenu(tile, MENU["lv"+tile.obj.level]);
	}
	message("menu open x : " + tile.x + " y : " + tile.y );
	var $menu = $("#menu");
	$menu[0].target = tile;
	tile.SelectTile();
	if (tile.obj.headTile) {
		tile.obj.headTile.touch.append($menu.show())
	} else {
		tile.touch.append($menu.show())
	}
}

function TileUI(tile) {
	this.Tile = tile
}

TileUI.prototype.Hover = function () {
	$(".hover").removeClass("hover");
	this.Tile.touch.addClass("hover");
}

TileUI.prototype.SelectTile = function () {
	$(".selected").removeClass("selected");
	if (this.Tile.obj.headTile) {
		var headTile = this.Tile.obj.headTile;
		headTile.touch.addClass("selected");
		var o = {x:headTile.x,y:headTile.y}
		for (var i = 0 ; i < 3 ; i++) {
			directByNum(o, i)
			var t = gGame.tiles[o.x + o.y * gConfig.Size];
			t.touch.addClass("selected");
		}
	} else {
		this.Tile.touch.addClass("selected");
	}
}
TileUI.prototype.SelectUpperLvTile = function(lv) {
	$(".selected").removeClass("selected");
	this.Tile.touch.addClass("selected");

	var checker = this.Tile.CheckLvRound(lv);
	if (checker.CheckLvF()) {
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			message(checker.candidate[i]);
			checker.candidate[i].touch.addClass("selected").addClass("hover");
		}
	} else {
		this.Tile.touch.addClass("selected");
	}
}
TileUI.prototype.startBuild = function() {
	this.Tile.obj.BuildProcessing = true
}
TileUI.prototype.BuildUp = function() {
	this.startBuild()
	var targetTile = this.Tile;
	this.Tile.touch.find(".hoverArea").addClass(this.Tile.TypeName());
	this.Tile.obj.find("img.floor").attr("src", "/images/tile/building_floor.png");
	if (this.Tile.obj.level < 5) {
		for (var i = 0 ; i < this.Tile.obj.level ; i++) {
			if (i == this.Tile.obj.level-1) {
				var $img = $("<img class='building lv"+(this.Tile.obj.level)+"' src='/images/building/construction.png'/>")
			} else {
				var $img = $("<img class='building lv"+(i+1)+"' src='/images/building/"+this.Tile.TypeName()+"_lv1.png'/>")
			}
			this.Tile.obj.append($img);
		}
	} else if (this.Tile.obj.level == 5) {
		this.Tile.obj.find(".building").detach();
		var $img = $("<img class='building lv5' src='/images/building/construction.png'/>")
		this.Tile.obj.append($img);
	} else if (this.Tile.obj.level == 6) {
		var checker = this.Tile.CheckLvRound(6);
		if (checker.CheckLvF()) {// buildable lvF
			var headTile = gGame.tiles[checker.maxCoordinate];
			for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
				var tile = checker.candidate[i];
				tile.obj.find(".building").detach();
				tile.obj.headTile = headTile;
				tile.UI.startBuild()
			}
			message("fleta!! " + tile.x + " : " + tile.y);

			var $img = $("<img class='building lv6' src='/images/building/construction.png'/>")
			headTile.obj.append($img);
			targetTile = headTile
		}
	}

	return targetTile;
}

TileUI.prototype.completBuilding = function (lv) {
	if (lv == 6) {
		var checker = this.Tile.obj.headTile.CheckLvRound(6)
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			var tile = checker.candidate[i];
			tile.obj.BuildProcessing = false;
			tile.obj.level = lv;
			if (tile.index !== tile.obj.headTile.index) {
				tile.obj.level = 0;
			}
		}
		this.Tile.obj.headTile.obj.find("img.floor").attr("src", "/images/tile/"+this.Tile.TypeName()+"_LvFLETA-Tile.png").addClass("lv6");
		var fileTail = "_LvFLETA"
	} else  {
		this.Tile.obj.level = lv
		this.Tile.obj.BuildProcessing = false
		if (this.Tile.obj.level < 5) {
			var fileTail = "_Lv1"
		} else {
			var fileTail = "_Lv5"
		}
	}
	this.Tile.obj.find(".lv"+lv+".building").attr("src", "/images/building/"+this.Tile.TypeName()+""+fileTail+".png")
	this.constructEffect()

	var $menu = $("#menu");
	if ($menu[0].target == this.Tile) {
		menuOpen(this.Tile)
	}
}

TileUI.prototype.distructionEffect = function(callback) {
	this.buildEffect("distructionEffect", callback)
}

TileUI.prototype.constructEffect = function(callback) {
	this.buildEffect("constructEffect", callback)
}

TileUI.prototype.buildEffect = function(type, callback) {
	if (this.Tile.obj.headTile) {
		var tile = this.Tile.obj.headTile
	} else {
		var tile = this.Tile
	}

	var effect = $("<div class='lv"+tile.obj.level+" buildEffect "+type+"'/>")
	tile.touch.append(effect);
	(function (effect, tile, callback) {
		setTimeout(function () {
			if (typeof callback === "function") {
				callback(tile.UI)
			}
			effect.remove()
		}, 1500)
	})(effect, tile, callback)

	if (tile.obj.level == 6 && type == "constructEffect") {
		tile.UI.fletaEffect()
	}
}

TileUI.prototype.fletaEffect = function() {
	var effect = $("<div class='buildEffect FLETAAnimation'/>")
	this.Tile.touch.append(effect);
	(function (effect) {
		setTimeout(function () {
			effect.remove()
		}, 3000)
	})(effect)
}

function newTouchDiv(index) {
	return $("<div tileindex="+index+"/>")
		.append($("<div class='scaleArea'><div class='hoverArea'/></div>"))
}

function newObjDiv(x,y,num) {
	return $("<div/>")
		.css("z-index", x*gConfig.Size+y)
		.append($("<img class='floor' src='/images/tile/base_floor/groundtiles_tile"+num+".png'>"))
}

var UIAlert = {}

UIAlert.Alert = function(btn, okFunc, cancelFunc) {
	UIAlert.btn = $("#"+btn);
	UIAlert.okFunc = okFunc;
	UIAlert.cancelFunc = cancelFunc;
	UIAlert.show()
}

UIAlert.show = function () {
	var $touch = $("#menu").parent()
	if (typeof UIAlert.alertUI === "undefined") {
		UIAlert.alertUI = $("#alertUI")
	}
	UIAlert.alertUI.attr("class", UIAlert.btn.attr("id"))
	$touch.append(UIAlert.alertUI)
	UIAlert.alertUI.show()
}

UIAlert.okOnclick = function () {
	UIAlert.okFunc()
	UIAlert.alertUI.hide()
	menuClose()
};

UIAlert.cancelOnclick = function () {
	if (typeof UIAlert.cancelFunc === "function") {
		UIAlert.cancelFunc()
	}
	UIAlert.alertUI.hide()
};
