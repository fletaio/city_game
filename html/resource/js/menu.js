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
		var btn = $("<button id=\""+funcs[key]+"\" value=\""+key+"\">"+key+"</button>")
		
		var $tooltip = $("#tooltip").clone();
		var $this = btn
		$tooltip.removeAttr("id")
		$tooltip.attr("class", "tooltip " + funcs[key])

		if (funcs[key] !== "Demolition") {
			var able = buildableResource(tile, funcs[key])

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
			var able = demolitionableResource(tile, funcs[key])

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
		if (able != true) {
			btn.addClass("disable")
			btn.attr("onclick", "event.stopPropagation();Alert('"+able+"');")
		} else {
			btn.attr("onclick", "event.stopPropagation();$('#menu')[0].target.RunCommand('"+funcs[key]+"');")
		}

		$tooltip.find("#needTime").html(time).attr("class", "")

		$("#menu").append($tooltip)
		$("#menu").append(btn)
    }
}

function menuOpen(tile) {
	UIAlert.hide()
	$(".tooltip").remove()
	if (tile.type) {
		var $selectedInfo = $("#selectedInfo")
		$selectedInfo.attr("class", buildingType(tile.type)).show()
		$selectedInfo.find(".building_type").html(buildingType(tile.type))
		var lv = tile.level
		if (tile.headTile) {
			lv = tile.headTile.level
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
	if (tile.headTile) {
		addMenu(tile.headTile, MENU["lv"+tile.headTile.level]);
	} else {
		addMenu(tile, MENU["lv"+tile.level]);
	}
	message("menu open x : " + tile.x + " y : " + tile.y );
	var $menu = $("#menu");
	$menu[0].target = tile;
	tile.UI.SelectTile();
	if (tile.headTile) {
		tile.headTile.UI.touch.append($menu.show())
	} else {
		tile.UI.touch.append($menu.show())
	}
}
