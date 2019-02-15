Tile.prototype.RunCommand = function(func) {
	var tile = this;
	if (typeof this[func] === "function") {
		message("command : "+ func + " x : " + this.x + " y : " + this.y );

		var utxo = loginInfo.popUTXO()
		if (!utxo) {
			alert(language["too fast"])
			return 
		}
		tile = this[func]();
		if (IsTile(tile)){
			menuOpen(tile);
			// setTimeout(function () {
			// 	var balance = parseInt($("#dollar[key='balance']").html())
			// 	balance -=  gBuildingDefine[tile.type][tile.obj.level].cost_usage
			// 	var height = gGame.height
			// 	if (func == "Demolition") {
			// 		onMessage({_init:true}, {data : "{\"point_height\":"+height+",\"point_balance\":"+balance+",\"x\":"+(tile.x)+",\"y\":"+(tile.y)+",\"area_type\":0,\"level\":0,\"type\":0,\"height\":"+gGame.height+"}"})
			// 	} else {
			// 		onMessage({_init:true}, {data : "{\"point_height\":"+height+",\"point_balance\":"+balance+",\"x\":"+(tile.x)+",\"y\":"+(tile.y)+",\"area_type\":"+tile.type+",\"level\":"+(tile.obj.level+1)+",\"type\":1,\"height\":"+gGame.height+"}"})
			// 	}
			// }, 100)
			sendServer(func, tile, utxo)
		} else {
			loginInfo.pushUTXO(utxo)
		}
	}
	return tile
}

Tile.prototype.Demolition = function() {
	menuClose();
	return this.obj.headTile||this;
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

function sendServer(func, tile, utxo) {
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
				loginInfo.pushUTXO(utxo)
				alert("error")
			}
		})
	} else if (func == "Upgrade" || func == "Commercial" || func == "Industrial" || func == "Residential" ) {
/*
	/api/games/ADDRESS_STRING/commands/upgrade
	* REQUEST
	{
		"seq": NEXT_SEQ_INT,
		"x": X_INT,
		"y": Y_INT,
		"area_type": TYPE_INT(1 : Commercial, 2 : Industrial, 3 : Residential),
		"target_level": LEVEL_INT
	}
*/
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
					"target_level": tile.obj.level
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
					loginInfo.pushUTXO(utxo)
					alert("error")
				}
			})
		}
	}

}


function commit(data) {
/*
	/api/games/ADDRESS_STRING/commands/commit
	* REQUEST
	{
		"type": TYPE_INT,
		"tx_hex": TRANSACTION_HEX,
		"sig_hex": SIGNATURE_HEX
	}
*/
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
			alert("commit error")
		}
	})

}