
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

Game.prototype.Update = function() {
	/*
	var forward_height = this.height - this.point_height;
	var base = {
		balance:       4,
		power_remained: 5,
		power_provided: 5,
		man_remained:   3,
		man_provided:   3,
	}
	used := &Resource{}
	provide := &Resource{
		balance:       this.point_balance + uint64(base.balance/2)*uint64(forward_height),
		power_remained: base.power_remained,
		man_remained:   base.man_remained,
	}
	for _, tile := range gd.Tiles {
		if tile != nil {
			bd := gBuildingDefine[tile.AreaType][tile.Level-1]
			used.man_remained += bd.ManUsage
			used.power_remained += bd.PowerUsage

			ConstructionHeight := tile.BuildHeight + bd.BuildTime*2
			if this.height < ConstructionHeight {
				if tile.Level == 1 {
					continue
				}
				bd = gBuildingDefine[tile.AreaType][tile.Level-2]
			}
			switch tile.AreaType {
			case CommercialType:
				if this.height > ConstructionHeight {
					if ConstructionHeight <= this.point_height {
						provide.balance += uint64(bd.Output/2) * uint64(forward_height)
					} else {
						if tile.BuildHeight <= this.point_height {
							provide.balance += uint64(bd.Output/2) * uint64(forward_height-(ConstructionHeight-this.point_height))
						} else {
							provide.balance += uint64(bd.Output/2) * uint64(this.height-ConstructionHeight)
							if tile.Level > 1 {
								prevbd := gBuildingDefine[tile.AreaType][tile.Level-2]
								provide.balance += uint64(prevbd.Output/2) * uint64(tile.BuildHeight-this.point_height)
							}
						}
					}
				}
			case IndustrialType:
				provide.power_remained += bd.Output
			case ResidentialType:
				provide.man_remained += bd.Output
			}
		}
	}
	return {
		balance:       provide.balance,
		power_remained: provide.power_remained - used.power_remained,
		power_provided: provide.power_remained,
		man_remained:   provide.man_remained - used.man_remained,
		man_provided:   provide.man_remained,
	};
	*/
}

var gGame = new Game();
var CommercialType = 1;
var IndustrialType = 2;
var ResidentialType = 3;

var MENU = {
	lv0 : {"Industrial":"Industrial", "Residential" : "Residential", "Commercial" : "Commercial"},
	lv1 : {"Demolition":"Demolition", "Upgrade":"Upgrade"},
	lv2 : {"Demolition":"Demolition", "Upgrade":"Upgrade"},
	lv3 : {"Demolition":"Demolition", "Upgrade":"Upgrade"},
	lv4 : {"Demolition":"Demolition", "Upgrade":"Upgrade"},
	lv5 : {"Demolition":"Demolition", "Fleta!":"Upgrade"},
	lv6 : {"Demolition":"Demolition"},
}
