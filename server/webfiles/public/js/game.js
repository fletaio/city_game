function Game(config) {
	// DOM
	this.obj = $(gConfig.Selector.Screen);
	this.touchpad = $(gConfig.Selector.Touchpad);

	// Logic
	this.handlers = [];
	this.tiles = [];
	this.coins = [];
	this.exps = [];
	this.height = 0;
	this.selected_tile = null;
	this.point_height = 0;
	this.point_balance = 0;
	this.loaded_height = 0;
	this.coin_map = {};
	this.exp_map = {};

	// UI
	this.is_menu_on = false;
	this.game_status_ui = new GameStatusUI();
	this.AddHandler(this.game_status_ui);
	this.build_menu_ui = new BuildMenuUI();
	this.AddHandler(this.build_menu_ui);
	this.upgrade_menu_ui = new UpgradeMenuUI();
	this.AddHandler(this.upgrade_menu_ui);

	this.selectedInfo = $("#selectedInfo")

	var _this = this;
	this.touchpad.click(function(e) {
		if (!islandMoved) {
			var posX = e.pageX - $(this).offset().left;
			var posY = e.pageY - $(this).offset().top;
			var width = parseInt($(this).width());
			var height = parseInt($(this).height());
	
			var rem = width/gConfig.Size;
			var a = posX/rem*2 - gConfig.Size;
			var b = posY/rem*2*2;
			var x = Math.floor((a + b)/2);
			var y = Math.floor((-a + b)/2);
			if(0 <= x && x < gConfig.Size) {
				if(0 <= y && y < gConfig.Size) {
					_this.OnTileClicked(x, y);
				}
			}
		}
	});
	this.touchpad.mousemove(function(e) {
		var posX = e.pageX - $(this).offset().left;
		var posY = e.pageY - $(this).offset().top;
		var width = parseInt($(this).width());
		var height = parseInt($(this).height());

		var rem = width/gConfig.Size;
		var a = posX/rem*2 - gConfig.Size;
		var b = posY/rem*2*2;
		var x = Math.floor((a + b)/2);
		var y = Math.floor((-a + b)/2);
		if(0 <= x && x < gConfig.Size) {
			if(0 <= y && y < gConfig.Size) {
				for(var i=0; i<_this.coins.length; i++) {
					var c = _this.coins[i];
					if(c != null && c.x == x && c.y == y && c.height <= _this.height) {
						_this.touchpad.css("cursor", "pointer")
						return;
					}
				}
				for(var i=0; i<_this.exps.length; i++) {
					var e = _this.exps[i];
					if(e.x == x && e.y == y) {
						_this.touchpad.css("cursor", "pointer")
						return;
					}
				}
			
			}
		}
		_this.touchpad.css("cursor", "inherit");

		var t = _this.tiles[x + (y*gConfig.Size)];
		$(".underconstruction.hover").removeClass("hover")
		if (t && t.obj) {
			t.obj.find(".underconstruction").addClass("hover")
		}
	});
}

Game.prototype.AddHandler = function(handler) {
	this.handlers.push(handler);
}

// Always index 0 is fleta target tile
Game.prototype.GetFletaTiles = function(tile) {
	if(tile.level != 5) {
		return null;
	}

	var subtiles = [];
	for(var i=0; i<9; i++) {
		var dx = i%3 - 1;
		var dy = Math.floor(i/3) - 1;
		if(0 <= tile.x + dx && tile.x + dx < gConfig.Size && 0 <= tile.y + dy && tile.y + dy < gConfig.Size) {
			var subtile = this.tiles[tile.x + dx + (tile.y + dy)*gConfig.Size];
			if(subtile.area_type == tile.area_type && subtile.level == 5 && subtile.is_building != true) {
				subtiles.push(subtile);
			} else {
				subtiles.push(null);
			}
		} else {
			subtiles.push(null);
		}
	}
	if(subtiles[0] && subtiles[1] && subtiles[3] && subtiles[4]) {
		return [subtiles[4], subtiles[0], subtiles[1], subtiles[3]];
	}
	if(subtiles[1] && subtiles[2] && subtiles[4] && subtiles[5]) {
		return [subtiles[5], subtiles[1], subtiles[2], subtiles[4]];
	}
	if(subtiles[4] && subtiles[5] && subtiles[7] && subtiles[8]) {
		return [subtiles[8], subtiles[4], subtiles[5], subtiles[7]];
	}
	if(subtiles[3] && subtiles[4] && subtiles[6] && subtiles[7]) {
		return [subtiles[7], subtiles[3], subtiles[4], subtiles[6]];
	}
	return null;
}

Game.prototype.Run = function() {
	var _this = this;
	this.Init(function() {
		ChangeUnit(gConfig.Unit);
		_this.OnHeightUpdated(_this.height);
		setInterval(function() {
			_this.height++;
			_this.OnHeightUpdated(_this.height);
		}, 500);
	});
}

Game.prototype.Init = function(callback) {
	var _this = this;
	connectToServer(Login.address);

	this.Reload(function(d) {
		gConfig.Size = Math.pow(d.tiles.length, 0.5);
		_this.define_map = d.define_map;
		_this.exp_defines = d.exp_defines;
		_this.build_menu_ui.Init(d.define_map);
		_this.txs = d.txs;
		_this.coins = d.coins;
		_this.exps = d.exps;
		_this.coin_count = d.coin_count;
		$("[key='coin_count']").text(d.coin_count);
		_this.total_exp = d.total_exp;
		$("[key='total_exp']").text(d.total_exp);
	
		_this.height = d.height;
		_this.point_height = d.point_height;
		_this.point_balance = d.point_balance;
		_this.loaded_height = d.height;
		_this.coin_count = d.coin_count;
		_this.game_status_ui.OnCoinCountUpdated(d.coin_count);
		_this.total_exp = d.total_exp;
		_this.game_status_ui.OnTotalExpUpdated(d.total_exp);
		for(var i=0; i<gConfig.Size*gConfig.Size; i++) {
			var obj = $("<div objIndex='"+i+"'></div>").appendTo(_this.obj);
			var x = i % gConfig.Size;
			var y = Math.floor(i / gConfig.Size);
			var tile = new Tile(obj, x, y);
			_this.tiles[i] = tile
			var tileData = d.tiles[i];
			if(tileData != null) {
				tile.SetData(d.height, tileData);
			} else {
				tile.removeAllImages();
			}
		}
		for(var i=0; i<_this.coins.length; i++) {
			var c = _this.coins[i];
			if(c != null && c.height <= _this.height) {
				_this.AddCoin(c);
			}
		}
		for(var i=0; i<_this.exps.length; i++) {
			var e = _this.exps[i];
			_this.AddExp(e);
		}
		_this.txs = d.txs;
		if(callback) {
			callback();
		}
	});
}

Game.prototype.Reload = function(callback) {
	$.ajax({
		type: "GET",
		url : "/api/games/"+Login.address,
		success : function (d) {
            if (typeof d === "string") {
                d = JSON.parse(d);
			}
			if(callback) {
				callback(d);
			}
		},
		error: function(d) {
			Alert(language["load fail"])
		}
	})
}

Game.prototype.addressDataProcess = function(d) {
	console.log("reload tile")
	this.txs = d.txs;
	this.coins = d.coins;
	this.exps = d.exps;
	this.coin_count = d.coin_count;
	$("[key='coin_count']").text(d.coin_count);
	this.total_exp = d.total_exp;
	$("[key='total_exp']").text(d.total_exp);

	this.height = d.height;
	this.loaded_height = d.height;
	this.point_height = d.point_height;
	this.point_balance = d.point_balance;
	for(var i=0; i<gConfig.Size*gConfig.Size; i++) {
		var obj = this.obj.find("[objIndex='"+i+"']")
		var x = i % gConfig.Size;
		var y = Math.floor(i / gConfig.Size);
		var tile = new Tile(obj, x, y);
		this.tiles[i] = tile
		var tileData = d.tiles[i];
		if(tileData != null) {
			tile.SetData(d.height, tileData);
		} else {
			tile.removeAllImages();
		}
	}
	for(var i=0; i<this.coins.length; i++) {
		var c = this.coins[i];
		if(c != null && c.height <= this.height) {
			this.AddCoin(c);
		}
	}
	for(var i=0; i<this.exps.length; i++) {
		var e = this.exps[i];
		this.AddExp(e);
	}
	this.txs = d.txs;
	console.log("reload tile end")
}

Game.prototype.AddCoin = function(c) {
	if(this.coin_map[c.index] == null) {
		var obj = $("<div></div>").appendTo(this.obj);
		obj.css("position", "absolute");
		obj.css("left", (gConfig.Size - 1 + c.x - c.y)/2 + "rem");
		obj.css("top", (c.x/2 + c.y/2)/2 + "rem");
		obj.css("width", (gConfig.Size/gConfig.Size) + "rem");
		obj.css("height", (gConfig.Size/2/gConfig.Size) + "rem");
		obj.css("z-index", 10000);
		var $img = $("<div class='fletaCoin'>").appendTo(obj);
		this.coin_map[c.index] = obj;
	}
}

Game.prototype.RemoveCoin = function(c) {
	if(this.coin_map[c.index]) {
		this.coin_map[c.index].detach();
		delete this.coin_map[c.index];
		this.coins[c.index] = null;
	}
}

Game.prototype.AddExp = function(e) {
	if(this.exp_map[e.x + e.y*gConfig.Size] == null) {
		var obj = $("<div></div>").appendTo(this.obj);
		obj.css("position", "absolute");
		obj.css("left", (gConfig.Size - 1 + e.x - e.y)/2 + "rem");
		obj.css("top", (e.x/2 + e.y/2)/2 + "rem");
		obj.css("width", (gConfig.Size/gConfig.Size) + "rem");
		obj.css("height", (gConfig.Size/2/gConfig.Size) + "rem");
		obj.css("z-index", 10000);
		var $img = $("<div class='fletaExp'>").appendTo(obj);
		this.exp_map[e.x + e.y*gConfig.Size] = obj;
	}
}

Game.prototype.RemoveExp = function(e) {
	if(this.exp_map[e.x + e.y*gConfig.Size]) {
		this.exp_map[e.x + e.y*gConfig.Size].detach();
		delete this.exp_map[e.x + e.y*gConfig.Size];
		for(var i=0; i<this.exps.length; i++) {
			if(this.exps[i] == e) {
				this.exps.splice(i, 1);
				break;
			}
		}
	}
}

Game.prototype.UnselectTile = function() {
	if(this.selected_tile != null) {
		this.selected_tile = null;
		this.obj.find(".selected_shown").removeClass("selected_shown").hide();
		hideInfo(this);
	}
}

Game.prototype.CloseMenuUI = function() {
	this.build_menu_ui.Close();
	this.upgrade_menu_ui.Close();
	this.closeSelectedInfo()
}

Game.prototype.UpdateResource = function(target_height) {
	var forward_height = target_height - this.point_height;
	var base = {
		balance: 4,
		add_balance: 4,
		power_remained: 5,
		power_provided: 5,
		man_remained: 3,
		man_provided: 3,
	};
	var used = {
		balance: 0,
		add_balance: 0,
		power_remained: 0,
		power_provided: 0,
		man_remained: 0,
		man_provided: 0
	};
	var provide = {
		balance: this.point_balance + Math.floor(base.balance/2)*forward_height,
		add_balance: Math.floor(base.balance/2),
		power_remained: base.power_remained,
		power_provided: 0,
		man_remained: base.man_remained,
		man_provided: 0
	};
	for(var i=0; i<this.tiles.length; i++) {
		var tile = this.tiles[i];
		var level = tile.level;
		var area_type = tile.area_type;
		var build_height = tile.build_height;
		if(tile.is_pending) {
			level = tile.target_level;
			area_type = tile.target_area_type;
			build_height = target_height;
		}
		if(level > 0) {
			var bd = this.define_map[area_type][level-1];
			used.man_remained += bd.acc_man_usage;
			used.power_remained += bd.acc_power_usage;
			var construction_height = build_height + bd.build_time*2

			if(level == 6) {
				var bd2 = this.define_map[area_type][level-2];
				used.power_remained += bd2.acc_power_usage * 3;
				used.man_remained += bd2.acc_man_usage * 3;
			}

			if(target_height < construction_height) {
				if(level == 1) {
					continue;
				} else if(level == 6) {
					var bd2 = this.define_map[area_type][level-2];
					switch(area_type) {
					case CommercialAreaType:
						provide.add_balance += Math.floor(bd2.output/2) * 3;
						provide.balance += Math.floor(bd2.output/2) * forward_height * 3;
						break;
					case IndustrialAreaType:
						provide.power_remained += bd2.output * 3;
						break;
					case ResidentialAreaType:
						provide.man_remained += bd2.output * 3;
						break;
					}
				}
				bd = this.define_map[area_type][level-2];
			}
			switch(area_type) {
			case CommercialAreaType:
				provide.add_balance += Math.floor(bd.output/2);
				if(construction_height <= this.point_height) {
					provide.balance += Math.floor(bd.output/2) * forward_height;
				} else {
					if(build_height <= this.point_height) {
						if (construction_height > this.point_height && forward_height > (construction_height-this.point_height)) {
							provide.balance += Math.floor(bd.output/2) * (forward_height-(construction_height-this.point_height));
						}
					} else {
						if (target_height > construction_height) {
							provide.balance += Math.floor(bd.output/2) * (target_height-construction_height);
						}
						if(level > 1) {
							var prevbd = this.define_map[tile.area_type][level-2];
							if (build_height > this.point_height) {
								provide.balance += Math.floor(prevbd.output/2) * (build_height-this.point_height);
							}
						}
					}
				}
				break;
			case IndustrialAreaType:
				provide.power_remained += bd.output;
				break;
			case ResidentialAreaType:
				provide.man_remained += bd.output;
				break;
			}
		}
	}
	var resource = {
		balance: provide.balance,
		add_balance: provide.add_balance,
		power_remained: provide.power_remained - used.power_remained,
		power_provided: provide.power_remained,
		man_remained: provide.man_remained - used.man_remained,
		man_provided: provide.man_remained
	};
	for(var i=0; i<this.handlers.length; i++) {
		var handler = this.handlers[i];
		if(handler.OnResourceUpdated != null) {
			handler.OnResourceUpdated(resource);
		}
	}
}

Game.prototype.OnHeightUpdated = function(height) {
	for(var i=0; i<this.tiles.length; i++) {
		this.tiles[i].Update(height);
	}
	for(var i=0; i<this.coins.length; i++) {
		var c = this.coins[i];
		if(c != null && c.height <= height && this.coin_map[c.index] == null) {
			this.AddCoin(c);
		}
	}
	this.UpdateResource(height);
}

Game.prototype.OnCoinClicked = function(c) {
	this.RemoveCoin(c);
	gNetwork.SendTX("getcoin", {
		x: c.x,
		y: c.y,
		index: c.index
	});
}

Game.prototype.OnExpClicked = function(e) {
	this.RemoveExp(e);
	gNetwork.SendTX("getexp", {
		x: e.x,
		y: e.y,
		area_type: e.area_type,
		level: e.level
	});
}

Game.prototype.OnTileClicked = function(x, y) {
	for(var i=0; i<this.coins.length; i++) {
		var c = this.coins[i];
		if(c != null && c.x == x && c.y == y && c.height <= this.height) {
			this.OnCoinClicked(c);
			return;
		}
	}
	for(var i=0; i<this.exps.length; i++) {
		var e = this.exps[i];
		if(e.x == x && e.y == y) {
			this.OnExpClicked(e);
			return;
		}
	}
	var tile = this.tiles[x + y*gConfig.Size];
	if(tile.level == 0) {
		for(var i=1; i<4; i++) {
			var dx = i%2;
			var dy = Math.floor(i/2);
			if(0 <= tile.x + dx && tile.x + dx < gConfig.Size && 0 <= tile.y + dy && tile.y + dy < gConfig.Size) {
				var subtile = this.tiles[tile.x + dx + (tile.y + dy)*gConfig.Size];
				if(subtile.level == 6) {
					tile = subtile;
					break;
				}
			}
		}
	}
	if(!tile.is_pending && !tile.is_building) {
		if(this.selected_tile != tile) {
			this.UnselectTile();
			this.selected_tile = tile;
			if(tile.level == 6) {
				for(var i=0; i<4; i++) {
					var dx = i%2-1;
					var dy = Math.floor(i/2)-1;
					if(0 <= tile.x + dx && tile.x + dx < gConfig.Size && 0 <= tile.y + dy && tile.y + dy < gConfig.Size) {
						var subtile = this.tiles[tile.x + dx + (tile.y + dy)*gConfig.Size];
						subtile.OnSelected(true);
					}
				}
			} else {
				this.selected_tile.OnSelected(true);
			}
			this.is_menu_on = false;
		}
		this.CloseMenuUI();
		if(this.is_menu_on) {
			this.UnselectTile();
			this.is_menu_on = false;
		} else {
			if(tile.level == 0) {
				this.build_menu_ui.Open(tile);
			} else if(tile.level < 6) {
				var subtiles = this.GetFletaTiles(tile);
				if(subtiles != null) {
					for(var i=0; i<subtiles.length; i++) {
						if(tile != subtiles[i]) {
							subtiles[i].OnAllocated();
						}
					}
					this.selected_tile = subtiles[0];
					this.upgrade_menu_ui.Open(subtiles[0]);
				} else {
					this.upgrade_menu_ui.Open(tile);
				}
				this.openSelectedInfo(this.selected_tile)
			} else {
				this.upgrade_menu_ui.Open(tile);
				this.openSelectedInfo(this.selected_tile)
			}
			this.is_menu_on = true;
		}
	}
}

var DemolitionTransactionType = 2;
var ConstructionTransactionType = 3;
var UpgradeTransactionType = 4;
var GetCoinTransactionType = 5;
var GetExpTransactionType = 6;

Game.prototype.OnNotified = function(noti) {
	if(noti.height < this.loaded_height) {
		return;
	}
	this.height = noti.height;
	this.point_height = noti.point_height;
	this.point_balance = noti.point_balance;

	switch(noti.type) {
	case DemolitionTransactionType:
		var tile = this.tiles[noti.x + noti.y*gConfig.Size];
		if(tile.is_pending) {
			if(noti.error.length > 0) {
				tile.CancelPending(noti.height);
			} else {
				tile.SetData(noti.height, {
					build_height: 0,
					area_type: 0,
					level: 0
				});
			}
		}
		break;
	case ConstructionTransactionType:
		var tile = this.tiles[noti.x + noti.y*gConfig.Size];
		if(tile.is_pending) {
			if(noti.error.length > 0) {
				tile.CancelPending(noti.height);
			} else {
				tile.SetData(noti.height, {
					build_height: noti.height,
					area_type: noti.area_type,
					level: 1
				});
				if(noti.exp != null) {
					this.exps.push(noti.exp);
					this.AddExp(noti.exp);
				}
			}
		}
		break;
	case UpgradeTransactionType:
		var tile = this.tiles[noti.x + noti.y*gConfig.Size];
		if(tile.is_pending) {
			if(noti.error.length > 0) {
				if(noti.level == 6) {
					for(var i=0; i<3; i++) {
						var dx = i%2-1;
						var dy = Math.floor(i/2)-1;
						if(0 <= tile.x + dx && tile.x + dx < gConfig.Size && 0 <= tile.y + dy && tile.y + dy < gConfig.Size) {
							var subtile = this.tiles[tile.x + dx + (tile.y + dy)*gConfig.Size];
							subtile.CancelPending(noti.height);
						}
					}
				}
				tile.CancelPending(noti.height);
			} else {
				if(noti.level == 6) {
					for(var i=0; i<3; i++) {
						var dx = i%2-1;
						var dy = Math.floor(i/2)-1;
						if(0 <= tile.x + dx && tile.x + dx < gConfig.Size && 0 <= tile.y + dy && tile.y + dy < gConfig.Size) {
							var subtile = this.tiles[tile.x + dx + (tile.y + dy)*gConfig.Size];
							subtile.SetData(noti.height, {
								build_height: 0,
								area_type: 0,
								level: 0
							});
						}
					}
					tile.SetData(noti.height, {
						build_height: noti.height,
						area_type: noti.area_type,
						level: noti.level
					});
				}
				tile.SetData(noti.height, {
					build_height: noti.height,
					area_type: noti.area_type,
					level: noti.level
				});
				if(noti.exp != null) {
					this.exps.push(noti.exp);
					this.AddExp(noti.exp);
				}
			}
		}
		break;
	case GetCoinTransactionType:
		if(noti.error.length > 0) {
			this.AddCoin(noti.coin)
		} else {
			this.coins[noti.coin.index] = noti.coin;
			this.coin_count++;
			this.game_status_ui.OnCoinCountUpdated(this.coin_count);
		}
		break;
	case GetExpTransactionType:
		if(noti.error.length > 0) {
			this.exps.push(noti.exp);
			this.AddExp(noti.exp);
		} else {
			var bd = gGame.define_map[noti.exp.area_type][noti.exp.level-1];
			this.total_exp += bd.exp;
			this.game_status_ui.OnTotalExpUpdated(this.total_exp);
		}
		break;
	}
}

Game.prototype.closeSelectedInfo = function() {
	this.selectedInfo.hide()
}

Game.prototype.openSelectedInfo = function(tile) {
    if (tile.area_type) {
        this.selectedInfo.attr("class", getAreaTypeName(tile.area_type)).show()
        this.selectedInfo.find(".building_type").html(getAreaTypeName(tile.area_type))
        var lv = tile.level
        if (lv == 6) {
            this.selectedInfo.find(".building_level").html("lvFLETA")
        } else {
            this.selectedInfo.find(".building_level").html("lv"+tile.level)
        }

        if (lv > 0) {
            this.selectedInfo.find(".resource").html("+"+toShortUnit(gGame.define_map[tile.area_type][lv-1].output))
        } else {
            this.selectedInfo.find(".resource").html("under construction")
        }
    } else {
        this.selectedInfo.hide()
    }

}


function Tile(obj, x, y, tileData) {
	this.obj = obj;
	this.x = x;
	this.y = y;
	this.build_height = 0;
	this.area_type = 0;
	this.level = 0;
	this.target_area_type = 0;
	this.target_level = 0;
	this.update_height = 0;
	this.is_building = false;
	this.is_pending = false;
	this.obj.css("position", "absolute");
	this.obj.css("left", (gConfig.Size - 1 + this.x - this.y)/2 + "rem");
	this.obj.css("top", (this.x/2 + this.y/2)/2 + "rem");
	this.obj.css("width", (gConfig.Size/gConfig.Size) + "rem");
	this.obj.css("height", (gConfig.Size/2/gConfig.Size) + "rem");
	this.obj.css("z-index", (this.x+this.y)*10);
	var $img = this.obj.find(".floor")
	if ($img.length == 0) {
		$img = $("<img class='floor' src='/public/images/tile/base_floor/groundtiles_tile"+getNum(this.x, this.y)+".png'>").appendTo(this.obj);
	}
	$img.css("z-index", (this.x+this.y)*10);
	var $selected = this.obj.find(".selected");
	if ($selected.length == 0) {
		$selected = $("<div class='selected'>").appendTo(this.obj);
	}
	$selected.css("z-index", (this.x+this.y)*10+9);
	this.obj.find(".selected").hide();
}

Tile.prototype.SetData = function(height, data) {
	this.build_height = data.build_height;
	this.area_type = data.area_type;
	this.level = data.level;
	this.is_pending = false;
	if(this.level > 0) {
		var bd = gGame.define_map[this.area_type][this.level-1];
		this.is_building = this.build_height + bd.build_time*2 > height;
	} else {
		this.is_building = false;
	}
	this.removeAllImages();
	if(this.is_building) {
		this.renderConstruction(this.level);
	}
	this.Update(height, true);
}

Tile.prototype.CancelPending = function(height) {
	if(this.is_pending) {
		this.target_area_type = 0;
		this.target_level = 0;
		this.is_pending = false;
		this.removeAllImages();
		this.renderBuilding(this.level, true);
	}
}

Tile.prototype.SetPending = function(target_area_type, target_level, is_fletasub) {
	if(this.is_pending != true) {
		this.is_pending = true;
		this.target_area_type = target_area_type;
		this.target_level = target_level;
		if(is_fletasub) {
			this.removeAllImages();
		} else {
			if(this.target_level == 0) { // Destruction
				this.renderPending(this.level);
				if(this.level < 5) {
					for(var i=1; i<=this.level; i++) {
						var zindex = (this.x+this.y)*10+1+(3-Math.abs(3-i))
						this.buildEffect("distructionEffect", zindex, i)
					}
					this.obj.css("z-index", (this.x+this.y)*10+1);
				} else if(this.level == 5) {
					var zindex = (this.x+this.y)*10+1+this.level
					this.buildEffect("distructionEffect", zindex)
					this.obj.css("z-index", (this.x+this.y)*10+1);
				} else {
					var zindex = (this.x+this.y)*10+1+this.level
					this.buildEffect("distructionEffect", zindex)
					this.obj.css("z-index", (this.x+this.y)*10-1);
				}
			} else { // Construction
				this.removeAllImages();
				this.renderConstruction(this.target_level);
				this.renderPending(this.target_level);
			}
		}
	}
}

Tile.prototype.Resize = function() {
}

Tile.prototype.Update = function(height, force) {
	var is_begin = (this.update_height == 0);
	if(this.update_height != height || force) {
		this.update_height = height;

		if(!this.is_pending) {
			var is_renderable = false;
			if(this.is_building) {
				var bd = gGame.define_map[this.area_type][this.level-1];
				if(this.build_height + bd.build_time*2 <= height) {
					this.removeConstruction();
					this.is_building = false;
					is_renderable = true;
				} else {
					this.obj.find(".underconstruction").text(secondToDate((this.build_height + bd.build_time*2 - height)/2));
				}
			}
			
			if ((is_renderable || is_begin) && this.is_building != true){
				this.renderBuilding(this.level, is_begin);
			}
		}
	}
}

Tile.prototype.buildEffect = function(type, zindex, lv) {
	lv = lv || this.level
	var effect = $("<div class='lv"+(lv)+" buildEffect "+type+"'/>")
	effect.css("z-index", zindex);
	this.obj.append(effect);
	(function (effect) {
		setTimeout(function () {
			effect.remove()
		}, 1500)
	})(effect)

	if (type == "constructEffect") {
		this.completEffect(zindex)
	}
}

Tile.prototype.fletaEffect = function(zindex) {
	var effect = $("<div class='FLETAAnimation lv"+(this.level)+"'/>")
	effect.css("z-index", zindex);
	this.obj.append(effect);
	(function (effect) {
		setTimeout(function () {
			effect.remove()
		}, 3000)
	})(effect)
}

Tile.prototype.completEffect = function(zindex) {
	var effect = $("<div class='completAnimation lv"+(this.level)+"'/>")
	effect.css("z-index", zindex);
	this.obj.append(effect);
	(function (effect, tileUI) {
		setTimeout(function () {
			effect.remove()
			if (tileUI.level == 6) {
				tileUI.fletaEffect(zindex)
			}
		}, 3000)
	})(effect, this)
}

Tile.prototype.OnSelected = function() {
	this.obj.find(".selected").addClass("selected_shown").css("filter", "").show();
}

Tile.prototype.OnAllocated = function() {
	this.obj.find(".selected").addClass("selected_shown").css("filter", "grayscale(100%)").show();
}

Tile.prototype.removeAllImages = function() {
	this.obj.find(".building").detach();
	this.obj.find(".underconstruction").detach();
}

Tile.prototype.removeConstruction = function() {
	this.obj.find(".construction").detach();
	this.obj.find(".underconstruction").detach();
}

Tile.prototype.renderConstruction = function(level) {
	if(level < 5) {
		for(var i=1; i<=level; i++) {
			if(i < level) {
				var $img = $("<img class='building lv"+i+"' src='/public/images/building/"+getAreaTypeName(this.area_type)+"_Lv1.png'/>").appendTo(this.obj);
				$img.css("z-index", (this.x+this.y)*10+1+i);
			} else {
				var $img = $("<img class='construction building lv"+i+"' src='/public/images/building/construction.png'/>").appendTo(this.obj);
				$img.css("z-index", (this.x+this.y)*10+1+i);
			}
		}
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		this.obj.css("z-index", (this.x+this.y)*10);
	} else if(level == 5) {
		var $img = $("<img class='construction building lv"+level+"' src='/public/images/building/construction.png'/>").appendTo(this.obj);
		$img.css("z-index", (this.x+this.y)*10+1+level);
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		this.obj.css("z-index", (this.x+this.y)*10);
	} else {
		var $img = $("<img class='construction building lv"+level+"' src='/public/images/building/construction.png'/>").appendTo(this.obj);
		$img.css("z-index", (this.x+this.y)*10+1+level);
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		$span.addClass("fleta");
		this.obj.css("z-index", (this.x+this.y)*10-1);
	}
}

Tile.prototype.renderPending = function(level) {
	if(level < 5) {
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		$span.text("Pending");
		this.obj.css("z-index", (this.x+this.y)*10);
	} else if(level == 5) {
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		$span.text("Pending");
		this.obj.css("z-index", (this.x+this.y)*10);
	} else {
		var $span = this.obj.find(".underconstruction");
		if($span.length == 0) {
			$span = $("<span class='underconstruction'/>").appendTo(this.obj);
			$span.css("z-index", (this.x+this.y)*10+8);
		}
		$span.addClass("fleta");
		$span.text("Pending");
		this.obj.css("z-index", (this.x+this.y)*10-1);
	}
}

Tile.prototype.renderBuilding = function(level, no_effect) {
	if(level < 5) {
		for(var i=1; i<=level; i++) {
			if(this.obj.find(".lv"+i).length == 0) {
				var $img = $("<img class='building lv"+i+"' src='/public/images/building/"+getAreaTypeName(this.area_type)+"_Lv1.png'/>").appendTo(this.obj);
				var zindex = (this.x+this.y)*10+1+(3-Math.abs(3-i))
				$img.css("z-index", zindex);
				if(!no_effect) {
					this.buildEffect("constructEffect", zindex)
				}
			}
		}
		this.obj.css("z-index", (this.x+this.y)*10);
	} else if(level == 5) {
		var $img = $("<img class='building lv"+level+"' src='/public/images/building/"+getAreaTypeName(this.area_type)+"_Lv5.png'/>").appendTo(this.obj);
		var zindex = (this.x+this.y)*10+1+level
		$img.css("z-index", zindex);
		this.obj.css("z-index", (this.x+this.y)*10);
		if(!no_effect) {
			this.buildEffect("constructEffect", zindex)
		}
	} else {
		var $img = $("<img class='building lv"+level+"' src='/public/images/tile/"+getAreaTypeName(this.area_type)+"_LvFLETA-Tile.png' style='opacity:1'/>").appendTo(this.obj);
		$img.css("z-index", (this.x+this.y)*10+level);
		var $img = $("<img class='building lv"+level+"' src='/public/images/building/"+getAreaTypeName(this.area_type)+"_LvFLETA.png'/>").appendTo(this.obj);
		var zindex = (this.x+this.y)*10+1+level
		$img.css("z-index", zindex);
		this.obj.css("z-index", (this.x+this.y)*10-1);
		if(!no_effect) {
			this.buildEffect("constructEffect", zindex)
		}
	}
}

function GameStatusUI(obj) {
	this.obj = obj;
}

GameStatusUI.prototype.OnResourceUpdated = function(resource) {
	var $scoreBoard = $("#scoreboard");
	for (var key in resource) {
		var $board = $scoreBoard.find("."+key);
		if ($board.length > 0) {
			if (key == "add_balance") {
				$board.text("(+"+(resource[key]*2)+"/s)");
			} else {
				$board.text(toShortUnit(resource[key]));
			}
		}
	}
}

GameStatusUI.prototype.OnTotalExpUpdated = function(exp) {
	var t = expIndexOf(exp)
	console.log("GameStatusUI", "OnTotalExpUpdated", exp, t);
	
	var eh = $("#expHeader");
	var ei = $("#expUI");
	var c = ei.attr("class");
	if (c != t.current["class"] && c != "" && typeof c != "undefined") {
		eh.addClass("do_effect");
	}
	eh.attr("lvstep", t.current["class"]);
	ei.attr("class", t.current["class"]);

	var lvEXp = exp - t.current.acc_exp;

	$("#expGauge").css("width", (lvEXp/t.next.exp*100)+"%");
	$("#currentLevel").html(t.current.level);
	$("#currentExp").html(lvEXp);
	$("#currentMaxExp").html(t.next.exp);
}

GameStatusUI.prototype.OnCoinCountUpdated = function(coin) {
	console.log("GameStatusUI", "OnCoinCountUpdated", coin);
	$("#scoreboard .coin_count").text(toShortUnit(coin));
}

function BuildMenuUI(obj) {
	this.obj = $("#build_menu");
	this.resource = null;
	this.is_opened = false;
}

BuildMenuUI.prototype.Init = function(define_map) {
	if (typeof this.define_map == "object") {
		return;
	}
	this.define_map = define_map;
	var _this = this;

	var $commercial = this.obj.find(".commercial");
	var bd = this.define_map[CommercialAreaType][0];
	$commercial.find(".needDemographic").text(numberWithCommas(bd.man_usage));
	$commercial.find(".needDollar").text(numberWithCommas(bd.cost_usage));
	$commercial.find(".needPower").text(numberWithCommas(bd.power_usage));
	$commercial.find(".needTime").text(secondToDate(bd.build_time));
	$commercial.find(".resource").text(bd.output+"/s");
	$commercial.click(function(e) {
		e.stopPropagation();
		if($(this).hasClass("disabled")) {
			return;
		}
		UIAlert.Alert("Commercial", function () {
			var err = getBuildError(_this.resource, _this.define_map[CommercialAreaType][0]);
			if(err != null) {
				Alert(err);
				return;
			}
			gGame.selected_tile.SetPending(CommercialAreaType, 1, false);
			gNetwork.SendTX("construction", {
				x: gGame.selected_tile.x,
				y: gGame.selected_tile.y,
				area_type: CommercialAreaType
			});
			_this.Close();
			gGame.UnselectTile();
		})
	});

	var $industrial = this.obj.find(".industrial");
	var bd = this.define_map[IndustrialAreaType][0];
	$industrial.find(".needDemographic").text(numberWithCommas(bd.man_usage));
	$industrial.find(".needDollar").text(numberWithCommas(bd.cost_usage));
	$industrial.find(".needTime").text(secondToDate(bd.build_time));
	$industrial.find(".resource").text(numberWithCommas(bd.output));
	$industrial.click(function(e) {
		e.stopPropagation();
		if($(this).hasClass("disabled")) {
			return;
		}
		UIAlert.Alert("Industrial", function () {
			var err = getBuildError(_this.resource, _this.define_map[IndustrialAreaType][0]);
			if(err != null) {
				Alert(err);
				return;
			}
			gGame.selected_tile.SetPending(IndustrialAreaType, 1, false);
			gNetwork.SendTX("construction", {
				x: gGame.selected_tile.x,
				y: gGame.selected_tile.y,
				area_type: IndustrialAreaType
			});
			_this.Close();
			gGame.UnselectTile();
		})
	});

	var $residential = this.obj.find(".residential");
	var bd = this.define_map[ResidentialAreaType][0];
	$residential.find(".needDollar").text(numberWithCommas(bd.cost_usage));
	$residential.find(".needPower").text(numberWithCommas(bd.power_usage));
	$residential.find(".needTime").text(secondToDate(bd.build_time));
	$residential.find(".resource").text(numberWithCommas(bd.output));
	$residential.click(function(e) {
		e.stopPropagation();
		if($(this).hasClass("disabled")) {
			return;
		}
		UIAlert.Alert("Residential", function () {
			var err = getBuildError(_this.resource, _this.define_map[ResidentialAreaType][0]);
			if(err != null) {
				Alert(err);
				return;
			}
			gGame.selected_tile.SetPending(ResidentialAreaType, 1, false);
			gNetwork.SendTX("construction", {
				x: gGame.selected_tile.x,
				y: gGame.selected_tile.y,
				area_type: ResidentialAreaType
			});
			_this.Close();
			gGame.UnselectTile();
		})
	});
}

BuildMenuUI.prototype.Open = function(tile) {
	if(!this.is_opened) {
		this.is_opened = true;
		UIAlert.hide();
		this.obj.show();
	}
	message("menu open x : " + tile.x + " y : " + tile.y );
	this.obj.css("left", (gConfig.Size - 1 + tile.x - tile.y)/2 + "rem");
	this.obj.css("top", (tile.x/2 + tile.y/2)/2 + "rem");
	if(this.resource != null) {
		this.OnResourceUpdated(this.resource);
	}
}

BuildMenuUI.prototype.Close = function() {
	if(this.is_opened) {
		this.is_opened = false;
		this.obj.hide();
	}
}

BuildMenuUI.prototype.OnResourceUpdated = function(resource) {
	this.resource = resource;
	if(this.is_opened) {
		var $commercial = this.obj.find(".commercial");
		if(getBuildError(this.resource, this.define_map[CommercialAreaType][0]) != null) {
			$commercial.addClass("disabled");
		} else {
			$commercial.removeClass("disabled");
		}
		var $industrial = this.obj.find(".industrial");
		if(getBuildError(this.resource, this.define_map[IndustrialAreaType][0]) != null) {
			$industrial.addClass("disabled");
		} else {
			$industrial.removeClass("disabled");
		}
		var $residential = this.obj.find(".residential");
		if(getBuildError(this.resource, this.define_map[ResidentialAreaType][0]) != null) {
			$residential.addClass("disabled");
		} else {
			$residential.removeClass("disabled");
		}
	}
}

function UpgradeMenuUI() {
	this.obj = $("#upgrade_menu");
	this.resource = null;
	this.is_opened = false;
	this.is_demolition_disabled = false;
	this.is_upgrade_disabled = false;
	this.fleta_subtiles = null;

	var _this = this;
	var $demolition = this.obj.find(".demolition");
	$demolition.click(function(e) {
		e.stopPropagation();
		if($(this).hasClass("disabled")) {
			return;
		}
		UIAlert.Alert("Demolition", function () {
			gGame.selected_tile.SetPending(0, 0, false);
			gNetwork.SendTX("demolition", {
				x: gGame.selected_tile.x,
				y: gGame.selected_tile.y,
			});
			_this.Close();
			gGame.UnselectTile();
		})
	});
	var $upgrade = this.obj.find(".upgrade");
	$upgrade.click(function(e) {
		e.stopPropagation();
		if($(this).hasClass("disabled")) {
			return;
		}
		UIAlert.Alert("Upgrade", function () {
			if(gGame.selected_tile.level < 6) {
				var err = getBuildError(_this.resource, gGame.define_map[gGame.selected_tile.area_type][gGame.selected_tile.level]);
				if(err != null) {
					Alert(err);
					return;
				}
				var target_tile = gGame.selected_tile;
				if(gGame.selected_tile.level == 5) {
					var subtiles = gGame.GetFletaTiles(gGame.selected_tile);
					if(subtiles != null) {
						for(var i=1; i<subtiles.length; i++) {
							subtiles[i].SetPending(0, 0, true);
						}
						target_tile = subtiles[0];
					}
				}
				target_tile.SetPending(target_tile.area_type, target_tile.level+1, false);
				gNetwork.SendTX("upgrade", {
					x: target_tile.x,
					y: target_tile.y,
					area_type: target_tile.area_type,
					target_level: target_tile.level+1
				});
				_this.Close();
				gGame.UnselectTile();
			}
		})
	});
}

UpgradeMenuUI.prototype.Open = function(tile) {
	if(!this.is_opened) {
		this.is_opened = true;
		UIAlert.hide();
		this.obj.show();
	}
	message("menu open x : " + tile.x + " y : " + tile.y );
	this.obj.css("left", (gConfig.Size - 1 + tile.x - tile.y)/2 + "rem");
	this.obj.css("top", (tile.x/2 + tile.y/2)/2 + "rem");

	var suffix = "";
	if(gGame.selected_tile.area_type == CommercialAreaType) {
		suffix = "/s";
	}
	var bd = gGame.define_map[gGame.selected_tile.area_type][gGame.selected_tile.level-1];
	var $demolition = this.obj.find(".demolition");
	$demolition.find(".needDemographic").text(numberWithCommas(bd.acc_man_usage));
	$demolition.find(".needDollar").text(0);
	$demolition.find(".needPower").text(numberWithCommas(bd.acc_power_usage));
	$demolition.find(".needTime").text("1s");
	$demolition.find(".resource").text("-" + bd.output+suffix).attr("class", "resource").addClass(getAreaTypeName(gGame.selected_tile.area_type));

	this.fleta_subtiles = null;
	if(gGame.selected_tile.level == 5) {
		this.fleta_subtiles = gGame.GetFletaTiles(gGame.selected_tile);
	}

	var $upgrade = this.obj.find(".upgrade");
	if(gGame.selected_tile.level < 6) {
		$upgrade.show();
		var bd = gGame.define_map[gGame.selected_tile.area_type][gGame.selected_tile.level];
		$upgrade.find(".needDemographic").text(numberWithCommas(bd.man_usage));
		$upgrade.find(".needDollar").text(numberWithCommas(bd.cost_usage));
		$upgrade.find(".needPower").text(numberWithCommas(bd.power_usage));
		$upgrade.find(".needTime").text(secondToDate(bd.build_time));
		$upgrade.find(".resource").text(numberWithCommas(bd.output)+suffix).attr("class", "resource").addClass(getAreaTypeName(gGame.selected_tile.area_type))
		if(this.resource != null) {
			this.OnResourceUpdated(this.resource);
		}
	} else {
		$upgrade.hide();
	}

	if(this.resource != null) {
		this.OnResourceUpdated(this.resource);
	}
}

UpgradeMenuUI.prototype.Close = function() {
	if(this.is_opened) {
		this.is_opened = false;
		this.obj.hide();
	}
}

UpgradeMenuUI.prototype.OnResourceUpdated = function(resource) {
	this.resource = resource;
	if(this.is_opened) {
		var bd = gGame.define_map[gGame.selected_tile.area_type][gGame.selected_tile.level-1];
		var $demolition = this.obj.find(".demolition");
		var has_demolition_disable = true;
		switch(gGame.selected_tile.area_type) {
		case CommercialAreaType:
			has_demolition_disable = false;
			break;
		case IndustrialAreaType:
			if(resource.power_remained - bd.output >= 0) {
				has_demolition_disable = false;
			}
			break;
		case ResidentialAreaType:
			if(resource.man_remained - bd.output >= 0) {
				has_demolition_disable = false;
			}
			break;
		}
		if(has_demolition_disable) {
			if(!this.is_demolition_disabled) {
				this.is_demolition_disabled = true;
				$demolition.addClass("disabled");
			}
		} else {
			if(this.is_demolition_disabled) {
				this.is_demolition_disabled = false;
				$demolition.removeClass("disabled");
			}
		}

		var $upgrade = this.obj.find(".upgrade");
		var has_upgrade_disable = true;
		if(gGame.selected_tile.level < 6) {
			if(gGame.selected_tile.level == 5 && this.fleta_subtiles == null) {
				// ignore
			} else if(getBuildError(this.resource, gGame.define_map[gGame.selected_tile.area_type][gGame.selected_tile.level]) != null) {
				// ignore
			} else {
				has_upgrade_disable = false;
			}
		}
		if(has_upgrade_disable) {
			if(!this.is_upgrade_disabled) {
				this.is_upgrade_disabled = true;
				$upgrade.addClass("disabled");
			}
		} else {
			if(this.is_upgrade_disabled) {
				this.is_upgrade_disabled = false;
				$upgrade.removeClass("disabled");
			}
		}
	}
}

var gConfig = {
	Selector: {
		Screen: "#screen",
		Touchpad: "#touchpad",
	},
	Unit: 58,
	Size: 32,
};

function Network(utxos) {
	this.utxos = utxos;
	this.waitQ = [];
}

Network.prototype.SendTX = function(command, data) {
	if(this.utxos.length > 0) {
		var utxo = this.utxos[0];
		this.utxos.splice(0, 1);
		this.Execute(command, data, utxo);
	} else {
		this.waitQ.push({
			command: command,
			data: data,
		});
	}
}

Network.prototype.OnNotified = function(noti) {
	if(this.waitQ.length > 0) {
		var work = this.waitQ[0];
		this.waitQ.splice(0, 1);
		this.Execute(work.command, work.data, noti.utxo);
	} else {
		this.utxos.push(noti.utxo);
	}
	console.log(noti);
}

Network.prototype.Execute = function(command, data, utxo) {
	var _this = this;
	data.utxo = utxo;
	$.ajax({
		type: "POST",
		url : "/api/games/"+Login.address+"/commands/" + command,
		data : JSON.stringify(data),
		success : function (d) {
			if (typeof d === "string") {
				d = JSON.parse(d)
			}
			_this.Commit(d);
		},
		error: function(d) {
			Alert(language["Failed to execute upgrade command"])
		}
	});
}

Network.prototype.Commit = function(data) {
	var msg = new Buffer(data.hash_hex, "hex");
	var sig = Login.key.sign(msg);
	var SIG_HEX = buf2hex(sig.r.toArrayLike(Buffer, "be", 32)) + buf2hex(sig.s.toArrayLike(Buffer, "be", 32)) + "0" + sig.recoveryParam;

	$.ajax({
		type: "POST",
		url : "/api/games/"+Login.address+"/commands/commit",
		data : JSON.stringify({
			"type": data.type,
			"tx_hex": data.tx_hex,
			"sig_hex": SIG_HEX
		}),
		success : function (d) {
		},
		error: function(d) {
			var msg = language[d.responseText]
			if (typeof msg == "string") {
				Alert(msg)
			} else {
				if (data.type == 2) {
					Alert(language["Failed to execute demolation command"])
				} else if (data.type == 3) {
					Alert(language["Failed to execute upgrade command"])
				} else {
					Alert(language["commit error"])
				}
			}
		}
	})
}

function ChangeUnit(unit) {
	if (!isNaN(unit)) {
		gConfig.Unit = unit;

		var h = [], i =0
		h[i++] = ".island{width:"+(gConfig.Size*1.12625)+"rem;height:"+(gConfig.Size*0.84875)+"rem;margin-left: -"+(gConfig.Size*1.12625)/2+"rem;margin-top: -"+(gConfig.Size*0.84875)/2+"rem;}"
		h[i++] = "#tileCase{top:"+(gConfig.Size*0.251875)+"rem;left:"+(gConfig.Size*0.0625)+"rem;width:"+(gConfig.Size)+"rem;height:"+(gConfig.Size/2)+"rem}"

		$("#cssControll").html(h.join("\n"));
		$("html").css("font-size", gConfig.Unit+"px");
		for(var i=0; i<gGame.tiles.length; i++) {
			gGame.tiles[i].Resize();
		}
	}
}

/*
function initGame () {
    ChangeUnit(gConfig.Unit)
    var jScreen = $("#tileCase");
    jScreen.css("width", (gConfig.Size)+"rem");
    jScreen.css("height", (gConfig.Size)/2+"rem");

	connectToServer(Login.address)
	loadTile()
	scoreReloader()
	addKeyShotcut()
}
*/
