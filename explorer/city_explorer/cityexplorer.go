package cityexplorer

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fletaio/common/util"

	"github.com/labstack/echo"

	"github.com/dgraph-io/badger"
	blockexplorer "github.com/fletaio/block_explorer"
	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/kernel"
)

type CityExplorer struct {
	be     *blockexplorer.BlockExplorer
	db     *badger.DB
	Kernel *kernel.Kernel
}

func NewCityExplorer(dbPath string, Kernel *kernel.Kernel, resourcePath string) (*CityExplorer, error) {
	be, err := blockexplorer.NewBlockExplorer(dbPath+"/blockExplorer", Kernel, resourcePath)
	if err != nil {
		return nil, err
	}

	opts := badger.DefaultOptions
	opts.Dir = dbPath
	opts.ValueDir = dbPath
	opts.Truncate = true
	opts.SyncWrites = true
	opts.ValueLogFileSize = 1 << 24
	lockfilePath := filepath.Join(opts.Dir, "LOCK")
	os.MkdirAll(dbPath, os.ModeDir)

	os.Remove(lockfilePath)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	{
	again:
		if err := db.RunValueLogGC(0.7); err != nil {
		} else {
			goto again
		}
	}

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
		again:
			if err := db.RunValueLogGC(0.7); err != nil {
			} else {
				goto again
			}
		}
	}()

	c := &CityExplorer{
		Kernel: Kernel,
		db:     db,
		be:     be,
	}

	ew := NewEventWatcher(c)
	Kernel.AddEventHandler(ew)

	c.be.AddAssets(Assets)
	c.be.InitURL()
	c.initURL()
	c.be.AddDataHandler(c)
	return c, nil

}

func (c *CityExplorer) initURL() {
	gc := &GameController{c.Kernel}
	sc := &ScoreController{c.Kernel, c.db}

	c.be.AddURL("/score/", "GET", func(c echo.Context) error {
		args := map[string]interface{}{}
		err := c.Render(http.StatusOK, "score/index.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/score/all", "GET", func(c echo.Context) error {
		args, err := sc.All(c.Request())
		if err != nil {
			log.Println(err)
			return err
		}

		err = c.Render(http.StatusOK, "score/all.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/score/user", "GET", func(e echo.Context) error {
		addrStr := e.Request().URL.Query().Get("addr")
		addr := common.MustParseAddress(addrStr)

		h := c.Kernel.Provider().Height()
		c.UpdateScore(nil, h, addr, "", c.Kernel.Loader())

		args, err := sc.User(e.Request())
		if err != nil {
			log.Println(err)
			return err
		}

		err = e.Render(http.StatusOK, "score/user.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/score/coincount", "GET", func(c echo.Context) error {
		args, err := sc.Coincount(c.Request())
		if err != nil {
			log.Println(err)
			return err
		}

		err = c.Render(http.StatusOK, "score/coincount.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/score/user_list", "GET", func(e echo.Context) error {
		args, err := sc.UserList(e.Request())
		if err != nil {
			log.Println(err)
			return err
		}

		err = e.Render(http.StatusOK, "score/user_list.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/score/v0.2_final", "GET", func(e echo.Context) error {
		args := map[string]interface{}{}
		err := e.Render(http.StatusOK, "score/v0.2_final.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})
	c.be.AddURL("/game/", "GET", func(c echo.Context) error {
		args, err := gc.Index(c.Request())
		if err != nil {
			log.Println(err)
			return err
		}

		err = c.Render(http.StatusOK, "game/index.html", args)
		if err != nil {
			log.Println(err)
		}
		return err
	})

	c.be.AddURL("/api/games/:address", "GET", func(e echo.Context) error {
		addrStr := e.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		var bs []byte
		var Height uint32
		var loader data.Loader
		loader = c.Kernel.Loader()
		c.Kernel.Lock()
		bs = loader.AccountData(addr, []byte("game"))
		Height = c.Kernel.Provider().Height()
		c.Kernel.Unlock()

		if len(bs) == 0 {
			return citygame.ErrNotExistAccount
		}

		gd := citygame.NewGameData(Height)

		if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
			return err
		}

		res := &WebGameRes{
			Height:       int(Height),
			PointHeight:  int(gd.PointHeight),
			PointBalance: int(gd.PointBalance),
			CoinCount:    int(gd.CoinCount),
			TotalExp:     int(gd.TotalExp),
			Coins:        gd.Coins,
			Exps:         gd.Exps,
			Tiles:        make([]*WebTile, len(gd.Tiles)),
			DefineMap:    citygame.GBuildingDefine,
			ExpDefines:   citygame.GExpDefine,
		}

		for i, tile := range gd.Tiles {
			if tile != nil {
				res.Tiles[i] = &WebTile{
					AreaType:    int(tile.AreaType),
					Level:       int(tile.Level),
					BuildHeight: int(tile.BuildHeight),
				}
			}
		}

		return e.JSON(http.StatusOK, res)
	})

}

func (c *CityExplorer) StartExplorer(port int) {
	// go func() {
	// 	for {
	// 		c.checkUserIdToAddr()
	// 		c.reflashAddrScore()
	// 		time.Sleep(10 * time.Minute)
	// 	}
	// }()
	c.be.StartExplorer(port)
}

func (c *CityExplorer) DataHandler(e echo.Context) (result interface{}, err error) {
	order := e.Param("order")

	switch order {
	case "totalScore.data":
		result = c.totalScore(e.Request())
	case "allScore.data":
		result = c.allScore(e.Request(), 20)
	case "allScoreLimit.data":
		result = c.allScore(e.Request(), 0)
	}
	if result == nil {
		err = errors.New("There is no matching url")
	}
	return
}

func (c *CityExplorer) checkUserIdToAddr() {
	if err := c.db.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("GameId")

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			value, err := item.ValueCopy(nil)
			if err != nil {
				continue
			}
			var Addr common.Address
			copy(Addr[:], value[:])

			userID := strings.Replace(string(k), "GameId", "", -1)

			loader := c.Kernel.Loader()
			if c.hasAccount(loader, Addr) == true {
				continue
			}

			height := util.BytesToUint32(Addr[:4])
		ACCOUNT:
			for height > 0 {
				// index := util.BytesToUint16(Addr[4:6])
				b, err := c.Kernel.Block(height)
				if err != nil {
					log.Println(err)
					break ACCOUNT
				}

				txs := b.Body.Transactions
				for i, tx := range txs {
					if catx, ok := tx.(*citygame.CreateAccountTx); ok {
						if catx.UserID == userID {
							coord := common.NewCoordinate(height, uint16(i))
							Addr = common.NewAddress(coord, 0)
							if c.hasAccount(loader, Addr) == true {
								if err := txn.Set([]byte("GameId"+userID), Addr[:]); err != nil {
									return err
								}
								break ACCOUNT
							}
						}
					}
				}

				log.Println(height)
				height--
			}
		}

		return nil
	}); err != nil {

	}
}

func (c *CityExplorer) hasAccount(loader data.Loader, Addr common.Address) bool {
	_, err := loader.Account(Addr)
	if err != nil {
		return false
	}
	return true
}

func (c *CityExplorer) reflashAddrScore() {
	addrs := []common.Address{}
	userIDs := []string{}
	if err := c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte("GameAddr")

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := item.Key()
			addrStr := strings.Replace(string(k), "GameAddr", "", -1)
			addr, err := common.ParseAddress(addrStr)
			if err != nil {
				log.Println(addrStr, err)
				continue
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			userID := string(value)

			addrs = append(addrs, addr)
			userIDs = append(userIDs, userID)
		}
		return nil
	}); err != nil {
		log.Println(err)
		return
	}

	if len(addrs) != len(userIDs) {
		log.Println("not match addrs and userIDs length")
		return
	}

	height := c.Kernel.Provider().Height()

	for i, addr := range addrs {
		userID := userIDs[i]
		c.UpdateScore(nil, height, addr, userID, c.Kernel.Loader())
	}

}

func (c *CityExplorer) CreatAddr(addr common.Address, tx *citygame.CreateAccountTx) {
	if err := c.db.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("GameId" + tx.UserID))
		if err == badger.ErrKeyNotFound {
			if err := txn.Set([]byte("GameAddr"+addr.String()), []byte(tx.UserID)); err != nil {
				return err
			}
			if err := txn.Set([]byte("GameId"+tx.UserID), addr[:]); err != nil {
				return err
			}
			if err := txn.Set([]byte("AddrComment"+tx.UserID), []byte(addr.String()+":"+tx.Comment)); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		log.Println(err)
	}
}

func (c *CityExplorer) UpdateScore(gd *citygame.GameData, height uint32, addr common.Address, userId string, loader data.Loader) {
	fromAcc, err := loader.Account(addr)
	if err != nil {
		return
	}
	acc, ok := fromAcc.(*citygame.Account)
	if !ok {
		return
	}

	if acc.Height+citygame.GExpireHeight < height {
		height = acc.Height + citygame.GExpireHeight
	}

	if gd == nil {
		gd = citygame.NewGameData(height)
		bs := loader.AccountData(addr, []byte("game"))
		if len(bs) == 0 {
			log.Println("addr : ", addr.String())
			return
		}
		if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
			return
		}
	}
	if err := c.db.Update(func(txn *badger.Txn) error {
		if userId == "" {
			item, err := txn.Get([]byte("GameAddr" + addr.String()))
			if err != nil {
				return err
			} else {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}
				userId = string(value)
			}
		}

		addrStr := addr.String()
		r := gd.Resource(height)

		level := r.Balance + uint64(r.ManProvided*4) + uint64(r.PowerProvided*6)

		sc := ScoreCase{
			UserID:        userId,
			Level:         level,
			Balance:       uint64(r.Balance),
			ManProvided:   uint64(r.ManProvided),
			PowerProvided: uint64(r.PowerProvided),
			CoinCount:     uint64(gd.CoinCount),
		}

		updateSortedKey(txn, Level, addrStr, sc)
		updateSortedKey(txn, Balance, addrStr, sc)
		updateSortedKey(txn, ManProvided, addrStr, sc)
		updateSortedKey(txn, PowerProvided, addrStr, sc)
		updateSortedKey(txn, CoinCount, addrStr, sc)

		return nil
	}); err != nil {
		return
	}

}

func updateSortedKey(txn *badger.Txn, sType ScoreType, addrStr string, sc ScoreCase) error {
	gameScoreAddr := []byte(fmt.Sprintf("%v:Addr:%v", getType(sType), addrStr))
	v := []byte(fmt.Sprintf("%v:Score:%020v%v", getType(sType), sc.getValue(sType), addrStr))

	item, err := txn.Get(gameScoreAddr)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return err
		}
	} else {
		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = txn.Delete(value)
		if err != nil {
			return err
		}
	}

	buf := &bytes.Buffer{}
	sc.WriteTo(buf)

	txn.Set(v, buf.Bytes())
	txn.Set(gameScoreAddr, v)

	return nil
}

type score struct {
	Rank     uint32
	Addr     string
	Reward   string
	Score    string
	AllScore *ScoreCase
}

type allScore struct {
	Total       []score
	Gold        []score
	Population  []score
	Electricity []score
	CoinCount   []score
}

func (c *CityExplorer) allScore(r *http.Request, limit uint32) (result []score) {
	param := r.URL.Query()
	sort := param.Get("sort")
	keyword := param.Get("keyword")

	c.db.View(func(txn *badger.Txn) error {
		st := getTypeFromString("Game" + sort)

		var prefix []byte
		if keyword != "" {
			item, err := txn.Get([]byte("GameId" + keyword))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return err
				}
			} else {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}

				addr := common.Address{}
				copy(addr[:], value)

				gameScoreAddr := []byte(fmt.Sprintf("%v:Addr:%v", getType(st), addr.String()))
				item, err := txn.Get(gameScoreAddr)
				if err != nil {
					if err != badger.ErrKeyNotFound {
						return err
					}
				} else {
					var err error
					prefix, err = item.ValueCopy(nil)
					if err != nil {
						return err
					}
				}
			}
		}

		result, _ = c.getScore(txn, prefix, st, limit)
		return nil
	})

	return
}

func (c *CityExplorer) totalScore(r *http.Request) (result *allScore) {
	result = &allScore{
		Total:       []score{},
		Gold:        []score{},
		Population:  []score{},
		Electricity: []score{},
		CoinCount:   []score{},
	}
	c.db.View(func(txn *badger.Txn) error {
		result.Total, _ = c.getScore(txn, nil, Level, 10)
		result.CoinCount, _ = c.getScore(txn, nil, CoinCount, 10)
		result.Gold, _ = c.getScore(txn, nil, Balance, 5)
		result.Population, _ = c.getScore(txn, nil, ManProvided, 5)
		result.Electricity, _ = c.getScore(txn, nil, PowerProvided, 5)
		return nil
	})

	return
}

func (c *CityExplorer) getScore(txn *badger.Txn, prefixKey []byte, sType ScoreType, limit uint32) ([]score, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	opts.Reverse = true
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(getType(sType) + ":Score:")
	var rank uint32
	if len(prefixKey) == 0 {
		prefixKey = append([]byte(getType(sType)+":Score:"), 0xFF)
	} else {
		opts2 := badger.DefaultIteratorOptions
		opts2.PrefetchValues = false
		opts2.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		startkey := append([]byte(getType(sType)+":Score:"), 0xFF)
		for it.Seek(startkey); !it.ValidForPrefix(prefixKey); it.Next() {
			rank++
		}
	}
	s := []score{}
	for it.Seek(prefixKey); it.ValidForPrefix(prefix); it.Next() {
		rank++
		item := it.Item()
		k := item.Key()
		err := item.Value(func(v []byte) error {
			buf := bytes.NewBuffer(v)
			sc := &ScoreCase{}
			sc.ReadFrom(buf)

			value := strings.TrimPrefix(string(k), getType(sType)+":Score:")
			num := value[:20]
			Addr := value[20:]

			var Reward string
			addr, err := common.ParseAddress(Addr)
			if err == nil {
				height := util.BytesToUint32(addr[:4])
				i := util.BytesToUint16(addr[4:6])
				b, err := c.Kernel.Block(height)
				if err == nil {
					if len(b.Body.Transactions) >= int(i) {
						tx := b.Body.Transactions[i]
						if catx, ok := tx.(*citygame.CreateAccountTx); ok {
							Reward = catx.Reward
						} else {
							panic("is not create account")
						}
					}
				} else {
					log.Println(Addr, height)
				}
			} else {
				log.Println(Addr)
			}

			s = append(s, score{
				Rank:     rank,
				Addr:     Addr,
				Reward:   Reward,
				Score:    num,
				AllScore: sc,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
		if limit != 0 && rank >= limit {
			break
		}
	}

	return s, nil

}

type WebGameRes struct {
	Height       int                                              `json:"height"`
	PointHeight  int                                              `json:"point_height"`
	PointBalance int                                              `json:"point_balance"`
	CoinCount    int                                              `json:"coin_count"`
	TotalExp     int                                              `json:"total_exp"`
	Coins        []*citygame.FletaCityCoin                        `json:"coins"`
	Exps         []*citygame.FletaCityExp                         `json:"exps"`
	Tiles        []*WebTile                                       `json:"tiles"`
	DefineMap    map[citygame.AreaType][]*citygame.BuildingDefine `json:"define_map"`
	ExpDefines   []*citygame.ExpDefine                            `json:"exp_defines"`
}
