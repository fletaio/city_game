package blockexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	citygame "git.fleta.io/fleta/city_game/city_game_context"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/kernel"
)

type GameController struct {
	kn *kernel.Kernel
}

type WebTile struct {
	AreaType    int `json:"area_type"`
	Level       int `json:"level"`
	BuildHeight int `json:"build_height"`
}

func (e *GameController) Index(r *http.Request) (map[string][]byte, error) {
	param := r.URL.Query()
	addrStr := param.Get("addr")
	addr, err := common.ParseAddress(addrStr)
	if err != nil {
		return nil, err
	}

	e.kn.Lock()
	bs := e.kn.Loader().AccountData(addr, []byte("game"))
	Height := e.kn.Provider().Height()
	e.kn.Unlock()
	log.Println("game/Index addr", addr.String())

	if len(bs) == 0 {
		return nil, citygame.ErrNotExistAccount
	}

	gd := citygame.NewGameData(Height)
	if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, err
	}

	Tiles := make([]*WebTile, len(gd.Tiles))
	for i, tile := range gd.Tiles {
		if tile != nil {
			Tiles[i] = &WebTile{
				AreaType:    int(tile.AreaType),
				Level:       int(tile.Level),
				BuildHeight: int(tile.BuildHeight),
			}
		}
	}

	data, _ := json.Marshal(Tiles)
	dmap, _ := json.Marshal(citygame.GBuildingDefine)

	return map[string][]byte{
		"Tiles":     data,
		"DefindMap": dmap,
		"Height":    []byte(fmt.Sprintf("%v", Height)),
	}, nil
}

func (e *GameController) User(r *http.Request) (map[string][]byte, error) {
	return nil, nil
}
