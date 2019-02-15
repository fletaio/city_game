function menuClose () {
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
			btn.attr("onclick", "alert('"+able+"')")
		} else {
			btn.attr("onclick", "$('#menu')[0].target.RunCommand('"+funcs[key]+"')")
		}
		$("#menu").append(btn)
    }
}

function menuOpen(tile) {
	if (!IsTile(tile)) {
		return
	}
	deleteMenu();
	addMenu(tile, MENU["lv"+tile.obj.level]);
	message("menu open x : " + tile.x + " y : " + tile.y );
	var $menu = $("#menu");
	$menu[0].target = tile;
	tile.SelectTile();
	var offset = tile.obj.offset()
	// $menu.css("top", offset.top).css("left", offset.left)
	tile.touch.append($menu.show())
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
		headTile = tile
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			var tile = checker.candidate[i];
			tile.obj.level = lv;
			console.log("completBuilding lv 6 " + tile.obj.level)
			tile.obj.BuildProcessing = false;
		}
		this.Tile.obj.headTile.obj.find("img.floor").attr("src", "/images/tile/"+this.Tile.TypeName()+"_LvFLETA-Tile.png").addClass("lv6");
		var fileTail = "_LvFLETA"
	} else  {
		this.Tile.obj.level = lv
		console.log("completBuilding " + this.Tile.obj.level)
		this.Tile.obj.BuildProcessing = false
		if (this.Tile.obj.level < 5) {
			var fileTail = "_Lv1"
		} else {
			var fileTail = "_Lv5"
		}
	}
	this.Tile.obj.find(".lv"+lv+".building").attr("src", "/images/building/"+this.Tile.TypeName()+""+fileTail+".png")

	var $menu = $("#menu");
	if ($menu[0].target == this.Tile) {
		menuOpen(this.Tile)
	}
}

function newTouchDiv(index) {
	return $("<div tileindex="+index+"/>")
		.append($("<div class='scaleArea'><div class='hoverArea'/></div>"))
}

function newObjDiv(x,y,num) {
	return $("<div/>")
		.css("z-index", x+y)
		.append($("<img class='floor' src='/images/tile/base_floor/groundtiles_tile"+num+".png'>"))
}