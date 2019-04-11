package cityexplorer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fletaio/common/util"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/core/kernel"
)

type ScoreController struct {
	kn *kernel.Kernel
	db *badger.DB
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

	isExpired := ""
	if acc.Height+citygame.GExpireHeight < Height {
		Height = acc.Height + citygame.GExpireHeight
		isExpired = "Expired"
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

	b, err := e.kn.Block(acc.Height)
	if err != nil {
		return nil, err
	}
	b.Header.Timestamp()

	return map[string]string{
		"ID":          userid,
		"Addr":        addrStr,
		"Level":       fmt.Sprintf("%v", level),
		"Coin":        fmt.Sprintf("%v", coinCount),
		"Gold":        fmt.Sprintf("%v", gr.Balance),
		"Population":  fmt.Sprintf("%v", gr.ManProvided),
		"Electricity": fmt.Sprintf("%v", gr.PowerProvided),
		"data":        string(data),
		"CreateTime":  fmt.Sprintf("%v", b.Header.Timestamp()),
		"Expired":     isExpired,
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

func (e *ScoreController) UserList(r *http.Request) (map[string]string, error) {
	data := []map[string]string{}
	err := e.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("AddrComment")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			userID := strings.Replace(string(k), "AddrComment", "", -1)
			m := map[string]string{}
			m["userID"] = userID

			value, err := item.ValueCopy(nil)
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			idComment := string(value)
			strs := strings.Split(idComment, ":")
			var comment string
			if len(strs) > 1 {
				comment = strings.Join(strs[1:], ":")
			}
			m["comment"] = comment

			userIDItem, err := txn.Get([]byte("GameId" + userID))
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			value, err = userIDItem.ValueCopy(nil)
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			var Addr common.Address
			copy(Addr[:], value)
			m["addr"] = Addr.String()

			loader := e.kn.Loader()
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			fromAcc, err := loader.Account(Addr)
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			acc, ok := fromAcc.(*citygame.Account)
			if !ok {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			b, err := e.kn.Block(acc.Height)
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}

			location, _ := time.LoadLocation("Asia/Seoul")
			hTime := time.Unix(0, int64(b.Header.Timestamp())).In(location)

			m["time"] = hTime.Format("2006-01-02T15:12:13")

			h := util.BytesToUint32(Addr[:4])
			i := util.BytesToUint16(Addr[4:6])
			b, err = e.kn.Block(h)
			if err != nil {
				m["error"] = err.Error()
				data = append(data, m)
				continue
			}
			txs := b.Body.Transactions
			if uint16(len(txs)) > i {
				tx := txs[i]
				if caTx, ok := tx.(*citygame.CreateAccountTx); ok {
					m["eth"] = caTx.Reward
				}
			}

			data = append(data, m)
		}
		return nil
	})

	bs, err := json.Marshal(data)
	return map[string]string{
		"data": string(bs),
	}, err
}
