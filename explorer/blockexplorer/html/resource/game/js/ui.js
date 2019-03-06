function menuClose () {
	UIAlert.hide()
	$("#selectedInfo").hide()
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
			btn.attr("onclick", "event.stopPropagation();Alert('"+able+"');")
		} else {
			btn.attr("onclick", "event.stopPropagation();$('#menu')[0].target.RunCommand('"+funcs[key]+"');")
		}
		var $tooltip = $("#tooltip").clone();
		var $this = btn
		$tooltip.removeAttr("id")
		$tooltip.attr("class", "tooltip " + funcs[key])

		if (funcs[key] !== "Demolition") {
			if (tile.level == 0) {
				var type = buildingNum($this.attr("id"))
			} else {
				var type = tile.type
			}
			var r = gBuildingDefine[type][tile.level]
			$tooltip.find("#needDollar").html(toShotUnit(r.cost_usage))
			$tooltip.find("#needPower").html(toShotUnit(r.power_usage))
			$tooltip.find("#needDemographic").html(toShotUnit(r.man_usage))
			$tooltip.find("#resource").attr("class", buildingType(type)).html("+"+toShotUnit(r.output)+"/s").attr("class", buildingType(type))

			var time = secondToDate(r.build_time);
		} else {
			var r = gBuildingDefine[tile.type][tile.level-1]
			$tooltip.find("#needDollar").html("+"+toShotUnit(r.cost_usage/2))
			if (tile.type == IndustrialType) {
				$tooltip.find("#needPower").html("-"+toShotUnit(r.output))
			} else {
				$tooltip.find("#needPower").html("+"+toShotUnit(r.power_usage))
			}
			if (tile.type == ResidentialType) {
				$tooltip.find("#needDemographic").html("-"+toShotUnit(r.output))
			} else {
				$tooltip.find("#needDemographic").html("+"+toShotUnit(r.man_usage))
			}
			$tooltip.find("#resource").remove()

			var time = "1s";
		}
		$tooltip.find("#needTime").html(time).attr("class", "")

		$("#menu").append($tooltip)
		$("#menu").append(btn)
    }
}

function secondToDate(time) {
	time = parseInt(time)
	var ss = time%60
	time = parseInt((time)/60)
	var mm = time%60
	if (time > 60 && ss > 0) {
		mm++
	}
	var hh = parseInt(time/60)
	var r = ""
	if (hh > 0) {
		r += hh+"h"
		if (mm != 0) {
			r += mm+"m"
		}
	} else if (mm > 0) {
		r += mm+"m"
	} else {
		r += ss+"s"
	}
	// r += ("0"+mm).substr(-2)+"m"
	return r
}

function menuOpen(tile) {
	UIAlert.hide()
	$(".tooltip").remove()
	if (tile.type) {
		var $selectedInfo = $("#selectedInfo")
		$selectedInfo.attr("class", buildingType(tile.type)).show()
		$selectedInfo.find(".building_type").html(buildingType(tile.type))
		var lv = tile.level
		if (tile.obj.headTile) {
			lv = tile.obj.headTile.level
		}
		if (lv == 6) {
			$selectedInfo.find(".building_level").html("lvFLETA")
		} else {
			$selectedInfo.find(".building_level").html("lv"+tile.level)
		}
		
		if (lv > 0) {
			$selectedInfo.find(".resource").html("+"+toShotUnit(gGame.define_map[tile.type][lv-1].output))
		} else {
			$selectedInfo.find(".resource").html("under construction")
		}
	} else {
		$("#selectedInfo").hide()
	}

	if (!IsTile(tile)) {
		return
	}
	deleteMenu();
	if (tile.obj.headTile) {
		addMenu(tile.obj.headTile, MENU["lv"+tile.obj.headTile.level]);
	} else {
		addMenu(tile, MENU["lv"+tile.level]);
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
TileUI.prototype.BuildUp = function(lv) {
	this.startBuild()
	var targetTile = this.Tile;
	this.Tile.touch.find(".hoverArea").addClass(this.Tile.TypeName());
	this.Tile.obj.find("img.floor").attr("src", "/game/images/tile/building_floor.png");
	if (lv < 5) {
		for (var i = 0 ; i < lv ; i++) {
			if (i == lv-1) {
				var $img = $("<img class='building lv"+(lv)+"' src='/game/images/building/construction.png'/>")
			} else {
				var $img = $("<img class='building lv"+(i+1)+"' src='/game/images/building/"+this.Tile.TypeName()+"_Lv1.png'/>")
			}
			this.Tile.obj.append($img);
		}
	} else if (lv == 5) {
		this.Tile.obj.find(".building").detach();
		var $img = $("<img class='building lv5' src='/game/images/building/construction.png'/>")
		this.Tile.obj.append($img);
	} else if (lv == 6) {
		var checker = this.Tile.CheckLvRound(5);
		if (checker.CheckLvF()) {// buildable lvF
			var headTile = gGame.tiles[checker.maxCoordinate];
			for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
				var tile = checker.candidate[i];
				tile.obj.find(".building").detach();
				tile.obj.headTile = headTile;
				tile.UI.startBuild()
			}
			message("fleta!! " + tile.x + " : " + tile.y);

			var $img = $("<img class='building lv6' src='/game/images/building/construction.png'/>")
			headTile.obj.append($img);
			targetTile = headTile
		}
	}

	return targetTile;
}

TileUI.prototype.completBuilding = function (lv, effect) {
	if (lv == 6) {
		var checker = this.Tile.obj.headTile.CheckLvRound(5)
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			var tile = checker.candidate[i];
			tile.obj.BuildProcessing = false;
			if (tile.index == tile.obj.headTile.index) {
				tile.level = 6;
			} else {
				tile.level = 0;
			}
			var p = tile.obj
			p.css("z-index", p.css("z-index")-1-gConfig.Size)
		}
		this.Tile.obj.headTile.obj.find("img.floor").attr("src", "/game/images/tile/"+this.Tile.TypeName()+"_LvFLETA-Tile.png").addClass("lv6");
		var fileTail = "_LvFLETA"
	} else  {
		this.Tile.level = lv
		this.Tile.obj.BuildProcessing = false
		if (this.Tile.level < 5) {
			var fileTail = "_Lv1"
		} else {
			var fileTail = "_Lv5"
		}
	}
	this.Tile.obj.find(".lv"+lv+".building").attr("src", "/game/images/building/"+this.Tile.TypeName()+""+fileTail+".png")
	if (effect != "noEffect") {
		this.constructEffect()
	}
	this.Tile.touch.find(".underconstruction").remove()


	var $menu = $("#menu");
	if ($menu[0].target == this.Tile) {
		menuOpen(this.Tile)
	}
}

TileUI.prototype.ShowBuildProcessingTime = function(sec) {
	var lv = this.Tile.level
	var uc = this.Tile.touch.find(".underconstruction")
	if (uc.length == 0) {
		uc = $("<div class='lv"+lv+" underconstruction'>")
		this.Tile.touch.append(uc)
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
	if (this.Tile.obj.headTile) {
		var tile = this.Tile.obj.headTile
	} else {
		var tile = this.Tile
	}

	var effect = $("<div class='lv"+tile.level+" buildEffect "+type+"'/>")
	tile.touch.append(effect);
	(function (effect, tile, callback) {
		setTimeout(function () {
			if (typeof callback === "function") {
				callback(tile.UI)
			}
			effect.remove()
		}, 1500)
	})(effect, tile, callback)

	tile.UI.completEffect()
}

TileUI.prototype.fletaEffect = function() {
	var effect = $("<div class='FLETAAnimation lv"+this.Tile.level+"'/>")
	this.Tile.touch.append(effect);
	(function (effect) {
		setTimeout(function () {
			effect.remove()
		}, 3000)
	})(effect)
}

TileUI.prototype.completEffect = function() {
	var effect = $("<div class='completAnimation lv"+this.Tile.level+"'/>")
	this.Tile.touch.append(effect);
	(function (effect, tile) {
		setTimeout(function () {
			effect.remove()
			if (tile.level == 6) {
				tile.UI.fletaEffect()
			}
		}, 3000)
	})(effect, this.Tile)
}

function newTouchDiv(index) {
	return $("<div tileindex="+index+"/>")
		.append($("<div class='scaleArea'><div class='hoverArea'/></div>"))
}

function newObjDiv(x,y,num) {
	return $("<div/>")
		.css("z-index", x*gConfig.Size+y)
		.append($("<img class='floor' src='/game/images/tile/base_floor/groundtiles_tile"+num+".png'>"))
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
UIAlert.hide = function () {
	UIAlert.okFunc = null;
	UIAlert.cancelFunc = null;
	if (typeof UIAlert.alertUI === "undefined") {
		UIAlert.alertUI = $("#alertUI")
	}
	UIAlert.alertUI.hide()
}

UIAlert.okOnclick = function () {
	if (UIAlert.okFunc) {
		UIAlert.okFunc()
	}
	UIAlert.hide()
	menuClose()
};

UIAlert.cancelOnclick = function () {
	if (typeof UIAlert.cancelFunc === "function") {
		UIAlert.cancelFunc()
	}
	UIAlert.hide()
};
