package blockexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"git.fleta.io/fleta/common"

	citygame "git.fleta.io/fleta/city_game/city_game_context"

	"git.fleta.io/fleta/core/kernel"
)

type ScoreController struct {
	kn *kernel.Kernel
}

func (e *ScoreController) All(r *http.Request) (map[string][]byte, error) {
	param := r.URL.Query()
	sort := param.Get("sort")
	keyword := param.Get("keyword")

	return map[string][]byte{
		"sort":    []byte(sort),
		"keyword": []byte(keyword),
	}, nil
}

func (e *ScoreController) User(r *http.Request) (map[string][]byte, error) {
	param := r.URL.Query()
	addrStr := param.Get("addr")
	userid := param.Get("userid")

	addr := common.MustParseAddress(addrStr)

	e.kn.Lock()
	bs := e.kn.Loader().AccountData(addr, []byte("game"))
	Height := e.kn.Provider().Height()
	e.kn.Unlock()

	if len(bs) == 0 {
		return nil, citygame.ErrNotExistAccount
	}

	gd := citygame.NewGameData(Height)
	if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, err
	}

	gr := gd.Resource(Height)
	data, _ := json.Marshal(gd)

	return map[string][]byte{
		"ID":          []byte(userid),
		"Addr":        []byte(addrStr),
		"Gold":        []byte(fmt.Sprintf("%v", gr.Balance)),
		"Population":  []byte(fmt.Sprintf("%v", gr.ManProvided)),
		"Electricity": []byte(fmt.Sprintf("%v", gr.PowerProvided)),
		"data":        data,
	}, nil
}
