var gConfig = {
	Unit: 64,
	Size: 16,
}

$(document).on('click', '#touchpad', function(e) {
	var point = {x:0, y:0};
	if(e.type == 'touchstart' || e.type == 'touchmove' || e.type == 'touchend' || e.type == 'touchcancel'){
		var touch = e.originalEvent.touches[0] || e.originalEvent.changedTouches[0];
		point.x = touch.pageX;
		point.y = touch.pageY;
	} else if (e.type=='click' || e.type == 'mousedown' || e.type == 'mouseup' || e.type == 'mousemove' || e.type == 'mouseover'|| e.type=='mouseout' || e.type=='mouseenter' || e.type=='mouseleave') {
		point.x = e.pageX;
		point.y = e.pageY;
	}

	var jScreen = $("#screen");
	var top = parseInt(jScreen.css("top"));

	var a = point.x*2/gConfig.Unit - gConfig.Size;
	var b = (point.y - top/2)*4/gConfig.Unit - 2;

	var x = Math.floor((a+b)/2) - 1;
	var y = Math.floor((b-a)/2) - 1;

	if(0 <= x && x < gConfig.Size && 0 <= y && y < gConfig.Size) {
		var tile = Tiles[x + y *gConfig.Size];
		tile.Build();
	}
});

function Tile(jScreen, x, y, num) {
	this.x = x;
	this.y = y;
	this.obj = $("<div/>").appendTo(jScreen);
	this.obj.css("position", "absolute");
	this.obj.css("z-index", x+y);
	this.obj.append($("<image src='tile/base_floor/groundtiles_tile"+num+".png' style='position:absolute; width:"+gConfig.Unit+"px; bottom:0px;'/>"));
	this.obj.level = 0;
	this.Resize();
}

Tile.prototype.Build = function() {
	switch(this.obj.level) {
	case 0:
		this.obj.level = 1;
		this.obj.children().eq(0).attr("src", "tile/building_floor.png");
		//var jImg = $("<image class='building' src='building/construction.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		var jImg = $("<image class='building' src='building/Industrial_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit/4)+"px");
		jImg.css("bottom", (gConfig.Unit/4)+"px");
		jImg.css("z-index", 1);
		break;
	case 1:
		this.obj.level = 2;
		var jImg = $("<image class='building' src='building/Industrial_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit*2/4)+"px");
		jImg.css("bottom", (gConfig.Unit/2/4)+"px");
		jImg.css("z-index", 2);
		break;
	case 2:
		this.obj.level = 3;
		var jImg = $("<image class='building' src='building/Industrial_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("left", (gConfig.Unit/4)+"px");
		jImg.css("z-index", 4);
		break;
	case 3:
		this.obj.level = 4;
		var jImg = $("<image class='building' src='building/Industrial_Lv1.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit/2)+"px");
		jImg.css("bottom", (gConfig.Unit/2/4)+"px");
		jImg.css("z-index", 3);
		break;
	case 4:
		this.obj.level = 5;
		this.obj.find(".building").detach();
		var jImg = $("<image class='building' src='building/Industrial_Lv5.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj);
		jImg.css("width", (gConfig.Unit)+"px");
		break;
	case 5:
		this.obj.level = 6;
		this.obj.find(".building").detach();
		var jImg = $("<image class='building' src='building/Biz_LvFLETA.png' style='position:absolute; bottom:0px;'/>").appendTo(this.obj)
		jImg.css("width", (gConfig.Unit*2)+"px");
		jImg.css("bottom", -(gConfig.Unit/4)+"px");
		break;
	}
};

Tile.prototype.Resize = function() {
	this.obj.css("width", gConfig.Unit+"px");
	this.obj.css("left", (gConfig.Unit*(gConfig.Size+this.x-this.y-1)/2) + "px");
	this.obj.css("bottom", gConfig.Unit*gConfig.Size/2 - (gConfig.Unit*(this.x+this.y+2)/2)/2 + "px");
	this.obj.children().eq(0).css("width", gConfig.Unit+"px");

	switch(this.obj.level) {
	case 6:
		//TODO
	case 5:
		var jBuilding = this.obj.find(".building").eq(0);
		jBuilding.css("width", (gConfig.Unit)+"px");
		break;
	case 4:
		var jBuilding = this.obj.find(".building").eq(3);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("bottom", (gConfig.Unit/2/4)+"px");
	case 3:
		var jBuilding = this.obj.find(".building").eq(2);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit/4)+"px");
	case 2:
		var jBuilding = this.obj.find(".building").eq(1);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit*2/4)+"px");
		jBuilding.css("bottom", (gConfig.Unit/2/4)+"px");
	case 1:
		var jBuilding = this.obj.find(".building").eq(0);
		jBuilding.css("width", (gConfig.Unit/2)+"px");
		jBuilding.css("left", (gConfig.Unit/4)+"px");
		jBuilding.css("bottom", (gConfig.Unit/4)+"px");
		break;
	}
}


var Tiles = [];
var jScreen = $("#screen");
jScreen.css("width", (gConfig.Unit*gConfig.Size)+"px");
jScreen.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
jScreen.css("top", (gConfig.Unit/2*4)+"px");
var jTouchPad = $("#touchpad");
jTouchPad.css("width", (gConfig.Unit*gConfig.Size)+"px");
jTouchPad.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
jTouchPad.css("top", (gConfig.Unit/2*4)+"px");
for(var i=0; i<gConfig.Size*gConfig.Size; i++) {
	var x = i%gConfig.Size;
	var y = parseInt(i/gConfig.Size);
	var num = (parseInt(Math.log2((x+1)*73)*100 + Math.log10((y+1)*4321)*100)%10+1);
	Tiles.push(new Tile(jScreen, x, y, num));
}

function ChangeUnit(unit) {
	gConfig.Unit = unit;

	var jScreen = $("#screen");
	jScreen.css("width", (gConfig.Unit*gConfig.Size)+"px");
	jScreen.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
	jScreen.css("top", (gConfig.Unit/2*4)+"px");
	var jTouchPad = $("#touchpad");
	jTouchPad.css("width", (gConfig.Unit*gConfig.Size)+"px");
	jTouchPad.css("height", (gConfig.Unit*gConfig.Size)/2+"px");
	jTouchPad.css("top", (gConfig.Unit/2*4)+"px");

	for(var i=0; i<Tiles.length; i++) {
		Tiles[i].Resize();
	}
}