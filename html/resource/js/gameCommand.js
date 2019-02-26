Tile.prototype.RunCommand = function(func, param) {
	if (typeof this[func] === "function") {
		if (func == "GetCoin") {
			var tile = this[func](param);
			sendServer(func, tile, param)
			return
		}
		(function (This) {
			var tile = This;
			UIAlert.Alert(func, function () {
				message("command : "+ func + " x : " + tile.x + " y : " + tile.y );

				// var utxo = loginInfo.popUTXO()
				// if (!utxo) {
				// 	Alert(language["too fast"])
				// 	return
				// }
				try{
					tile = tile[func](param);
				} catch(e) {
					console.log(e)
				}
				if (IsTile(tile)){
					menuOpen(tile);
					// setTimeout(function () {
					// 	var balance = parseInt($("#dollar[key='balance']").html())
					// 	balance -=  gBuildingDefine[tile.type][tile.level].cost_usage
					// 	var height = gGame.height
					// 	if (func == "Demolition") {
					// 		onMessage({_init:true}, {data : "{\"point_height\":"+height+",\"point_balance\":"+balance+",\"x\":"+(tile.x)+",\"y\":"+(tile.y)+",\"area_type\":0,\"level\":0,\"type\":0,\"height\":"+gGame.height+"}"})
					// 	} else {
					// 		onMessage({_init:true}, {data : "{\"point_height\":"+height+",\"point_balance\":"+balance+",\"x\":"+(tile.x)+",\"y\":"+(tile.y)+",\"area_type\":"+tile.type+",\"level\":"+(tile.level+1)+",\"type\":1,\"height\":"+gGame.height+"}"})
					// 	}
					// }, 100)
					sendServer(func, tile, param)
				} else {
					// loginInfo.pushUTXO(utxo)
				}
			})
		})(this)
	}
}

Tile.prototype.Demolition = function() {
	menuClose();
	return this.headTile||this;
}
Tile.prototype.Industrial = function() {
	return this.Build(IndustrialType);
}
Tile.prototype.Residential = function() {
	return this.Build(ResidentialType);
}
Tile.prototype.Commercial = function() {
	return this.Build(CommercialType);
}
Tile.prototype.Upgrade = function() {
	return this.Build();
}
Tile.prototype.GetCoin = function(param) {
	var ps = param.split(":")
	if (ps.length === 3) {
		gGame.coin_list[ps[2]].height += 100000
		this.UI.removeCoin(ps[2])
	}
	return this;
}

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

function sendServer(func, tile, param) {
	var q = new SendQueue(func, tile, param)
	q.Enqueue()
	q.Do()
}

function SendQueue(func, tile, param) {
	this.func = func;
	this.tile = tile;
	this.param = param;
}
SendQueue.quere = [];
SendQueue.utxo = [];
SendQueue.NewUTXO = function (UTXO) {
	var send;
	if (!!(send = SendQueue.quere.splice(0, 1)) && send.length > 0) {
		send[0].sendServer(UTXO)
	} else {
		SendQueue.utxo.push(UTXO)
	}
}
SendQueue.prototype.Enqueue = function () {
	SendQueue.quere.push(this)
}
SendQueue.prototype.Do = function () {
	var utxo, send;
	while (!!(utxo = SendQueue.utxo.splice(0, 1)) && utxo.length > 0) {
		if (!!(send = SendQueue.quere.splice(0, 1)) && send.length > 0) {
			send[0].sendServer(utxo[0])
		} else {
			SendQueue.utxo.push(utxo[0])
			break;
		}
	}
}
SendQueue.prototype.sendServer = function(utxo) {
	var func = this.func;
	var tile = this.tile;
	var param = this.param;
	if (func == "Demolition") {
		$.ajax({
			type: "POST",
			url : "/api/games/"+loginInfo.Addr+"/commands/demolition",
			data : JSON.stringify({
				"utxo": utxo,
				"x": tile.x,
				"y": tile.y
			}),
			success : function (d) {
				if (typeof d === "string") {
					d = JSON.parse(d)
				}
				/*
					"type": 2,
					"tx_hex": TRANSACTION_HEX,
					"hash_hex": HASH_HEX
					*/
				commit(d)
			},
			error: function(d) {
				SendQueue.NewUTXO(utxo)
				Alert(language["Failed to execute demolation command"])
			}
		})
	} else if (func == "Upgrade" || func == "Commercial" || func == "Industrial" || func == "Residential" ) {
		var area_type = tile.type||buildingNum(func);
		{
			$.ajax({
				type: "POST",
				url : "/api/games/"+loginInfo.Addr+"/commands/upgrade",
				data : JSON.stringify({
					"utxo": utxo,
					"x": tile.x,
					"y": tile.y,
					"area_type": area_type,
					"target_level": +tile.level+1
				}),
				success : function (d) {
					if (typeof d === "string") {
						d = JSON.parse(d)
					}
					/*
						"type": 3,
						"tx_hex": TRANSACTION_HEX,
						"hash_hex": HASH_HEX
					*/
					commit(d)
				},
				error: function(d) {
					SendQueue.NewUTXO(utxo)
					Alert(language["Failed to execute upgrade command"])
				}
			})
		}
	} else if (func == "GetCoin") {
		var ps = param.split(":")
		if (ps.length == 3) {
			tile.x
			tile.y
			var coinType = ps[0]
			var height = ps[1]
			var hash = ps[2]


			$.ajax({
				type: "POST",
				url : "/api/games/"+loginInfo.Addr+"/commands/getcoin",
				data : JSON.stringify({
					"utxo": utxo,
					"x": tile.x,
					"y": tile.y,
					"coin_type": +coinType,
					"height": +height,
					"hash": hash,
				}),
				success : function (d) {
					if (typeof d === "string") {
						d = JSON.parse(d)
					}
					/*
						"type": 3,
						"tx_hex": TRANSACTION_HEX,
						"hash_hex": HASH_HEX
					*/
					commit(d)
				},
				error: function(d) {
					SendQueue.NewUTXO(utxo)
					Alert(language["Failed to execute upgrade command"])
				}
			})

		} else {
			alert("not Enough parameter")
		}
	}

}


function commit(data) {
	var msg = new Buffer(data.hash_hex, "hex");
	var sig = loginInfo.Key.sign(msg);
	var SIG_HEX = buf2hex(sig.r.toArrayLike(Buffer, "be", 32)) + buf2hex(sig.s.toArrayLike(Buffer, "be", 32)) + "0" + sig.recoveryParam;

	$.ajax({
		type: "POST",
		url : "/api/games/"+loginInfo.Addr+"/commands/commit",
		data : JSON.stringify({
			"type": data.type,
			"tx_hex": data.tx_hex,
			"sig_hex": SIG_HEX
		}),
		success : function (d) {
		},
		error: function(d) {
			if (data.type == 2) {
				Alert(language["Failed to execute demolation command"])
			} else if (data.type == 3) {
				Alert(language["Failed to execute upgrade command"])
			} else {
				Alert(language["commit error"])
			}
		}
	})

}