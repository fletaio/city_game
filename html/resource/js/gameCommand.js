Tile.prototype.RunCommand = function(func) {
	var tile = this;
	if (typeof this[func] === "function") {
		message("command : "+ func + " x : " + this.x + " y : " + this.y );
		tile = this[func]();
		menuOpen(tile);
		sendServer(func, tile)
	}

	return tile
}

Tile.prototype.Demolition = function() {
	if (this.obj.level == 6) {
		var checker = this.CheckLvRound(6)
		for ( var i = 0 ; i < checker.candidate.length ; i++ ) {
			Tiles[checker.candidate[i]].Remove().UpdateInfo();
		}
	} else {
		this.Remove();
	}
	menuClose();
	return this;
}
Tile.prototype.Industrial = function() {
	return this.Build("Industrial");
}
Tile.prototype.Residential = function() {
	return this.Build("Residential");
}
Tile.prototype.Commercial = function() {
	return this.Build("Commercial");
}
Tile.prototype.Upgrade = function() {
	return this.Build();
}

function sendServer(func, tile) {
	switch (func) {
		case "Demolition":
		$.ajax({
			type: "POST",
			url : "/api/games/"+loginInfo.Addr+"/commands/demolition",
			data : {
				"seq": loginInfo.Seq,
				"x": tile.x,
				"y": tile.y
			},
			success : function (d) {
				/*
					"type": 2,
					"tx_hex": TRANSACTION_HEX,
					"hash_hex": HASH_HEX
				 */
				d = JSON.parse(d)
				commit(d)
			},
			error: function(d) {
				alert("error")
			}
		})
	
			break;
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

		case "Upgrade":
		var tileType = tile.Type;
		case "Commercial":
		case "Industrial":
		case "Residential":
		var tileType = tileType|func

		switch (tileType) {
			case "Commercial":
				var areaType = 1
				break;
			case "Industrial":
				var areaType = 2
				break;
			case "Residential":
				var areaType = 3
				break;
		}
		{
			$.ajax({
				type: "POST",
				url : "/api/games/"+loginInfo.Addr+"/commands/upgrade",
				data : {
					"seq": loginInfo.Seq,
					"x": tile.x,
					"y": tile.y,
					"area_type": areaType,
					"target_level": tile.obj.level
				},
				success : function (d) {
					/*
						"type": 3,
						"tx_hex": TRANSACTION_HEX,
						"hash_hex": HASH_HEX
					 */
					d = JSON.parse(d)
					commit(d)
				},
				error: function(d) {
					alert("error")
				}
			})
		}

		default:
			break;
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
		data : {
			"type": data.type,
			"tx_hex": data.tx_hex,
			"sig_hex": SIG_HEX
		},
		success : function (d) {
			loginInfo.Seq++
		},
		error: function(d) {
			alert("commit error")
		}
	})

}