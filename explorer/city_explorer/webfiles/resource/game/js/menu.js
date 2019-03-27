function menuClose () {
	UIAlert.hide()
	$("#txLog [index]").show()
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
			$tooltip.find("#needDollar").html(toShortUnit(r.cost_usage))
			$tooltip.find("#needPower").html(toShortUnit(r.power_usage))
			$tooltip.find("#needDemographic").html(toShortUnit(r.man_usage))

			if (funcs[key] !== "Commercial") {
				$tooltip.find("#resource").attr("class", buildingType(type)).html("+"+toShortUnit(r.output)).attr("class", buildingType(type))
			} else {
				$tooltip.find("#resource").attr("class", buildingType(type)).html("+"+toShortUnit(r.output)+"/s").attr("class", buildingType(type))
			}

			var time = secondToDate(r.build_time);
		} else {
			var able = demolitionableResource(tile, funcs[key])

			var r = gBuildingDefine[tile.type][tile.level-1]
			$tooltip.find("#needDollar").html("+"+toShortUnit(r.cost_usage/2))
			if (tile.type == IndustrialType) {
				$tooltip.find("#needPower").html("-"+toShortUnit(r.output))
			} else {
				$tooltip.find("#needPower").html("+"+toShortUnit(r.power_usage))
			}
			if (tile.type == ResidentialType) {
				$tooltip.find("#needDemographic").html("-"+toShortUnit(r.output))
			} else {
				$tooltip.find("#needDemographic").html("+"+toShortUnit(r.man_usage))
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
			$selectedInfo.find(".resource").html("+"+toShortUnit(gGame.define_map[tile.type][lv-1].output))
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

function addTx (tx) {
	var txLog = $("#txLog")
	var t = addTx.templete

	tx.explorer = gConfig.Explorer||""
	tx.index = tx.x+tx.y*gConfig.Size
	switch(tx.tx_type) {
		case 2:
			tx.type = "Demolition"
			break;
		case 3:
			tx.type = "Construction"
			break;
		case 4:
			tx.type = "Upgrade"
			break;
		case 5:
			tx.type = "GetCoin"
			break;
		default:
		return;
	}

	for (var k in tx) {
		if (tx.hasOwnProperty(k)) {
			var v = tx[k]
			t = t.replace(new RegExp("{"+k+"}", 'g'), v)
		}
	}
	txLog.append(t)
	txLog[0].scrollTop = txLog[0].scrollHeight;
}
addTx.templete = `
<div index="{index}">
	<p>{type} Tile = x : {x} , y : {y} </p>
	<p><a href="{explorer}/transactionDetail?hash={hash}" target="_blank"><span class="descript">txHash : </span><span>{hash}</span></a></p>
</div>
`
