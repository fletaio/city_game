function menuClose () {
	$(".selected").removeClass("selected");
	$(".hover").removeClass("hover");
	$menu = $("#menu");
	delete $menu[0].target;
	$menu.hide();
	deleteMenu();
	$("#cover").hide();
}

function deleteMenu() {
    $("#menu").html("")
}

function addMenu(funcs) {
    for (var key in funcs) {
        var btn = $("<button id=\""+funcs[key]+"\" onclick=\"$('#menu')[0].target.RunCommand('"+funcs[key]+"')\" value=\""+key+"\">"+key+"</button>")
        $("#menu").append(btn)
    }
}

function menuOpen(tile) {
	if (!IsTile(tile)) {
		return
	}
	deleteMenu();
	addMenu(MENU["lv"+tile.obj.level]);
	message("menu open x : " + tile.x + " y : " + tile.y );
	$("#cover").show();
	var $menu = $("#menu");
	$menu[0].target = tile;
	tile.SelectTile();
	$menu.show();
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
	if (this.Tile.obj.level == 6) {
		var head = this.Tile.obj.headTile;
		head.UI.SelectUpperLvTile(6);
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
TileUI.prototype.BuildUpLv1 = function(type) {
	if (this.Tile.obj.find(".lv1").length == 0) {
		this.startBuild()
		this.Tile.touch.find(".hoverArea").addClass(this.Tile.Type);
		this.Tile.obj.find("img.floor").attr("src", "/images/tile/building_floor.png");
		var $img = $("<img class='building lv1' src='/images/building/construction.png'/>")
		this.Tile.obj.append($img);
		(function (tile) {
			setTimeout(function () {
				tile.completBuilding("lv1")
				// $img.attr("src", "/images/building/"+type+"_Lv1.png")
			}, buildingTime.lv1)
		})(this)
	}
}
TileUI.prototype.BuildUpLv2 = function(type) {
	this.BuildUpLv1(type)
	if (this.Tile.obj.find(".lv2").length == 0) {
		this.startBuild()
		var $img = $("<img class='building lv2' src='/images/building/construction.png'/>")
		this.Tile.obj.append($img);
		(function (tile) {
			setTimeout(function () {
				tile.completBuilding("lv2")
				// $img.attr("src", "/images/building/"+type+"_Lv1.png")
			}, buildingTime.lv2)
		})(this)
	}
}
TileUI.prototype.BuildUpLv3 = function(type) {
	this.BuildUpLv2(type)
	if (this.Tile.obj.find(".lv3").length == 0) {
		this.startBuild()
		var $img = $("<img class='building lv3' src='/images/building/construction.png'/>")
		this.Tile.obj.append($img);
		(function (tile) {
			setTimeout(function () {
				tile.completBuilding("lv3")
				// $img.attr("src", "/images/building/"+type+"_Lv1.png")
			}, buildingTime.lv3)
		})(this)
	}
}
TileUI.prototype.BuildUpLv4 = function(type) {
	this.BuildUpLv3(type)
	if (this.Tile.obj.find(".lv4").length == 0) {
		this.startBuild()
		var $img = $("<img class='building lv4' src='/images/building/construction.png'/>")
		this.Tile.obj.append($img);
		(function (tile) {
			setTimeout(function () {
				tile.completBuilding("lv4")
				// $img.attr("src", "/images/building/"+type+"_Lv1.png")
			}, buildingTime.lv4)
		})(this)
	}
}
TileUI.prototype.BuildUpLv5 = function(type) {
	this.startBuild()
	this.Tile.obj.find(".building").detach();
	var $img = $("<img class='building lv5' src='/images/building/construction.png'/>")
	this.Tile.obj.append($img);
	(function (tile) {
		setTimeout(function () {
			tile.completBuilding("lv5")
			// $img.attr("src", "/images/building/"+type+"_Lv5.png")
		}, buildingTime.lv5)
	})(this)
}
TileUI.prototype.BuildUpLv6 = function(checker, type) {
	this.startBuild()
	var headTile = Tiles[checker.maxCoordinate];
	for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
		var tile = checker.candidate[i];
		tile.obj.find(".building").detach();
		tile.obj.headTile = headTile;
		tile.UI.startBuild()
	}

	message("fleta!! " + tile.x + " : " + tile.y);

	headTile.obj.find("img.floor").attr("src", "/images/tile/"+headTile.Type+"_LvFLETA-Tile.png").addClass("lvF");
	var $img = $("<img class='building lvF' src='/images/building/construction.png'/>")
	headTile.obj.append($img);
	(function (headTile) {
		setTimeout(function () {
			headTile.completBuilding("lvF")
		}, buildingTime.lv5)
	})(headTile.UI)
}

TileUI.prototype.completBuilding = function (lv) {
	if (this.Tile.obj.level == 5) {
		var checker = this.Tile.obj.headTile.CheckLvRound()
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			var tile = checker.candidate[i];
			tile.obj.level = 6;
			tile.UpdateInfo();
		}
		var fileTail = "_LvFLETA"
	} else  {
		this.Tile.obj.level++
		if (this.Tile.obj.level < 5) {
			var fileTail = "_Lv1"
		} else {
			var fileTail = "_Lv5"
		}
	}
	this.Tile.obj.find("."+lv+".building").attr("src", "/images/building/"+this.Tile.Type+""+fileTail+".png")
	this.Tile.obj.BuildProcessing = false
	this.Tile.UpdateInfo();

	var $menu = $("#menu");
	if ($menu[0].target == this.Tile) {
		menuOpen(this.Tile)
	}
}

function newTouchDiv() {
	return $("<div/>")
		.append($("<div class='scaleArea'><div class='hoverArea'/></div>"))
		.append($("<span/>"))
}

function newObjDiv(x,y,num) {
	return $("<div/>")
		.css("z-index", x+y)
		.append($("<img class='floor' src='/images/tile/base_floor/groundtiles_tile"+num+".png'>"))
}