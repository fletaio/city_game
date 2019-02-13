Tile.prototype.RunCommand = function(func) {
	if (typeof this[func] === "function") {
		message("command : "+ func + " x : " + this.x + " y : " + this.y );
		this[func]();
		menuOpen(this);
		if (this.obj.level == 6) {
			return this.obj.headTile
		}
	}
	return this
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
	this.Build("Industrial");
	return this;
}
Tile.prototype.Residential = function() {
	this.Build("Residential");
	return this;
}
Tile.prototype.Commercial = function() {
	this.Build("Commercial");
	return this;
}
Tile.prototype.Upgrade = function() {
	this.Build();
	return this;
}
