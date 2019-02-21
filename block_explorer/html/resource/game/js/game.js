function initGame () {
    ChangeUnit(gConfig.Unit)
    var jScreen = $("#tileCase");
    jScreen.css("width", (gConfig.Size)+"rem");
    jScreen.css("height", (gConfig.Size)/2+"rem");

	gBuildingDefine = DefineMap
	gGame.define_map = gBuildingDefine
	gGame.height = BlockHeight
	loadTile(UserTiles)
	scoreReloader()
}

function scoreReloader() {
	scoreReloader.obj = setInterval(function () {
		gGame.height++;
		updateResource(gGame.Update());
	}, 500)
}

function updateResource(resource) {
	if (typeof resource === "string") {
		resource = JSON.parse(resource)
	}
	var $scoreBoard = $("#scoreboard")
	for (var key in resource) {
		var $board = $scoreBoard.find("[key='"+key+"']")
		if ($board.length > 0) {
			if (key == "add_balance") {
				$board.html("(+"+resource[key]+"/s)")
			} else {
				$board.html(toShotUnit(resource[key]))
			}
		}
	}

}

function loadTile(tiles) {
	var $touchpad = $("#touchpad");
	var jScreen = $("#screen");

	gConfig.Size = Math.pow(tiles.length, 0.5);
	for(var i=0; i<tiles.length; i++) {
		var x = i%gConfig.Size;
		var y = parseInt(i/gConfig.Size);

		var num = getNum(x, y)
		if (tiles[i]) {
			var tile = new Tile(jScreen, $touchpad, x, y, num, tiles[i].area_type, tiles[i].level , tiles[i].build_height)
		} else {
			var tile = new Tile(jScreen, $touchpad, x, y, num)
		}
		gGame.tiles.push(tile);
		tile.init()
	}

}

function Tile(jScreen, $touchpad, x, y, num, type, level, build_height) {
	this.x = x;
	this.y = y;
	this.index = x+y*gConfig.Size;
	newTouchDiv()
	this.touch = newTouchDiv(this.index)
	$touchpad.append(this.touch)
	this.num = num
	this.obj = newObjDiv(x, y, this.num);
	jScreen.append(this.obj)
	this.level = level||0;
	this.build_height = build_height||0;
	this.type = type||null;
}

Tile.prototype.init = function () {
	this.Resize();
	this.UI = new TileUI(this)
	if (this.level > 0) {
		if (this.level <= 6) {
			if (this.level == 6) {
				var o = {x:this.x,y:this.y}
				this.obj.headTile = this
				for (var i = 0 ; i < 3 ; i++) {
					directByNum(o, i)
					var t = gGame.tiles[o.x + o.y * gConfig.Size];
					t.obj.headTile = this
					t.level = 5
					t.type = this.type
				}
			}
			this.level--
			this.UI.BuildUp(this.level+1)
			if(this.build_height + gGame.define_map[this.type][this.level].build_time*2 <= gGame.height) {
				this.UI.completBuilding(this.level+1, "noEffect")
			}
		}
	}
}


function IsTile(tile) {
	if (typeof tile === "undefined") {
		return false;
	}
	return tile.Symbol == Tile.Symbol;
}

Tile.Symbol = Symbol("Tile");
Tile.prototype.Symbol = Tile.Symbol;
Tile.prototype.TypeName = function() {
	switch(this.type) {
	case CommercialType:
		return "Commercial";
	case IndustrialType:
		return "Industrial";
	case ResidentialType:
		return "Residential";
	default:
		return "empty";
	}
}

Tile.prototype.Hover = function() {
	this.UI.Hover();
	printInfo(this.x, this.y);
	return this;
}

Tile.prototype.SelectTile = function() {
	this.UI.SelectTile();
	return this;
}

function LvFTiles () {
	this.maxCoordinate = 0;
	this.candidate = [];
	this.indexer = {};
	this.headTile = null;
}
function IsLvFTiles(lvFTiles) {
	if (typeof lvFTiles === "undefined") {
		return false;
	}
	return lvFTiles.Symbol == LvFTiles.Symbol;
}
LvFTiles.Symbol = Symbol("LvFTiles");
LvFTiles.prototype.Symbol = LvFTiles.Symbol;
LvFTiles.prototype.PutCandidate = function(tile) {
	if (!IsTile(tile)) {
		throw "is not Tile";
	}
	if (typeof this.level === "undefined") {
		this.level = tile.level;
	} else if (this.level != tile.level) {
		return false;
	}
	if (typeof this.type === "undefined") {
		this.type = tile.type;
	} else if (this.type != tile.type) {
		return false;
	}

	if (typeof this.indexer[tile.x] === "undefined") {
		this.indexer[tile.x] = {};
	}
	this.indexer[tile.x][tile.y] = tile.x + tile.y*gConfig.Size;
	if (this.maxCoordinate < this.indexer[tile.x][tile.y]) {
		this.maxCoordinate = this.indexer[tile.x][tile.y];
	}
	this.Is = undefined;
	this.candidate.push(tile);
}
LvFTiles.prototype.CheckLvF = function() {
	if (typeof this.Is === "undefined") {
		this.Is = false
		if (this.candidate.length == 4) {
			var x = [], y = [];
			for (var k in this.indexer) {
				x.push(k);
				for (var k2 in this.indexer[k]) {
					if (y.indexOf(k2) < 0) {
						y.push(k2);
					}
				}
			}
			if ((x.length == 2 && y.length == 2) && 
				(Math.abs(x[0]-x[1]) == 1 && Math.abs(y[0]-y[1]) == 1)) {
				this.headTile = gGame.tiles[this.maxCoordinate]
				this.Is = true
			}
		}
	}
	return this.Is
}

Tile.prototype.CheckLvRound = function(checkLv) {
	if (typeof checkLv === "undefined") {
		checkLv = 5;
	}
	if (checkLv == 6) {
		var o = {x : this.obj.headTile.x, y : this.obj.headTile.y};
	} else {
		var o = {x : this.x, y : this.y};
	}

	var tile = gGame.tiles[o.x + o.y *gConfig.Size];
	var type = tile.type;
	var checker = new LvFTiles();
	if (tile.level != checkLv) {
		return checker;
	}
	for ( var i = 0 ; i < 4 ; i++ ) {
		var tile = gGame.tiles[o.x + o.y * gConfig.Size];

		if (tile.level == checkLv && type == tile.type) {
			for ( var j = i ; j < i+4 ; j++ ) {
				directByNum(o, j%4);
				if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
					var tile = gGame.tiles[o.x + o.y * gConfig.Size];
					if (typeof tile !== "undefined") {
						if (tile.level == checkLv && type == tile.type) {
							checker.PutCandidate(tile)
						}
					}
				}
			}
		}
		if (checker.CheckLvF() == true) {
			break;
		}
		checker = new LvFTiles()
	}
	return checker;
}

Tile.prototype._remove = function() {
	this.obj.find(".building").detach();
	this.level = 0;
	this.touch.find(".hoverArea").attr("class", "hoverArea");
	this.obj.find(".floor").attr("src", "/game/images/tile/base_floor/groundtiles_tile"+this.num+".png").attr("class", "floor");
	this.obj.css("z-index", this.x*gConfig.Size+this.y)

	delete this.type;
	delete this.obj.headTile;
	delete this.obj.BuildProcessing;
	return this;
}

Tile.prototype.Remove = function() {
	if (this.level == 6) {
		var headTile = this.obj.headTile
		var o = {x:headTile.x,y:headTile.y}
		for (var i = 0 ; i < 3 ; i++) {
			directByNum(o, i)
			var t = gGame.tiles[o.x + o.y * gConfig.Size];
			t._remove();
		}
	}
	this._remove();
	menuClose();
	return this;
}

Tile.prototype.ValidateBuild = function() {
	var able = buildableResource(this);
	if (able !== true) {
		return false;
	}

	if (this.obj.BuildProcessing == true) {
		message("It is not possible to build on a tile under construction.")
		return false;
	}

	if (this.level == 5) {
		var checker = this.CheckLvRound();
		if (!checker.CheckLvF()) {
			return false;
		}
	}

	return true;
}

Tile.prototype.Build = function(type) {
	this.type = type||this.type
	if (this.ValidateBuild()) {
		if (this.level == 5) {
			var checker = this.CheckLvRound();
			var headTile = gGame.tiles[checker.maxCoordinate];
			for (var i = 0 ; i < checker.candidate.length; i++) {
				var t = checker.candidate[i];
				// t.level = 6;
				t.obj.headTile = headTile;
				t.build_height_old = t.build_height;
				t.build_height = gGame.height;
			}
		} else {
			// this.level++;
			this.build_height_old = this.build_height;
			this.build_height = gGame.height;
		}

		var ret = this.UI.BuildUp(this.level+1);
		return ret;
	}
	return false
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
	h[i++] = ".island{width:"+(gConfig.Size*1.12625)+"rem;height:"+(gConfig.Size*0.84875)+"rem}"
	h[i++] = "#tileCase{top:"+(gConfig.Size*0.251875)+"rem;left:"+(gConfig.Size*0.0625)+"rem}"

	$("#cssControll").html(h.join("\n"));
	$("html").css("font-size", gConfig.Unit+"px");
}
