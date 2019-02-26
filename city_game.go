package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	blockexplorer "git.fleta.io/fleta/city_game/block_explorer"

	citygame "git.fleta.io/fleta/city_game/city_game_context"

	"git.fleta.io/fleta/city_game/template"
	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/block"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/formulator"
	"git.fleta.io/fleta/core/kernel"
	"git.fleta.io/fleta/core/key"
	"git.fleta.io/fleta/core/message_def"
	"git.fleta.io/fleta/core/observer"
	"git.fleta.io/fleta/core/transaction"
	"git.fleta.io/fleta/framework/peer"
	"git.fleta.io/fleta/framework/router"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

var (
	libPath string
)

func init() {
	var pwd string
	{
		pc := make([]uintptr, 10) // at least 1 entry needed
		runtime.Callers(1, pc)
		f := runtime.FuncForPC(pc[0])
		pwd, _ = f.FileLine(pc[0])

		path := strings.Split(pwd, "/")
		pwd = strings.Join(path[:len(path)-1], "/")
	}

	libPath = pwd + "/"
}

var t *template.Template

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// {
	// 	addr := common.MustParseAddress("3URbvUjYHd")

	// 	buf := bytes.NewBuffer(addr[:6])
	// 	c := &common.Coordinate{}
	// 	c.ReadFrom(buf)
	// 	log.Println(c.Height, c.Index)

	// 	buf = bytes.NewBuffer(addr[6:12])
	// 	c.ReadFrom(buf)
	// 	log.Println(c.Height, c.Index)

	// 	return
	// }
	t = template.NewTemplate(&template.TemplateConfig{
		TemplatePath: libPath + "html/pages/",
		LayoutPath:   libPath + "html/layout/",
	})

	var Key key.Key
	if bs, err := hex.DecodeString("fb4d410401e2cb9eb4d9ae497b9f7c585eb0bfb88f6e0b4adfe54e9451d809ea"); err != nil {
		panic(err)
	} else {
		k, err := key.NewMemoryKeyFromBytes(bs)
		if err != nil {
			panic(err)
		}
		Key = k
	}

	obstrs := []string{
		"cca49818f6c49cf57b6c420cdcd98fcae08850f56d2ff5b8d287fddc7f9ede08",
		"39f1a02bed5eff3f6247bb25564cdaef20d410d77ef7fc2c0181b1d5b31ce877",
		"2b97bc8f21215b7ed085cbbaa2ea020ded95463deef6cbf31bb1eadf826d4694",
		"3b43d728deaa62d7c8790636bdabbe7148a6641e291fd1f94b157673c0172425",
		"e6cf2724019000a3f703db92829ecbd646501c0fd6a5e97ad6774d4ad621f949",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	ObserverKeys := make([]common.PublicHash, 0, len(obstrs))

	NetAddressMap := map[common.PublicHash]string{}
	NetAddressMapForFr := map[common.PublicHash]string{}
	for i, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			Num := strconv.Itoa(i + 1)
			pubhash := common.NewPublicHash(Key.PublicKey())
			NetAddressMap[pubhash] = "127.0.0.1:300" + Num
			NetAddressMapForFr[pubhash] = "127.0.0.1:500" + Num
			ObserverKeys = append(ObserverKeys, pubhash)
		}
	}
	ObserverKeyMap := map[common.PublicHash]bool{}
	for _, pubhash := range ObserverKeys {
		ObserverKeyMap[pubhash] = true
	}

	frstrs := []string{
		"13db949719b42eac09a8d7eeb7d9d259d595657f810c50aeb249250483652f98",
	}

	frkeys := make([]key.Key, 0, len(frstrs))
	for _, v := range frstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			frkeys = append(frkeys, Key)
		}
	}

	ObserverHeights := []uint32{}

	obs := []*observer.Observer{}
	for _, obkey := range obkeys {
		GenCoord := common.NewCoordinate(0, 0)
		act := data.NewAccounter(GenCoord)
		tran := data.NewTransactor(GenCoord)
		if err := initChainComponent(act, tran); err != nil {
			panic(err)
		}

		StoreRoot := "./observer/" + common.NewPublicHash(obkey.PublicKey()).String()

		//os.RemoveAll(StoreRoot)

		ks, err := kernel.NewStore(StoreRoot+"/kernel", 1, act, tran, true)
		if err != nil {
			panic(err)
		}

		GenesisContextData := data.NewContextData(data.NewEmptyLoader(ks.ChainCoord(), ks.Accounter(), ks.Transactor()), nil)
		if err := initGenesisContextData(ks, GenesisContextData); err != nil {
			panic(err)
		}

		rd := &mockRewarder{}
		kn, err := kernel.NewKernel(&kernel.Config{
			ChainCoord:     GenCoord,
			ObserverKeyMap: ObserverKeyMap,
		}, ks, rd, GenesisContextData)
		if err != nil {
			panic(err)
		}

		cfg := &observer.Config{
			ChainCoord:     GenCoord,
			Key:            obkey,
			ObserverKeyMap: NetAddressMap,
		}
		ob, err := observer.NewObserver(cfg, kn)
		if err != nil {
			panic(err)
		}
		obs = append(obs, ob)

		ObserverHeights = append(ObserverHeights, kn.Provider().Height())
	}

	Formulators := []string{}
	FormulatorHeights := []uint32{}

	var GameKernel *kernel.Kernel
	frs := []*formulator.Formulator{}
	for _, frkey := range frkeys {
		GenCoord := common.NewCoordinate(0, 0)
		act := data.NewAccounter(GenCoord)
		tran := data.NewTransactor(GenCoord)
		if err := initChainComponent(act, tran); err != nil {
			panic(err)
		}

		StoreRoot := "./formulator/" + common.NewPublicHash(frkey.PublicKey()).String()

		//os.RemoveAll(StoreRoot)

		ks, err := kernel.NewStore(StoreRoot+"/kernel", 1, act, tran, true)
		if err != nil {
			panic(err)
		}

		GenesisContextData := data.NewContextData(data.NewEmptyLoader(ks.ChainCoord(), ks.Accounter(), ks.Transactor()), nil)
		if err := initGenesisContextData(ks, GenesisContextData); err != nil {
			panic(err)
		}

		rd := &mockRewarder{}
		kn, err := kernel.NewKernel(&kernel.Config{
			ChainCoord:     GenCoord,
			ObserverKeyMap: ObserverKeyMap,
		}, ks, rd, GenesisContextData)
		if err != nil {
			panic(err)
		}

		cfg := &formulator.Config{
			ChainCoord:     GenCoord,
			Key:            frkey,
			ObserverKeyMap: NetAddressMapForFr,
			Formulator:     common.MustParseAddress("3CUsUpvEK"),
			Router: router.Config{
				Network: "tcp",
				Port:    7000,
			},
			Peer: peer.Config{
				StorePath: StoreRoot + "/peers",
			},
		}
		fr, err := formulator.NewFormulator(cfg, kn)
		if err != nil {
			panic(err)
		}
		frs = append(frs, fr)

		Formulators = append(Formulators, cfg.Formulator.String())
		FormulatorHeights = append(FormulatorHeights, kn.Provider().Height())

		GameKernel = kn
	}

	for i, ob := range obs {
		go func(BindOb string, BindFr string, ob *observer.Observer) {
			ob.Run(BindOb, BindFr)
		}(":300"+strconv.Itoa(i+1), ":500"+strconv.Itoa(i+1), ob)
	}

	for _, fr := range frs {
		go func(fr *formulator.Formulator) {
			fr.Run()
		}(fr)
	}

	basePath := "./test/"
	be, err := blockexplorer.NewBlockExplorer(basePath, GameKernel)
	if err != nil {
		panic(err)
	}

	ew := NewEventWatcher(be)
	GameKernel.AddEventHandler(ew)

	e := echo.New()
	e.Static("/js", libPath+"html/resource/js")
	e.Static("/css", libPath+"html/resource/css")
	e.Static("/images", libPath+"html/resource/images")
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err)
		c.HTML(code, err.Error())
	}

	e.GET("/", func(c echo.Context) error {
		urlPath := c.Request().URL.Path[1:]
		data, err := t.Route(c.Request(), urlPath)
		if err != nil {
			return err
		}
		return c.HTML(http.StatusOK, string(data))
	})
	e.GET("/websocket/:address", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
		if err != nil {
			return err
		}
		defer conn.Close()

		bs := make([]byte, 32)
		crand.Read(bs)
		h := hash.Hash(bs)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(h.String())); err != nil {
			return err
		}
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		sig, err := common.ParseSignature(string(msg))
		if err != nil {
			return err
		}

		var pubhash common.PublicHash
		if pubkey, err := common.RecoverPubkey(h, sig); err != nil {
			return err
		} else {
			pubhash = common.NewPublicHash(pubkey)
		}

		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}
		a, err := GameKernel.Loader().Account(addr)
		if err != nil {
			return err
		}
		acc := a.(*citygame.Account)
		if !acc.KeyHash.Equal(pubhash) {
			return citygame.ErrInvalidAddress
		}

		ew.AddWriter(addr, conn)
		defer ew.RemoveWriter(addr, conn)

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
		}
	})

	gAPI := e.Group("/api")
	gAPI.GET("/chain/height", func(c echo.Context) error {
		res := &WebHeightRes{
			Height: int(GameKernel.Provider().Height()),
		}
		return c.JSON(http.StatusOK, res)
	})
	gAPI.GET("/accounts", func(c echo.Context) error {
		pubkeyStr := c.QueryParam("pubkey")
		if len(pubkeyStr) == 0 {
			return citygame.ErrInvalidPublicKey
		}

		pubkey, err := common.ParsePublicKey(pubkeyStr)
		if err != nil {
			return err
		}

		pubhash := common.NewPublicHash(pubkey)
		KeyHashID := append(citygame.PrefixKeyHash, pubhash[:]...)

		loader := GameKernel.Loader()
		var rootAddress common.Address
		if bs := loader.AccountData(rootAddress, KeyHashID); len(bs) > 0 {
			var addr common.Address
			copy(addr[:], bs)
			utxos := []uint64{}
			for i := 0; i < citygame.GameAccountChannelSize; i++ {
				utxos = append(utxos, util.BytesToUint64(loader.AccountData(addr, []byte("utxo"+strconv.Itoa(i)))))
			}
			res := &WebAccountRes{
				Address: addr.String(),
				UTXOs:   utxos,
			}
			return c.JSON(http.StatusOK, res)
		} else {
			return citygame.ErrNotExistAccount
		}
	})
	var CreateAccountChannelID uint64
	var UsingChannelCount int64
	gAPI.POST("/accounts", func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		var req WebAccountReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		if len(req.UserID) < 4 {
			return citygame.ErrShortUserID
		}
		if len(req.Reward) < 1 {
			return citygame.ErrInvalidReward
		}

		pubkey, err := common.ParsePublicKey(req.PublicKey)
		if err != nil {
			return err
		}

		pubhash := common.NewPublicHash(pubkey)
		KeyHashID := append(citygame.PrefixKeyHash, pubhash[:]...)
		UserIDHashID := append(citygame.PrefixUserID, []byte(req.UserID)...)
		RewardHashID := append(citygame.PrefixReward, []byte(req.Reward)...)

		var TxHash hash.Hash256
		loader := GameKernel.Loader()
		var rootAddress common.Address
		if bs := loader.AccountData(rootAddress, KeyHashID); len(bs) > 0 {
			var addr common.Address
			copy(addr[:], bs)
			utxos := []uint64{}
			for i := 0; i < citygame.GameAccountChannelSize; i++ {
				utxos = append(utxos, util.BytesToUint64(loader.AccountData(addr, []byte("utxo"+strconv.Itoa(i)))))
			}
			res := &WebAccountRes{
				Address: addr.String(),
				UTXOs:   utxos,
			}
			return c.JSON(http.StatusOK, res)
		}
		if bs := loader.AccountData(rootAddress, UserIDHashID); len(bs) > 0 {
			return citygame.ErrExistUserID
		}
		if bs := loader.AccountData(rootAddress, RewardHashID); len(bs) > 0 {
			return citygame.ErrExistReward
		}

		t, err := loader.Transactor().NewByType(CreateAccountTransctionType)
		if err != nil {
			return err
		}

		defer atomic.AddInt64(&UsingChannelCount, -1)
		if atomic.AddInt64(&UsingChannelCount, 1) >= citygame.CreateAccountChannelSize {
			return citygame.ErrQueueFull
		}

		cnid := atomic.AddUint64(&CreateAccountChannelID, 1) % citygame.CreateAccountChannelSize

		utxoid := util.BytesToUint64(loader.AccountData(rootAddress, []byte("utxo"+strconv.FormatUint(cnid, 10))))

		tx := t.(*citygame.CreateAccountTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxoid)}
		tx.KeyHash = pubhash
		tx.UserID = req.UserID
		tx.Reward = req.Reward

		TxHash = tx.Hash()

		if sig, err := Key.Sign(TxHash); err != nil {
			return err
		} else if err := GameKernel.AddTransaction(tx, []common.Signature{sig}); err != nil {
			return err
		}

		timer := time.NewTimer(10 * time.Second)

		cp := GameKernel.Provider()
		SentHeight := cp.Height()
		for {
			select {
			case <-timer.C:
				return c.NoContent(http.StatusRequestTimeout)
			default:
				height := cp.Height()
				if SentHeight < height {
					SentHeight = height

					var rootAddress common.Address
					if bs := loader.AccountData(rootAddress, KeyHashID); len(bs) > 0 {
						var addr common.Address
						copy(addr[:], bs)
						utxos := []uint64{}
						for i := 0; i < citygame.GameAccountChannelSize; i++ {
							utxos = append(utxos, util.BytesToUint64(loader.AccountData(addr, []byte("utxo"+strconv.Itoa(i)))))
						}
						res := &WebAccountRes{
							Address: addr.String(),
							UTXOs:   utxos,
						}
						return c.JSON(http.StatusOK, res)
					}
				}
				time.Sleep(250 * time.Millisecond)
			}
		}
	})
	gAPI.GET("/reports/:address", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		GameKernel.Lock()
		bs := GameKernel.Loader().AccountData(addr, []byte("game"))
		Height := GameKernel.Provider().Height()
		GameKernel.Unlock()

		if len(bs) == 0 {
			return citygame.ErrNotExistAccount
		}

		gd := citygame.NewGameData(Height)
		if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
			return err
		}

		gr := gd.Resource(Height)
		res := &WebReportRes{
			Height:        int(Height),
			PointHeight:   int(gd.PointHeight),
			Balance:       int(gr.Balance),
			PowerRemained: int(gr.PowerRemained),
			PowerProvided: int(gr.PowerProvided),
			ManRemained:   int(gr.ManRemained),
			ManProvided:   int(gr.ManProvided),
		}
		return c.JSON(http.StatusOK, res)
	})
	gAPI.GET("/games/:address", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		GameKernel.Lock()
		bs := GameKernel.Loader().AccountData(addr, []byte("game"))
		Height := GameKernel.Provider().Height()
		GameKernel.Unlock()

		if len(bs) == 0 {
			return citygame.ErrNotExistAccount
		}

		gd := citygame.NewGameData(Height)

		txs := []*UTXO{}
		for i := 0; i < citygame.GameAccountChannelSize; i++ {
			utxoID := util.BytesToUint64(GameKernel.Loader().AccountData(addr, []byte("utxo"+strconv.Itoa(i))))
			txIn := transaction.NewTxIn(utxoID)
			vin := []*transaction.TxIn{txIn}
			for len(vin) > 0 {
				txIn := vin[0]
				utxoID := transaction.MarshalID(txIn.Height, txIn.Index, txIn.N)
				b, err := GameKernel.Block(txIn.Height)
				if err == nil {
					t := b.Body.Transactions[txIn.Index]
					var x int
					var y int
					switch tx := t.(type) {
					case *citygame.DemolitionTx:
						vin = tx.Vin
						x = int(tx.X)
						y = int(tx.Y)
					case *citygame.UpgradeTx:
						vin = tx.Vin
						x = int(tx.X)
						y = int(tx.Y)
					case *citygame.ConstructionTx:
						vin = tx.Vin
						x = int(tx.X)
						y = int(tx.Y)
					case *citygame.GetCoinTx:
						vin = tx.Vin
						x = int(tx.X)
						y = int(tx.Y)
					default:
						vin = nil
					}

					index := sort.Search(len(txs), func(i int) bool { return txs[i].ID > utxoID })
					txs = append(txs, nil)
					copy(txs[index+1:], txs[index:])
					txs[index] = &UTXO{
						ID:     utxoID,
						X:      x,
						Y:      y,
						Type:   int(t.Type()),
						Height: txIn.Height,
						Hash:   t.Hash().String(),
					}

				}
			}
		}

		if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
			return err
		}

		res := &WebGameRes{
			Height:       int(Height),
			PointHeight:  int(gd.PointHeight),
			PointBalance: int(gd.PointBalance),
			Tiles:        make([]*WebTile, len(gd.Tiles)),
			Txs:          txs,
			DefineMap:    citygame.GBuildingDefine,
		}
		clbs := GameKernel.Loader().AccountData(addr, []byte("CoinList"))
		bf := bytes.NewBuffer(clbs)
		if cl, err := citygame.CLReadFrom(bf); err == nil {
			res.CoinList = cl
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

		ccbs := GameKernel.Loader().AccountData(addr, []byte("GetCoinCount"))
		if len(ccbs) == 4 {
			coinCount := util.BytesToUint32(ccbs)
			res.CoinCount = int(coinCount)
		} else {
			res.CoinCount = 0
		}

		return c.JSON(http.StatusOK, res)
	})
	gAPI.POST("/games/:address/commands/demolition", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		var req WebDemolitionReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		if req.UTXO == 0 {
			return citygame.ErrInvalidUTXO
		}
		if req.X > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.Y > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}

		loader := GameKernel.Loader()

		t, err := loader.Transactor().NewByType(DemolitionTransactionType)
		if err != nil {
			return err
		}

		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		tx := t.(*citygame.DemolitionTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(req.UTXO)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    int(DemolitionTransactionType),
			TxHex:   hex.EncodeToString(buffer.Bytes()),
			HashHex: tx.Hash().String(),
		})
	})
	gAPI.POST("/games/:address/commands/upgrade", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		var req WebUpgradeReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		if req.UTXO == 0 {
			return citygame.ErrInvalidUTXO
		}
		if req.X > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.Y > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.AreaType <= int(citygame.EmptyAreaType) || req.AreaType >= int(citygame.EndOfAreaType) {
			return citygame.ErrInvalidAreaType
		}
		if req.TargetLevel >= len(citygame.GBuildingDefine[citygame.AreaType(req.AreaType)])+1 {
			return citygame.ErrInvalidLevel
		}

		loader := GameKernel.Loader()

		var t transaction.Transaction
		if req.TargetLevel == 1 {
			t, err = loader.Transactor().NewByType(ConstructionTransactionType)
		} else {
			t, err = loader.Transactor().NewByType(UpgradeTransactionType)
		}
		if err != nil {
			return err
		}

		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		var tx *citygame.UpgradeTx
		if req.TargetLevel == 1 {
			ctx := t.(*citygame.ConstructionTx)
			tx = ctx.UpgradeTx
		} else {
			tx = t.(*citygame.UpgradeTx)
		}

		tx.Timestamp_ = uint64(time.Now().UnixNano())
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(req.UTXO)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.AreaType = citygame.AreaType(req.AreaType)
		tx.TargetLevel = uint8(req.TargetLevel)

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		var txType int
		if req.TargetLevel == 1 {
			txType = int(ConstructionTransactionType)
		} else {
			txType = int(UpgradeTransactionType)
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    txType,
			TxHex:   hex.EncodeToString(buffer.Bytes()),
			HashHex: tx.Hash().String(),
		})
	})
	gAPI.POST("/games/:address/commands/getcoin", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		var req WebGetCoinReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		if req.UTXO == 0 {
			return citygame.ErrInvalidUTXO
		}
		if req.X > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.Y > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		coinType := uint8(req.CoinType)
		if coinType != uint8(citygame.ConstructCoinType) && coinType != uint8(citygame.TimeCoinType) {
			return citygame.ErrInvalidCoinType
		}

		loader := GameKernel.Loader()

		t, err := loader.Transactor().NewByType(GetCoinTransactionType)
		if err != nil {
			return err
		}

		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		tx := t.(*citygame.GetCoinTx)

		tx.Timestamp_ = uint64(time.Now().UnixNano())
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(req.UTXO)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.CoinType = citygame.CoinType(req.CoinType)
		tx.TargetHash = hash.MustParseHex(req.Hash)
		tx.TargetHeight = req.Height

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    int(GetCoinTransactionType),
			TxHex:   hex.EncodeToString(buffer.Bytes()),
			HashHex: tx.Hash().String(),
		})
	})
	gAPI.POST("/games/:address/commands/commit", func(c echo.Context) error {
		body, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}
		defer c.Request().Body.Close()

		var req WebCommitReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}

		txBytes, err := hex.DecodeString(req.TxHex)
		if err != nil {
			return err
		}

		sigBytes, err := hex.DecodeString(req.SigHex)
		if err != nil {
			return err
		}

		loader := GameKernel.Loader()

		tx, err := loader.Transactor().NewByType(transaction.Type(req.Type))
		if err != nil {
			return err
		}
		if _, err := tx.ReadFrom(bytes.NewReader(txBytes)); err != nil {
			return err
		}

		var sig common.Signature
		if _, err := sig.ReadFrom(bytes.NewReader(sigBytes)); err != nil {
			return err
		}

		if err := GameKernel.AddTransaction(tx, []common.Signature{sig}); err != nil {
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	go be.StartExplorer(9088)

	e.Start(":9080")
}

// EventWatcher TODO
type EventWatcher struct {
	sync.Mutex
	writerMap map[common.Address]*websocket.Conn
	be        *blockexplorer.BlockExplorer
}

// NewEventWatcher returns a EventWatcher
func NewEventWatcher(be *blockexplorer.BlockExplorer) *EventWatcher {
	ew := &EventWatcher{
		writerMap: map[common.Address]*websocket.Conn{},
		be:        be,
	}
	return ew
}

// AddWriter TODO
func (ew *EventWatcher) AddWriter(addr common.Address, w *websocket.Conn) {
	ew.Lock()
	defer ew.Unlock()

	if old, has := ew.writerMap[addr]; has {
		old.WriteJSON(&WebTileNotify{
			Type: 99,
		})
		old.Close()
	}
	ew.writerMap[addr] = w
}

// RemoveWriter TODO
func (ew *EventWatcher) RemoveWriter(addr common.Address, w *websocket.Conn) {
	ew.Lock()
	defer ew.Unlock()

	if old, has := ew.writerMap[addr]; has {
		old.Close()
	}
	delete(ew.writerMap, addr)
}

// OnCreateContext called when a context creation (error prevent using context)
func (ew *EventWatcher) OnCreateContext(kn *kernel.Kernel, ctx *data.Context) error {
	return nil
}

// OnProcessBlock called when processing a block to the chain (error prevent processing block)
func (ew *EventWatcher) OnProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) error {
	return nil
}

// OnPushTransaction called when pushing a transaction to the transaction pool (error prevent push transaction)
func (ew *EventWatcher) OnPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) error {
	return nil
}

// AfterProcessBlock called when processed block to the chain
func (ew *EventWatcher) AfterProcessBlock(kn *kernel.Kernel, b *block.Block, s *block.ObserverSigned, ctx *data.Context) {
	for i, t := range b.Body.Transactions {
		switch tx := t.(type) {
		case *citygame.DemolitionTx:
			wtn, gd, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), i)
			if err != nil {
				continue
			}
			wtn.Type = 0
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)
			wtn.AreaType = int(0)
			wtn.Level = int(0)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())

			ew.Notify(tx.Address, wtn)
			ew.be.UpdateScore(gd, b.Header.Height(), tx.Address, "")
		case *citygame.UpgradeTx, *citygame.ConstructionTx:
			var utx *citygame.UpgradeTx
			switch _tx := tx.(type) {
			case *citygame.UpgradeTx:
				utx = _tx
			case *citygame.ConstructionTx:
				utx = _tx.UpgradeTx
			}
			wtn, gd, err := getWebTileNotify(ctx, utx.Address, b.Header.Height(), i)
			if err != nil {
				continue
			}

			wtn.Type = 1
			wtn.X = int(utx.X)
			wtn.Y = int(utx.Y)
			wtn.AreaType = int(utx.AreaType)
			wtn.Level = int(utx.TargetLevel)

			wtn.Tx.X = int(utx.X)
			wtn.Tx.Y = int(utx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())

			clbs := ctx.AccountData(utx.Address, []byte("CoinList"))
			bf := bytes.NewBuffer(clbs)
			if cl, err := citygame.CLReadFrom(bf); err == nil {
				wtn.CoinList = cl
			}
			ew.Notify(utx.Address, wtn)

			ew.be.UpdateScore(gd, b.Header.Height(), utx.Address, "")
		case *citygame.GetCoinTx:
			wtn, _, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), i)
			if err != nil {
				continue
			}
			wtn.Type = 2
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())

			clbs := ctx.AccountData(tx.Address, []byte("CoinList"))
			bf := bytes.NewBuffer(clbs)
			if cl, err := citygame.CLReadFrom(bf); err == nil {
				wtn.CoinList = cl
				ew.Notify(tx.Address, wtn)
			}
		}
	}
}

func (ew *EventWatcher) AfterPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) {
}

// DoTransactionBroadcast called when a transaction need to be broadcast
func (ew *EventWatcher) DoTransactionBroadcast(kn *kernel.Kernel, msg *message_def.TransactionMessage) {
}

// DebugLog TEMP
func (ew *EventWatcher) DebugLog(kn *kernel.Kernel, args ...interface{}) {}

func getWebTileNotify(ctx *data.Context, addr common.Address, height uint32, index int) (*WebTileNotify, *citygame.GameData, error) {
	gd := citygame.NewGameData(height)
	bs := ctx.AccountData(addr, []byte("game"))
	if len(bs) == 0 {
		return nil, nil, citygame.ErrNotExistGameData
	}
	if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, nil, err
	}
	id := transaction.MarshalID(height, uint16(index), 0)
	var errorMsg string
	for i := 0; i < citygame.GameAccountChannelSize; i++ {
		newid := util.BytesToUint64(ctx.AccountData(addr, []byte("utxo"+strconv.Itoa(index))))
		if id == newid {
			bs := ctx.AccountData(addr, []byte("result"+strconv.Itoa(index)))
			if len(bs) > 0 {
				errorMsg = string(bs)
			}
			break
		}
	}

	ccbs := ctx.AccountData(addr, []byte("GetCoinCount"))
	coinCount := util.BytesToUint32(ccbs)

	return &WebTileNotify{
		Height:       int(height),
		PointHeight:  int(gd.PointHeight),
		PointBalance: int(gd.PointBalance),
		CoinCount:    int(coinCount),
		UTXO:         int(id),
		Tx: &UTXO{
			ID: id,
		},
		Error: errorMsg,
	}, gd, nil
}

// Notify TODO
func (ew *EventWatcher) Notify(addr common.Address, noti *WebTileNotify) {
	ew.Lock()
	defer ew.Unlock()

	if conn, has := ew.writerMap[addr]; has {
		conn.WriteJSON(noti)
	}
}

type WebAccountReq struct {
	PublicKey string `json:"public_key"`
	UserID    string `json:"user_id"`
	Reward    string `json:"reward"`
}

type WebAccountRes struct {
	Address string   `json:"address"`
	UTXOs   []uint64 `json:"utxos"`
}

type WebReportRes struct {
	Height        int `json:"height"`
	PointHeight   int `json:"point_height"`
	Balance       int `json:"balance"`
	PowerRemained int `json:"power_remained"`
	PowerProvided int `json:"power_provided"`
	ManRemained   int `json:"man_remained"`
	ManProvided   int `json:"man_provided"`
}

type WebGameRes struct {
	Height       int                                              `json:"height"`
	PointHeight  int                                              `json:"point_height"`
	PointBalance int                                              `json:"point_balance"`
	CoinCount    int                                              `json:"coin_count"`
	Tiles        []*WebTile                                       `json:"tiles"`
	Txs          []*UTXO                                          `json:"txs"`
	CoinList     map[string]*citygame.FletaCityCoin               `json:"fleta_city_coins"`
	DefineMap    map[citygame.AreaType][]*citygame.BuildingDefine `json:"define_map"`
}

type WebHeightRes struct {
	Height int `json:"height"`
}

type WebDemolitionReq struct {
	UTXO uint64 `json:"utxo"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type WebUpgradeReq struct {
	UTXO        uint64 `json:"utxo"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	AreaType    int    `json:"area_type"`
	TargetLevel int    `json:"target_level"`
}

type WebGetCoinReq struct {
	UTXO     uint64 `json:"utxo"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Hash     string `json:"hash"`
	Height   uint32 `json:"height"`
	CoinType int    `json:"coin_type"`
}

type WebTxRes struct {
	Type    int    `json:"type"`
	TxHex   string `json:"tx_hex"`
	HashHex string `json:"hash_hex"`
}

type WebCommitReq struct {
	Type   int    `json:"type"`
	TxHex  string `json:"tx_hex"`
	SigHex string `json:"sig_hex"`
}

type WebTileNotify struct {
	Type         int                                `json:"type"`
	X            int                                `json:"x"`
	Y            int                                `json:"y"`
	AreaType     int                                `json:"area_type"`
	Level        int                                `json:"level"`
	Height       int                                `json:"height"`
	PointHeight  int                                `json:"point_height"`
	PointBalance int                                `json:"point_balance"`
	CoinCount    int                                `json:"coin_count"`
	UTXO         int                                `json:"utxo"`
	Tx           *UTXO                              `json:"tx"`
	CoinList     map[string]*citygame.FletaCityCoin `json:"fleta_city_coins"`
	Error        string                             `json:"error"`
}

type WebTile struct {
	AreaType    int `json:"area_type"`
	Level       int `json:"level"`
	BuildHeight int `json:"build_height"`
}

type UTXO struct {
	ID     uint64 `json:"id"`
	Type   int    `json:"tx_type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Height uint32 `json:"height"`
	Hash   string `json:"hash"`
}
