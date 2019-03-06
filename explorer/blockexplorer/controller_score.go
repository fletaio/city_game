package blockexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/kernel"
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
		"Gold":        []byte(fmt.Sprintf("%v", gr.Balance)),
		"Population":  []byte(fmt.Sprintf("%v", gr.ManProvided)),
		"Electricity": []byte(fmt.Sprintf("%v", gr.PowerProvided)),
		"data":        data,
	}, nil
}

func (e *ScoreController) Coincount(r *http.Request) (map[string][]byte, error) {
	currHeight := e.kn.Provider().Height()
	m := map[string]map[string]int{}
	fields := map[string]bool{}
	for i := uint32(1); i <= currHeight; i++ {
		b, err := e.kn.Block(i)
		if err != nil {
			continue
		}
		txs := b.Body.Transactions
		for _, tx := range txs {
			switch tx := tx.(type) {
			case *citygame.GetCoinTx:
				cd := common.Coordinate{}
				buf := bytes.NewBuffer(tx.Address[:])
				cd.ReadFrom(buf)
				ob, err := e.kn.Block(cd.Height)
				if err != nil {
					continue
				}
				if uint16(len(ob.Body.Transactions)) > cd.Index {
					accTx := ob.Body.Transactions[cd.Index]
					switch accTx := accTx.(type) {
					case *citygame.CreateAccountTx:
						data, has := m[accTx.UserID]
						if has == false {
							data = map[string]int{}
							m[accTx.UserID] = data
						}
						location, _ := time.LoadLocation("Etc/GMT")
						hTime := time.Unix(0, int64(b.Header.Timestamp())).In(location)
						str := hTime.Format(time.RFC3339)
						var k string
						if str[:13] < str[:11]+"05" { //안지남
							hTime = hTime.Add(-time.Hour * 24)
							str2 := hTime.Format(time.RFC3339)
							k = str2[:10]
						} else { //지남
							k = str[:10]
						}
						fields[k] = true
						i, _ := data[k]
						i++
						data[k] = i
					}
				}
			}
		}
	}

	jm, _ := json.Marshal(m)
	jFields, _ := json.Marshal(fields)

	return map[string][]byte{
		"coin":   jm,
		"fields": jFields,
	}, nil
}
