function Tile(x, y, type, level, build_height) {
	this.x = x;
	this.y = y;
    this.index = x+y*gConfig.Size;
	this.level = level||0;
	this.build_height = build_height||0;
	this.type = type||null;
}

Tile.prototype.init = function (tileUI) {
	this.UI = tileUI
	this.UI.init(this)
	
	if (this.level > 0 && this.level <= 6) {
		var checker = this.CheckLvRound(6)
		var type = this.type
		this.level--
		var level = this.level
		checker.CheckLvF(function (t, headT) {
			t.level = level
			t.type = type
		})
		this.UI.BuildUp()
		if(this.build_height + gGame.define_map[this.type][this.level].build_time*2 <= gGame.height) {
			this.completBuilding("noEffect")
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

Tile.prototype._remove = function() {
	this.UI.Remove()
	this.level = 0;

	delete this.type;
	delete this.headTile;
	delete this.candidate;
	delete this.BuildProcessing;
	return this;
}

Tile.prototype.Remove = function() {
	if (this.headTile) {
		var c = this.headTile.candidate
		for (var i in c) {
			c[i]._remove();
		}
	} else {
		this._remove();
	}
	menuClose();
	return this;
}

Tile.prototype.ValidateBuild = function() {
	var able = buildableResource(this);
	if (able !== true) {
		return false;
	}

	if (this.BuildProcessing == true) {
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
	if (this.BuildProcessing != true) {
		var temp = this.type
		this.type = type||this.type
		if (this.ValidateBuild()) {
			var checker = this.CheckLvRound();
			if (checker.CheckLvF(function (t, headT) {
				t.build_height_old = t.build_height;
				t.build_height = undefined;
			}) == false) {
				this.build_height_old = this.build_height;
				this.build_height = undefined;
			}
	
			var ret = this.UI.BuildUp();
			return ret;
		} else {
			this.type = temp;
		}
		return false
	}
};

Tile.prototype.completBuilding = function (effect) {
	var checker = this.CheckLvRound(5)
	if (checker.CheckLvF(function (t, headT) {
		t.level++;
	}) == false ) {
		this.level++;
	}
	this.UI.completBuilding(effect)
}

Tile.prototype.CheckLvRound = function(checkLv) {
	var lt = new LvFTiles()
	return lt.CheckLvRound(this, checkLv)
};

function LvFTiles () {
	this.init()
}
function IsLvFTiles(lvFTiles) {
	if (typeof lvFTiles === "undefined") {
		return false;
	}
	return lvFTiles.Symbol == LvFTiles.Symbol;
}
LvFTiles.Symbol = Symbol("LvFTiles");
LvFTiles.prototype.Symbol = LvFTiles.Symbol;
LvFTiles.prototype.init = function() {
	this.maxCoordinate = 0;
	this.candidate = [];
	this.indexer = {};
	this.headTile = null;
}
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
LvFTiles.prototype.CheckLvF = function(callback) {
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
				this.Is = true
			}
		}
	}
	if (this.Is == true) {
		if (typeof callback === "function") {
			var headTile = gGame.tiles[this.maxCoordinate]
			for (var i in this.candidate) {
				this.candidate[i].headTile = headTile
				callback(this.candidate[i], headTile)
			}
			headTile.candidate = this.candidate
		}
	}
	return this.Is
}
LvFTiles.prototype.CheckLvRound = function(spot, checkLv) {
	if (typeof spot === "undefined") {
		throw "is not checkable type"
	}
	if (typeof spot.x === "undefined") {
		throw "is not checkable type"
	}
	if (typeof spot.y === "undefined") {
		throw "is not checkable type"
	}
	if (typeof checkLv === "undefined") {
		checkLv = 5;
	}
	if (checkLv == 6) {
		var _tile = gGame.tiles[spot.x + spot.y *gConfig.Size];
		if (_tile.headTile) {
			spot = {x : _tile.headTile.x, y : _tile.headTile.y};
		}
	}
	var tile = gGame.tiles[spot.x + spot.y *gConfig.Size];

	var type = tile.type;
	if (tile.level != checkLv) {
		return this;
	}
	var o = {x:tile.x, y: tile.y};
	for ( var i = 0 ; i < 4 ; i++ ) {
		if (tile.level == checkLv && type == tile.type) {
			for ( var j = i ; j < i+4 ; j++ ) {
				directByNum(o, j%4);
				if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
					var tile = gGame.tiles[o.x + o.y * gConfig.Size];
					if (typeof tile !== "undefined") {
						if (tile.level == checkLv && type == tile.type) {
							this.PutCandidate(tile)
						}
					}
				}
			}
		}
		if (this.CheckLvF() == true) {
			break;
		}
		this.init()
	}
	return this;
}
