package citygame

import (
	"io"
	"sort"

	"github.com/fletaio/common/util"
)

// consts
const (
	GTileSize = 32
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
			if tile.Level == 6 && TargetHeight > ConstructionHeight {
				bd2 := GBuildingDefine[tile.AreaType][tile.Level-2]
				used.ManRemained += bd2.AccManUsage * 3
				used.PowerRemained += bd2.AccPowerUsage * 3
			}

			if TargetHeight < ConstructionHeight {
				if tile.Level == 1 {
					continue
				}
				bd = GBuildingDefine[tile.AreaType][tile.Level-2]
			}
			switch tile.AreaType {
			case CommercialAreaType:
				if TargetHeight > ConstructionHeight {
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
/*
var GBuildingDefine = map[AreaType][]*BuildingDefine{
	CommercialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  400,
			BuildTime:  30,
			Output:     4,
			Exp:        1,
			ManUsage:   2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  2400,
			BuildTime:  140,
			Output:     10,
			Exp:        2,
			ManUsage:   3,
			PowerUsage: 4,
		},
		&BuildingDefine{
			CostUsage:  12000,
			BuildTime:  700,
			Output:     24,
			Exp:        3,
			ManUsage:   8,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  60000,
			BuildTime:  3500,
			Output:     64,
			Exp:        4,
			ManUsage:   40,
			PowerUsage: 30,
		},
		&BuildingDefine{
			CostUsage:  300000,
			BuildTime:  18000,
			Output:     160,
			Exp:        5,
			ManUsage:   200,
			PowerUsage: 80,
		},
		&BuildingDefine{
			CostUsage:  6000000,
			BuildTime:  86400,
			Output:     1600,
			Exp:        6,
			ManUsage:   4000,
			PowerUsage: 1500,
		},
	},
	IndustrialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage: 200,
			BuildTime: 60,
			Output:    5,
			Exp:       1,
			ManUsage:  1,
		},
		&BuildingDefine{
			CostUsage: 1700,
			BuildTime: 200,
			Output:    14,
			Exp:       2,
			ManUsage:  2,
		},
		&BuildingDefine{
			CostUsage: 12000,
			BuildTime: 700,
			Output:    96,
			Exp:       3,
			ManUsage:  8,
		},
		&BuildingDefine{
			CostUsage: 80000,
			BuildTime: 2700,
			Output:    390,
			Exp:       4,
			ManUsage:  54,
		},
		&BuildingDefine{
			CostUsage: 450000,
			BuildTime: 12000,
			Output:    1440,
			Exp:       5,
			ManUsage:  300,
		},
		&BuildingDefine{
			CostUsage: 9100000,
			BuildTime: 57000,
			Output:    33000,
			Exp:       6,
			ManUsage:  6100,
		},
	},
	ResidentialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  300,
			BuildTime:  45,
			Output:     3,
			Exp:        1,
			PowerUsage: 2,
		},
		&BuildingDefine{
			CostUsage:  2000,
			BuildTime:  170,
			Output:     10,
			Exp:        2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  12000,
			BuildTime:  700,
			Output:     64,
			Exp:        3,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  66000,
			BuildTime:  3200,
			Output:     564,
			Exp:        4,
			PowerUsage: 35,
		},
		&BuildingDefine{
			CostUsage:  360000,
			BuildTime:  15000,
			Output:     4000,
			Exp:        5,
			PowerUsage: 100,
		},
		&BuildingDefine{
			CostUsage:  7200000,
			BuildTime:  72000,
			Output:     101000,
			Exp:        6,
			PowerUsage: 1800,
		},
	},
}
*/

var GBuildingDefine = map[AreaType][]*BuildingDefine{
	CommercialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  1,
			BuildTime:  1,
			Output:     4,
			Exp:        1,
			ManUsage:   2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  2,
			BuildTime:  1,
			Output:     10,
			Exp:        2,
			ManUsage:   3,
			PowerUsage: 4,
		},
		&BuildingDefine{
			CostUsage:  3,
			BuildTime:  1,
			Output:     24,
			Exp:        3,
			ManUsage:   8,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  4,
			BuildTime:  1,
			Output:     64,
			Exp:        4,
			ManUsage:   40,
			PowerUsage: 30,
		},
		&BuildingDefine{
			CostUsage:  5,
			BuildTime:  1,
			Output:     160,
			Exp:        5,
			ManUsage:   200,
			PowerUsage: 80,
		},
		&BuildingDefine{
			CostUsage:  6,
			BuildTime:  1,
			Output:     1600,
			Exp:        6,
			ManUsage:   4000,
			PowerUsage: 1500,
		},
	},
	IndustrialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage: 1,
			BuildTime: 1,
			Output:    5,
			Exp:       1,
			ManUsage:  1,
		},
		&BuildingDefine{
			CostUsage: 2,
			BuildTime: 1,
			Output:    14,
			Exp:       2,
			ManUsage:  2,
		},
		&BuildingDefine{
			CostUsage: 3,
			BuildTime: 1,
			Output:    96,
			Exp:       3,
			ManUsage:  8,
		},
		&BuildingDefine{
			CostUsage: 4,
			BuildTime: 1,
			Output:    390,
			Exp:       4,
			ManUsage:  54,
		},
		&BuildingDefine{
			CostUsage: 5,
			BuildTime: 1,
			Output:    1440,
			Exp:       5,
			ManUsage:  300,
		},
		&BuildingDefine{
			CostUsage: 6,
			BuildTime: 1,
			Output:    33000,
			Exp:       6,
			ManUsage:  6100,
		},
	},
	ResidentialAreaType: []*BuildingDefine{
		&BuildingDefine{
			CostUsage:  1,
			BuildTime:  1,
			Output:     3,
			Exp:        1,
			PowerUsage: 2,
		},
		&BuildingDefine{
			CostUsage:  2,
			BuildTime:  1,
			Output:     10,
			Exp:        2,
			PowerUsage: 3,
		},
		&BuildingDefine{
			CostUsage:  3,
			BuildTime:  1,
			Output:     64,
			Exp:        3,
			PowerUsage: 12,
		},
		&BuildingDefine{
			CostUsage:  4,
			BuildTime:  1,
			Output:     564,
			Exp:        4,
			PowerUsage: 35,
		},
		&BuildingDefine{
			CostUsage:  5,
			BuildTime:  1,
			Output:     4000,
			Exp:        5,
			PowerUsage: 100,
		},
		&BuildingDefine{
			CostUsage:  6,
			BuildTime:  1,
			Output:     101000,
			Exp:        6,
			PowerUsage: 1800,
		},
	},
}
