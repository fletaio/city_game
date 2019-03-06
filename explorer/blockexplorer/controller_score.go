package blockexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/kernel"
)

type ScoreController struct {
	kn *kernel.Kernel
	db *badger.DB
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
	var totalUsedBalance uint64
	if err := e.db.View(func(txn *badger.Txn) error {
		key := []byte("UsedBalance" + addr.String())
		item, err := txn.Get(key)
		if err != nil {
		} else {
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			totalUsedBalance = util.BytesToUint64(value)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	e.kn.Lock()
	bs := e.kn.Loader().AccountData(addr, []byte("game"))
	Height := e.kn.Provider().Height()
	ccbs := e.kn.Loader().AccountData(addr, []byte("GetCoinCount"))
	var coinCount uint32
	if len(ccbs) == 4 {
		coinCount = util.BytesToUint32(ccbs)
	}
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
		"Coin":        []byte(fmt.Sprintf("%v", coinCount)),
		"Gold":        []byte(fmt.Sprintf("%v", totalUsedBalance+gr.Balance)),
		"Population":  []byte(fmt.Sprintf("%v", gr.ManProvided)),
		"Electricity": []byte(fmt.Sprintf("%v", gr.PowerProvided)),
		"data":        data,
	}, nil
}
