package cityexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/core/kernel"
)

type ScoreController struct {
	kn *kernel.Kernel
}

func (e *ScoreController) All(r *http.Request) (map[string]string, error) {
	param := r.URL.Query()
	sort := param.Get("sort")
	keyword := param.Get("keyword")

	return map[string]string{
		"sort":    sort,
		"keyword": keyword,
	}, nil
}

func (e *ScoreController) User(r *http.Request) (map[string]string, error) {
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

	fromAcc, err := e.kn.Loader().Account(addr)
	if err != nil {
		return nil, err
	}
	acc, ok := fromAcc.(*citygame.Account)
	if !ok {
		return nil, err
	}

	if acc.Height+citygame.GExpireHeight < Height {
		Height = acc.Height + citygame.GExpireHeight
	}

	gd := citygame.NewGameData(Height)
	if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, err
	}
	coinCount := gd.CoinCount

	var pos, min int
	max := len(citygame.GExpDefine)
	for {
		pos = (min + max) / 2
		if gd.TotalExp > citygame.GExpDefine[pos].AccExp {
			min = pos
		} else if gd.TotalExp < citygame.GExpDefine[pos].AccExp {
			max = pos
		} else {
			break
		}

		if gd.TotalExp >= citygame.GExpDefine[pos].AccExp && gd.TotalExp < citygame.GExpDefine[pos+1].AccExp {
			break
		}
	}

	level := int(citygame.GExpDefine[pos].Level)

	gr := gd.Resource(Height)
	data, _ := json.Marshal(gd)

	return map[string]string{
		"ID":          userid,
		"Addr":        addrStr,
		"Level":       fmt.Sprintf("%v", level),
		"Coin":        fmt.Sprintf("%v", coinCount),
		"Gold":        fmt.Sprintf("%v", gr.Balance),
		"Population":  fmt.Sprintf("%v", gr.ManProvided),
		"Electricity": fmt.Sprintf("%v", gr.PowerProvided),
		"data":        string(data),
	}, nil
}

func (e *ScoreController) Coincount(r *http.Request) (map[string]string, error) {
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
						location, _ := time.LoadLocation("Asia/Seoul")
						hTime := time.Unix(0, int64(b.Header.Timestamp())).In(location)
						str := hTime.Format("2006-01-02T15:12:13")
						var k string
						if str[:13] < str[:11]+"14" { //안지남
							hTime = hTime.Add(-time.Hour * 24)
							str2 := hTime.Format("2006-01-02T15:13:14")
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

	return map[string]string{
		"coin":   string(jm),
		"fields": string(jFields),
	}, nil
}
