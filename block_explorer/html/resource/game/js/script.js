var gConfig = {
	Unit: 64,
    Size: 16,
}

function Game() {
	this.height = 0;
	this.point_height = 0;
	this.point_balance = 0;
	this.tiles = [];
	this.define_map = null;
}

var currentResource = {}

Game.prototype.Update = function() {
	var forward_height = this.height - this.point_height;
	var base = {
		balance:       4,
		power_remained: 5,
		power_provided: 5,
		man_remained:   3,
		man_provided:   3,
	}
	var used = new Resource();
	var provide = new Resource(this.point_balance + parseInt(base.balance/2)*parseInt(forward_height), base.power_remained, base.man_remained);

	var addBalance = 0
	for (var k in gGame.tiles) {
		var tile = gGame.tiles[k];
		if (tile.level) {
			var bd = gBuildingDefine[tile.type][tile.level-1];
			used.man_remained += bd.man_usage||0;
			used.power_remained += bd.power_usage||0;

			if (tile.type == CommercialType) {
				addBalance += bd.output;
			}

			var ConstructionHeight = tile.build_height + bd.build_time*2
			if (this.height < ConstructionHeight) {
				if (tile.level == 1) {
					continue;
				}
				bd = gBuildingDefine[tile.type][tile.level-2];
			}
			switch (tile.type) {
			case CommercialType:
				if (this.height > ConstructionHeight) {
					if (ConstructionHeight <= this.point_height) {
						provide.balance += bd.output/2 * parseInt(forward_height);
					} else {
						if (tile.build_height <= this.point_height) {
							provide.balance += bd.output/2 * parseInt(forward_height-(ConstructionHeight-this.point_height));
						} else {
							provide.balance += bd.output/2 * parseInt(this.height-ConstructionHeight);
							if (tile.level > 1) {
								var prevbd = gBuildingDefine[tile.type][tile.level-2];
								provide.balance += prevbd.output/2 * parseInt(tile.build_height-this.point_height);
							}
						}
					}
				}
				break;
			case IndustrialType:
				provide.power_remained += bd.output;
				break;
			case ResidentialType:
				provide.man_remained += bd.output;
				break;
			}

		}
		if (tile.obj.BuildProcessing) {
			var sTile = tile
			if (tile.obj.headTile) {
				sTile = tile.obj.headTile;
			}
			var buildCompletHeight = sTile.build_height + gGame.define_map[sTile.type][sTile.level].build_time*2
			
			if(buildCompletHeight <= gGame.height) {
				sTile.UI.completBuilding(tile.level+1)
			} else {
				sTile.UI.ShowBuildProcessingTime((buildCompletHeight-gGame.height)/2)
			}
		}
	}

	currentResource.balance = provide.balance;
	currentResource.power = provide.power_remained - used.power_remained;
	currentResource.men = provide.man_remained - used.man_remained;

	return {
		balance:       currentResource.balance,
		power_remained: currentResource.power,
		power_provided: provide.power_remained,
		man_remained:   currentResource.men,
		man_provided:   provide.man_remained,
		add_balance:   addBalance,
	};
}

function Resource (balance, power_remained, man_remained) {
	this.balance = balance||0;
	this.power_remained = power_remained||0;
	this.man_remained = man_remained||0;
}

var gGame = new Game();
var CommercialType = 1;
var IndustrialType = 2;
var ResidentialType = 3;

var MENU = {}

function buildingType(num) {
	switch (num) {
		case 1:
		return "Commercial"
		case 2:
		return "Industrial"
		case 3:
		return "Residential"
	}
}

function buildingNum(str) {
	switch (str) {
		case "Commercial":
		return 1
		case "Industrial":
		return 2
		case "Residential":
		return 3
	}
}

function buildableResource(tile, type) {
	if (type == "Demolition") {
		return true
	}
	if (tile.obj.BuildProcessing == true) {
		return false;
	}
	type = tile.type||buildingNum(type)
	if (type) {
		var cost = gBuildingDefine[type][tile.level];
		if (typeof cost.cost_usage != "undefined" && currentResource.balance < cost.cost_usage) {
			return language["not enough balance"]
		}
		if (typeof cost.man_usage != "undefined" && currentResource.men < cost.man_usage) {
			return language["not enough people"]
		}
		if (typeof cost.power_usage != "undefined" && currentResource.power < cost.power_usage) {
			return language["not enough power"]
		}
	}

	if (tile.level == 5) {
		var checker = tile.CheckLvRound()
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			if (checker.candidate[i].obj.BuildProcessing == true) {
				return language["BuildProcessing not finished"]
			}
		}
		if (checker.CheckLvF()!=true) {
			return language["not enough lv5 building"]
		}
	}

	return true
}

var gBuildingDefine = {
	"1": [
		{ "cost_usage": 400, "build_time": 30, "output": 4, "man_usage": 2, "power_usage": 3, "acc_man_usage": 2, "acc_power_usage": 3 },
		{ "cost_usage": 2400, "build_time": 140, "output": 10, "man_usage": 3, "power_usage": 4, "acc_man_usage": 5, "acc_power_usage": 7 }, 
		{ "cost_usage": 12000, "build_time": 700, "output": 24, "man_usage": 8, "power_usage": 12, "acc_man_usage": 13, "acc_power_usage": 19 }, 
		{ "cost_usage": 60000, "build_time": 3500, "output": 64, "man_usage": 40, "power_usage": 30, "acc_man_usage": 53, "acc_power_usage": 49 }, 
		{ "cost_usage": 300000, "build_time": 18000, "output": 160, "man_usage": 200, "power_usage": 80, "acc_man_usage": 253, "acc_power_usage": 129 }, 
		{ "cost_usage": 6000000, "build_time": 86400, "output": 1600, "man_usage": 4000, "power_usage": 1500, "acc_man_usage": 4253, "acc_power_usage": 1629 }
	],
	"2": [
		{ "cost_usage": 200, "build_time": 60, "output": 5, "man_usage": 1, "power_usage": 0, "acc_man_usage": 1, "acc_power_usage": 0 },
		{ "cost_usage": 1700, "build_time": 200, "output": 14, "man_usage": 2, "power_usage": 0, "acc_man_usage": 3, "acc_power_usage": 0 },
		{ "cost_usage": 12000, "build_time": 700, "output": 96, "man_usage": 8, "power_usage": 0, "acc_man_usage": 11, "acc_power_usage": 0 }, 
		{ "cost_usage": 80000, "build_time": 2700, "output": 390, "man_usage": 54, "power_usage": 0, "acc_man_usage": 65, "acc_power_usage": 0 }, 
		{ "cost_usage": 450000, "build_time": 12000, "output": 1440, "man_usage": 300, "power_usage": 0, "acc_man_usage": 365, "acc_power_usage": 0 }, 
		{ "cost_usage": 9100000, "build_time": 57000, "output": 33000, "man_usage": 6100, "power_usage": 0, "acc_man_usage": 6465, "acc_power_usage": 0 }
	],
	"3": [
		{ "cost_usage": 300, "build_time": 45, "output": 3, "man_usage": 0, "power_usage": 2, "acc_man_usage": 0, "acc_power_usage": 2 }, 
		{ "cost_usage": 2000, "build_time": 170, "output": 10, "man_usage": 0, "power_usage": 3, "acc_man_usage": 0, "acc_power_usage": 5 }, 
		{ "cost_usage": 12000, "build_time": 700, "output": 64, "man_usage": 0, "power_usage": 12, "acc_man_usage": 0, "acc_power_usage": 17 }, 
		{ "cost_usage": 66000, "build_time": 3200, "output": 564, "man_usage": 0, "power_usage": 35, "acc_man_usage": 0, "acc_power_usage": 52 }, 
		{ "cost_usage": 360000, "build_time": 15000, "output": 4000, "man_usage": 0, "power_usage": 100, "acc_man_usage": 0, "acc_power_usage": 152 }, 
		{ "cost_usage": 7200000, "build_time": 72000, "output": 101000, "man_usage": 0, "power_usage": 1800, "acc_man_usage": 0, "acc_power_usage": 1952 }
	]
}

