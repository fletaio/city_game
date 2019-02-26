function initGame () {
    ChangeUnit(gConfig.Unit)
    var jScreen = $("#tileCase");
    jScreen.css("width", (gConfig.Size)+"rem");
    jScreen.css("height", (gConfig.Size)/2+"rem");

	connectToServer(loginInfo.Addr)
	loadTile()
	scoreReloader()
	addKeyShotcut()
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
			gGame.define_map = gBuildingDefine;
			gGame.txs = d.txs;
			gGame.coin_list = {}
			for (var k in d.fleta_city_coins) {
				gGame.coin_list[k] = new FletaCityCoin(d.fleta_city_coins[k])
			}
			if (d.coin_count) {
				$("[key='coin_count']").html(d.coin_count)
			}

			gGame.height = d.height;
			gGame.point_height = d.point_height;
			gGame.point_balance = d.point_balance;
			for(var i=0; i<d.tiles.length; i++) {
				var x = i%gConfig.Size;
				var y = parseInt(i/gConfig.Size);

				if (d.tiles[i]) {
					var tile = new Tile(x, y, d.tiles[i].area_type, d.tiles[i].level , d.tiles[i].build_height)
				} else {
					var tile = new Tile(x, y)
				}
				gGame.tiles.push(tile);
				if (tile.level == 6) {
					var o = {x:tile.x, y:tile.y}
					for ( var j = 0 ; j < 4 ; j++ ) {
						directByNum(o, j%4);
						if (o.x >= 0 && o.x < gConfig.Size && o.y >= 0 && o.y < gConfig.Size) {
							var t = gGame.tiles[o.x + o.y * gConfig.Size];
							t.level = tile.level
							t.build_height = tile.build_height
							t.type = tile.type
						}
					}
				}
				var num = getNum(x, y)
				tile.init(new TileUI(jScreen, $touchpad, num))
			}

			for (var i = 0 ; i < gGame.txs.length ; i++) {
				addTx(gGame.txs[i])
			}
		},
		error: function(d) {
			Alert(language["load fail"])
		}
	})

}

function ChangeUnit(unit) {
	gConfig.Unit = unit;

	var h = [], i =0
	h[i++] = ".island{width:"+(gConfig.Size*1.12625)+"rem;height:"+(gConfig.Size*0.84875)+"rem}"
	h[i++] = "#tileCase{top:"+(gConfig.Size*0.251875)+"rem;left:"+(gConfig.Size*0.0625)+"rem}"

	$("#cssControll").html(h.join("\n"));
	$("html").css("font-size", gConfig.Unit+"px");
}
