function initGame () {
    ChangeUnit(gConfig.Unit)
    var jScreen = $("#tileCase");
    jScreen.css("width", (gConfig.Size)+"rem");
    jScreen.css("height", (gConfig.Size)/2+"rem");

	connectToServer(loginInfo.Addr)
	loadTile()
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
			$board.html(resource[key])
		}
	}

}

function loadTile() {
	$.ajax({
		type: "GET",
		url : "/api/games/"+loginInfo.Addr,
		success : function (d) {
            if (typeof d === "string") {
                d = JSON.parse(d)
			}
			console.log("init game")
			console.log(d)
			if (d.define_map) {
				gBuildingDefine = d.define_map;
			}

			var $touchpad = $("#touchpad");
			var jScreen = $("#screen");
		
			gConfig.Size = Math.pow(d.tiles.length, 0.5);
			gGame.define_map = d.define_map;
			gGame.height = d.height;
			gGame.point_height = d.point_height;
			gGame.point_balance = d.point_balance;
			for(var i=0; i<d.tiles.length; i++) {
				var x = i%gConfig.Size;
				var y = parseInt(i/gConfig.Size);

				var num = getNum(x, y)
				if (d.tiles[i]) {
					var tile = new Tile(jScreen, $touchpad, x, y, num, d.tiles[i].area_type, d.tiles[i].level , d.tiles[i].build_height)
				} else {
					var tile = new Tile(jScreen, $touchpad, x, y, num)
				}
				gGame.tiles.push(tile);
				tile.init()
			}

			/*
				"height": HEIGHT_INT,
				"point_height": POINT_HEIGHT_INT,
				"point_balance": POINT_BALANCE_INT,
				"tiles": [{
					"area_type": AREA_TYPE_INT,
					"level": LEVEL_INT,
					"build_height": BUILD_HEIGHT_INT
				}]
			*/
		},
		error: function(d) {
			alert("error")
		}
	})

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
	this.obj.level = level||0;
	this.build_height = build_height||0;
	this.type = type||null;
}

Tile.prototype.init = function () {
	this.Resize();
	this.UI = new TileUI(this)
	if (this.obj.level > 0) {
		if (this.obj.level <= 6) {
			if (this.obj.level == 6) {
				var o = {x:this.x,y:this.y}
				this.obj.headTile = this
				for (var i = 0 ; i < 3 ; i++) {
					directByNum(o, i)
					var t = gGame.tiles[o.x + o.y * gConfig.Size];
					t.obj.headTile = this
					t.obj.level = 6
					t.type = this.type
				}
			}
			this.UI.BuildUp()
			if(this.build_height + gGame.define_map[this.type][this.obj.level-1].build_time*2 <= gGame.height) {
				this.UI.completBuilding(this.obj.level)
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
		this.level = tile.obj.level;
	} else if (this.level != tile.obj.level) {
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
	if (tile.obj.level != checkLv) {
		return checker;
	}
	for ( var i = 0 ; i < 4 ; i++ ) {
		var tile = gGame.tiles[o.x + o.y * gConfig.Size];

		if (tile.obj.level == checkLv && type == tile.type) {
			for ( var j = i ; j < i+4 ; j++ ) {
				directByNum(o, j%4);
				if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
					var tile = gGame.tiles[o.x + o.y * gConfig.Size];
					if (typeof tile !== "undefined") {
						if (tile.obj.level == checkLv && type == tile.type) {
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
	this.obj.level = 0;
	this.touch.find(".hoverArea").attr("class", "hoverArea");
	this.obj.find(".floor").attr("src", "/images/tile/base_floor/groundtiles_tile"+this.num+".png").attr("class", "floor");

	delete this.type;
	delete this.obj.headTile;
	delete this.obj.BuildProcessing;
	return this;
}

Tile.prototype.Remove = function() {
	if (this.obj.level == 6) {
		var checker = this.CheckLvRound(6);
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			checker.candidate[i]._remove();
		}
	} else {
		this._remove();
	}
	menuClose();
	menuOpen(this);
	return this;
}

Tile.prototype.ValidateBuild = function() {
	//TODO check resource

	var able = buildableResource(this);
	if (able !== true) {
		return false;
	}

	if (this.obj.BuildProcessing == true) {
		message("It is not possible to build on a tile under construction.")
		return false;
	}

	if (this.obj.level == 5) {
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
		console.log("build : " + gGame.height)
		if (this.obj.level == 5) {
			var checker = this.CheckLvRound();
			var headTile = gGame.tiles[checker.maxCoordinate];
			for (var i = 0 ; i < checker.candidate.length; i++) {
				var t = checker.candidate[i];
				t.obj.level = 6;
				t.obj.headTile = headTile;
				t.build_height_old = t.build_height;
				t.build_height = gGame.height;
			}
		} else {
			this.obj.level++;
			this.build_height_old = this.build_height;
			this.build_height = gGame.height;
		}

		var ret = this.UI.BuildUp();
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
	h[i++] = ".island{width:"+(gConfig.Size*1.086875)+"rem;height:"+(gConfig.Size*0.805)+"rem}"
	h[i++] = "#tileCase{top:"+(gConfig.Size*0.251875)+"rem;left:"+(gConfig.Size*0.04375)+"rem}"

	$("#cssControll").html(h.join("\n"));

	$("html").css("font-size", gConfig.Unit+"px");
}
