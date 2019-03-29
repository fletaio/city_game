package citygame

import (
	"io"
	"sort"

	"github.com/fletaio/common/util"
)

// consts
const (
	GTileSize     = 32
	GExpireHeight = 86400 * 2 * 5 // 5 days
)

// AreaType is a area type of the target tile
type AreaType uint8

// CoinType is a coin type of the construct or time
type CoinType uint8

// area types
const (
	EmptyAreaType       = AreaType(0)
	CommercialAreaType  = AreaType(1)
	IndustrialAreaType  = AreaType(2)
	ResidentialAreaType = AreaType(3)
	EndOfAreaType       = AreaType(4)
)

// TimeCoinGenTime id define coin regenerate time
const TimeCoinGenTime = uint32(0.5 * 2 * 60 * 5) //blocktime * 1/blocktime * 1minute * 5

// Tile reprents a information of the target tile
type Tile struct {
	AreaType    AreaType
	Level       uint8
	BuildHeight uint32
}

// WriteTo is a serialization function
func (tile *Tile) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint8(w, uint8(tile.AreaType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, tile.Level); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, tile.BuildHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tile *Tile) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tile.AreaType = AreaType(v)
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tile.Level = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		tile.BuildHeight = v
	}
	return read, nil
}

// NewTile returns a Tile
func NewTile(AreaType AreaType, BuildHeight uint32) *Tile {
	tile := &Tile{
		AreaType:    AreaType,
		Level:       1,
		BuildHeight: BuildHeight,
	}
	return tile
}

// Resource is a status of the game data
type Resource struct {
	Balance       uint64
	PowerRemained uint32
	PowerProvided uint32
	ManRemained   uint32
	ManProvided   uint32
}

// GameData stores all data of the game
type GameData struct {
	PointBalance uint64
	PointHeight  uint32
	CoinCount    uint32
	TotalExp     uint64
	Coins        []*FletaCityCoin
	Exps         []*FletaCityExp
	MaxLevels    []uint8
	Tiles        []*Tile
}

// NewGameData returns a GameData
func NewGameData(TargetHeight uint32) *GameData {
	gd := &GameData{
		PointBalance: 900,
		PointHeight:  TargetHeight,
		Coins:        []*FletaCityCoin{},
		Exps:         []*FletaCityExp{},
		MaxLevels:    make([]uint8, GTileSize*GTileSize),
		Tiles:        make([]*Tile, GTileSize*GTileSize),
	}
	return gd
}

// Count returns the building count by a area type
func (gd *GameData) Count() map[AreaType]int {
	CountMap := map[AreaType]int{}
	for _, tile := range gd.Tiles {
		if tile != nil {
			CountMap[tile.AreaType]++
		}
	}
	return CountMap
}

// Resource returns remained resources by the target height
func (gd *GameData) Resource(TargetHeight uint32) *Resource {
	ForwardHeight := TargetHeight - gd.PointHeight
	base := &Resource{
		Balance:       4,
		PowerRemained: 5,
		PowerProvided: 5,
		ManRemained:   3,
		ManProvided:   3,
	}
	used := &Resource{}
	provide := &Resource{
		Balance:       gd.PointBalance + uint64(base.Balance/2)*uint64(ForwardHeight),
		PowerRemained: base.PowerRemained,
		ManRemained:   base.ManRemained,
	}
	for _, tile := range gd.Tiles {
		if tile != nil {
			bd := GBuildingDefine[tile.AreaType][tile.Level-1]
			used.ManRemained += bd.AccManUsage
			used.PowerRemained += bd.AccPowerUsage
			ConstructionHeight := tile.BuildHeight + bd.BuildTime*2

			if tile.Level == 6 {
				bd2 := GBuildingDefine[tile.AreaType][tile.Level-2]
				used.ManRemained += bd2.AccManUsage * 3
				used.PowerRemained += bd2.AccPowerUsage * 3
			}
			if TargetHeight < ConstructionHeight {
				if tile.Level == 1 {
					continue
				} else if tile.Level == 6 {
					bd2 := GBuildingDefine[tile.AreaType][tile.Level-2]
					switch tile.AreaType {
					case CommercialAreaType:
						provide.Balance += uint64(bd2.Output/2) * uint64(ForwardHeight) * 3
					case IndustrialAreaType:
						provide.PowerRemained += bd2.Output * 3
					case ResidentialAreaType:
						provide.ManRemained += bd2.Output * 3
					}
				}
				bd = GBuildingDefine[tile.AreaType][tile.Level-2]
			}
			switch tile.AreaType {
			case CommercialAreaType:
				if ConstructionHeight <= gd.PointHeight {
					provide.Balance += uint64(bd.Output/2) * uint64(ForwardHeight)
				} else {
					if tile.BuildHeight <= gd.PointHeight {
						provide.Balance += uint64(bd.Output/2) * uint64(ForwardHeight-(ConstructionHeight-gd.PointHeight))
					} else {
						provide.Balance += uint64(bd.Output/2) * uint64(TargetHeight-ConstructionHeight)
						if tile.Level > 1 {
							prevbd := GBuildingDefine[tile.AreaType][tile.Level-2]
							provide.Balance += uint64(prevbd.Output/2) * uint64(tile.BuildHeight-gd.PointHeight)
						}
					}
				}
			case IndustrialAreaType:
				provide.PowerRemained += bd.Output
			case ResidentialAreaType:
				provide.ManRemained += bd.Output
			}
		}
	}
	res := &Resource{
		Balance:       provide.Balance,
		PowerRemained: provide.PowerRemained - used.PowerRemained,
		PowerProvided: provide.PowerRemained,
		ManRemained:   provide.ManRemained - used.ManRemained,
		ManProvided:   provide.ManRemained,
	}
	return res
}

// UpdatePoint updates the last point of the game data
func (gd *GameData) UpdatePoint(TargetHeight uint32, Balance uint64) {
	gd.PointHeight = TargetHeight
	gd.PointBalance = Balance
}

// WriteTo is a serialization function
func (gd *GameData) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint64(w, gd.PointBalance); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, gd.PointHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, gd.CoinCount); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, gd.TotalExp); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint16(w, uint16(len(gd.Coins))); err != nil {
		return 0, err
	} else {
		wrote += n
		for _, c := range gd.Coins {
			if n, err := c.WriteTo(w); err != nil {
				return 0, err
			} else {
				wrote += n
			}
		}
	}
	if n, err := util.WriteUint16(w, uint16(len(gd.Exps))); err != nil {
		return 0, err
	} else {
		wrote += n
		for _, e := range gd.Exps {
			if n, err := e.WriteTo(w); err != nil {
				return 0, err
			} else {
				wrote += n
			}
		}
	}
	if n, err := util.WriteUint16(w, uint16(len(gd.MaxLevels))); err != nil {
		return 0, err
	} else {
		wrote += n
		for _, v := range gd.MaxLevels {
			if n, err := util.WriteUint8(w, v); err != nil {
				return 0, err
			} else {
				wrote += n
			}
		}
	}

	coords := []int{}
	tileMap := map[int]*Tile{}
	for coord, tile := range gd.Tiles {
		if tile != nil {
			coords = append(coords, coord)
			tileMap[coord] = tile
		}
	}
	sort.Ints(coords)

	if n, err := util.WriteUint32(w, uint32(len(coords))); err != nil {
		return wrote, err
	} else {
		wrote += n
		for _, coord := range coords {
			if n, err := util.WriteUint32(w, uint32(coord)); err != nil {
				return wrote, err
			} else {
				wrote += n
				tile := tileMap[coord]
				if n, err := tile.WriteTo(w); err != nil {
					return wrote, err
				} else {
					wrote += n
				}
			}
		}
	}

	return wrote, nil
}

// ReadFrom is a deserialization function
func (gd *GameData) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		gd.PointBalance = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		gd.PointHeight = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		gd.CoinCount = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		gd.TotalExp = v
	}
	if Len, n, err := util.ReadUint16(r); err != nil {
		return read, err
	} else {
		read += n
		gd.Coins = make([]*FletaCityCoin, 0, Len)
		for i := 0; i < int(Len); i++ {
			c := &FletaCityCoin{}
			if n, err := c.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
			}
			gd.Coins = append(gd.Coins, c)
		}
	}
	if Len, n, err := util.ReadUint16(r); err != nil {
		return read, err
	} else {
		read += n
		gd.Exps = make([]*FletaCityExp, 0, Len)
		for i := 0; i < int(Len); i++ {
			e := &FletaCityExp{}
			if n, err := e.ReadFrom(r); err != nil {
				return read, err
			} else {
				read += n
			}
			gd.Exps = append(gd.Exps, e)
		}
	}
	if Len, n, err := util.ReadUint16(r); err != nil {
		return read, err
	} else {
		read += n
		gd.MaxLevels = make([]uint8, 0, Len)
		for i := 0; i < int(Len); i++ {
			if v, n, err := util.ReadUint8(r); err != nil {
				return read, err
			} else {
				read += n
				gd.MaxLevels = append(gd.MaxLevels, v)
			}
		}
	}

	if Len, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		gd.Tiles = make([]*Tile, GTileSize*GTileSize)
		for i := 0; i < int(Len); i++ {
			if coord, n, err := util.ReadUint32(r); err != nil {
				return read, err
			} else {
				read += n
				tile := &Tile{}
				if n, err := tile.ReadFrom(r); err != nil {
					return read, err
				} else {
					read += n
					gd.Tiles[coord] = tile
				}
			}
		}
	}
	return read, nil
}

// BuildingDefine defines a building information
type BuildingDefine struct {
	CostUsage     uint64 `json:"cost_usage"`      // all
	BuildTime     uint32 `json:"build_time"`      // all
	Output        uint32 `json:"output"`          // all
	Exp           uint32 `json:"exp"`             // all
	ManUsage      uint32 `json:"man_usage"`       // commerial, industrial
	PowerUsage    uint32 `json:"power_usage"`     // commerial, residential
	AccManUsage   uint32 `json:"acc_man_usage"`   // commerial, industrial
	AccPowerUsage uint32 `json:"acc_power_usage"` // commerial, residential
}

func init() {
	for _, bds := range GBuildingDefine {
		var AccManUsage uint32
		var AccPowerUsage uint32
		for _, bd := range bds {
			AccManUsage += bd.ManUsage
			AccPowerUsage += bd.PowerUsage
			bd.AccManUsage = AccManUsage
			bd.AccPowerUsage = AccPowerUsage
		}
	}
}

// GBuildingDefine is ingame construction data
var GBuildingDefine = map[AreaType][]*BuildingDefine{
	CommercialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  400,
			BuildTime:  1,
			Output:     400,
			Exp:        1,
			ManUsage:   2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  2400,
			BuildTime:  1,
			Output:     900,
			Exp:        2,
			ManUsage:   3,
			PowerUsage: 5,
		},
		&BuildingDefine{
			CostUsage:  12000,
			BuildTime:  7,
			Output:     1400,
			Exp:        3,
			ManUsage:   5,
			PowerUsage: 8,
		},
		&BuildingDefine{
			CostUsage:  60000,
			BuildTime:  35,
			Output:     2100,
			Exp:        4,
			ManUsage:   8,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  300000,
			BuildTime:  180,
			Output:     3200,
			Exp:        5,
			ManUsage:   12,
			PowerUsage: 18,
		},
		&BuildingDefine{
			CostUsage:  6000000,
			BuildTime:  864,
			Output:     19200,
			Exp:        6,
			ManUsage:   72,
			PowerUsage: 108,
		},
	},
	IndustrialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage: 200,
			BuildTime: 1,
			Output:    5,
			Exp:       1,
			ManUsage:  1,
		},
		&BuildingDefine{
			CostUsage: 1200,
			BuildTime: 2,
			Output:    11,
			Exp:       2,
			ManUsage:  2,
		},
		&BuildingDefine{
			CostUsage: 6000,
			BuildTime: 7,
			Output:    17,
			Exp:       3,
			ManUsage:  3,
		},
		&BuildingDefine{
			CostUsage: 30000,
			BuildTime: 27,
			Output:    26,
			Exp:       4,
			ManUsage:  5,
		},
		&BuildingDefine{
			CostUsage: 150000,
			BuildTime: 120,
			Output:    39,
			Exp:       5,
			ManUsage:  8,
		},
		&BuildingDefine{
			CostUsage: 3000000,
			BuildTime: 570,
			Output:    234,
			Exp:       6,
			ManUsage:  48,
		},
	},
	ResidentialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  300,
			BuildTime:  1,
			Output:     2,
			Exp:        1,
			PowerUsage: 2,
		},
		&BuildingDefine{
			CostUsage:  1800,
			BuildTime:  1,
			Output:     5,
			Exp:        2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  9000,
			BuildTime:  7,
			Output:     8,
			Exp:        3,
			PowerUsage: 5,
		},
		&BuildingDefine{
			CostUsage:  45000,
			BuildTime:  32,
			Output:     12,
			Exp:        4,
			PowerUsage: 8,
		},
		&BuildingDefine{
			CostUsage:  225000,
			BuildTime:  150,
			Output:     18,
			Exp:        5,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  4500000,
			BuildTime:  720,
			Output:     108,
			Exp:        6,
			PowerUsage: 72,
		},
	},
}

// ExpDefine defines a exp level
type ExpDefine struct {
	Level   uint8  `json:"level"`
	Exp     uint64 `json:"exp"`
	AccExp  uint64 `json:"acc_exp"`
	Class   string `json:"class"`
	CoinGen uint8  `json:"coin_gen"`
}

var GExpDefine = []*ExpDefine{
	&ExpDefine{Level: 1, Exp: 0, AccExp: 0, Class: "lv_bronze", CoinGen: 5},
	&ExpDefine{Level: 2, Exp: 10, AccExp: 10, Class: "lv_bronze", CoinGen: 5},
	&ExpDefine{Level: 3, Exp: 15, AccExp: 25, Class: "lv_bronze", CoinGen: 5},
	&ExpDefine{Level: 4, Exp: 20, AccExp: 45, Class: "lv_bronze", CoinGen: 5},
	&ExpDefine{Level: 5, Exp: 25, AccExp: 70, Class: "lv_bronze", CoinGen: 5},
	&ExpDefine{Level: 6, Exp: 30, AccExp: 100, Class: "lv_bronze", CoinGen: 7},
	&ExpDefine{Level: 7, Exp: 50, AccExp: 150, Class: "lv_bronze", CoinGen: 7},
	&ExpDefine{Level: 8, Exp: 70, AccExp: 220, Class: "lv_silver", CoinGen: 7},
	&ExpDefine{Level: 9, Exp: 100, AccExp: 320, Class: "lv_silver", CoinGen: 7},
	&ExpDefine{Level: 10, Exp: 140, AccExp: 460, Class: "lv_silver", CoinGen: 7},
	&ExpDefine{Level: 11, Exp: 190, AccExp: 650, Class: "lv_silver", CoinGen: 10},
	&ExpDefine{Level: 12, Exp: 250, AccExp: 900, Class: "lv_silver", CoinGen: 10},
	&ExpDefine{Level: 13, Exp: 320, AccExp: 1220, Class: "lv_silver", CoinGen: 10},
	&ExpDefine{Level: 14, Exp: 400, AccExp: 1620, Class: "lv_silver", CoinGen: 10},
	&ExpDefine{Level: 15, Exp: 500, AccExp: 2120, Class: "lv_gold", CoinGen: 10},
	&ExpDefine{Level: 16, Exp: 700, AccExp: 2820, Class: "lv_gold", CoinGen: 13},
	&ExpDefine{Level: 17, Exp: 800, AccExp: 3620, Class: "lv_gold", CoinGen: 13},
	&ExpDefine{Level: 18, Exp: 1000, AccExp: 4620, Class: "lv_gold", CoinGen: 13},
	&ExpDefine{Level: 19, Exp: 1500, AccExp: 6120, Class: "lv_gold", CoinGen: 13},
	&ExpDefine{Level: 20, Exp: 2000, AccExp: 8120, Class: "lv_gold", CoinGen: 13},
	&ExpDefine{Level: 21, Exp: 2500, AccExp: 10620, Class: "lv_gold", CoinGen: 15},
	&ExpDefine{Level: 22, Exp: 3000, AccExp: 13620, Class: "lv_fleta", CoinGen: 15},
	&ExpDefine{Level: 23, Exp: 3500, AccExp: 17120, Class: "lv_fleta", CoinGen: 15},
	&ExpDefine{Level: 24, Exp: 4000, AccExp: 21120, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 25, Exp: 4500, AccExp: 25620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 26, Exp: 5000, AccExp: 30620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 27, Exp: 5500, AccExp: 36120, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 28, Exp: 6000, AccExp: 42120, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 29, Exp: 6500, AccExp: 48620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 30, Exp: 7000, AccExp: 55620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 31, Exp: 7500, AccExp: 63120, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 32, Exp: 8000, AccExp: 71120, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 33, Exp: 8500, AccExp: 79620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 34, Exp: 9000, AccExp: 88620, Class: "lv_fleta", CoinGen: 20},
	&ExpDefine{Level: 35, Exp: 9500, AccExp: 98120, Class: "lv_fleta", CoinGen: 20},
}
