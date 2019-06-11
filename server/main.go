package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/block"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/message_def"
	"github.com/fletaio/core/node"
	"github.com/fletaio/core/reward"
	"github.com/fletaio/core/transaction"
	"github.com/fletaio/framework/closer"
	"github.com/fletaio/framework/config"
	"github.com/fletaio/framework/peer"
	"github.com/fletaio/framework/router"
	"github.com/fletaio/framework/router/evilnode"
	"github.com/fletaio/framework/rpc"
)

// consts
const (
	CreateAccountChannelSize = 100
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Config is a configuration for the cmd
type Config struct {
	SeedNodes    []string
	ObserverKeys []string
	Port         int
	APIPort      int
	ServerPort   int
	KeyHex       string
	StoreRoot    string
	ForceRecover bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./data"
	}

	ObserverKeyMap := map[common.PublicHash]bool{}
	for _, k := range cfg.ObserverKeys {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		ObserverKeyMap[pubhash] = true
	}

	GenCoord := common.NewCoordinate(0, 0)
	act := data.NewAccounter(GenCoord)
	tran := data.NewTransactor(GenCoord)
	evt := data.NewEventer(GenCoord)
	if err := initChainComponent(act, tran, evt); err != nil {
		panic(err)
	}

	GenesisContextData, err := initGenesisContextData(act, tran, evt)
	if err != nil {
		panic(err)
	}

	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	var ks *kernel.Store
	if s, err := kernel.NewStore(cfg.StoreRoot+"/kernel", BlockchainVersion, act, tran, evt, cfg.ForceRecover); err != nil {
		if cfg.ForceRecover || err != badger.ErrTruncateNeeded {
			panic(err)
		} else {
			fmt.Println(err)
			fmt.Println("Do you want to recover database(it can be failed)? [y/n]")
			var answer string
			fmt.Scanf("%s", &answer)
			if strings.ToLower(answer) == "y" {
				if s, err := kernel.NewStore(cfg.StoreRoot+"/kernel", BlockchainVersion, act, tran, evt, true); err != nil {
					panic(err)
				} else {
					ks = s
				}
			} else {
				os.Exit(1)
			}
		}
	} else {
		ks = s
	}
	cm.Add("kernel.Store", ks)

	rd := &reward.TestNetRewarder{}
	kn, err := kernel.NewKernel(&kernel.Config{
		ChainCoord:              GenCoord,
		ObserverKeyMap:          ObserverKeyMap,
		MaxBlocksPerFormulator:  8,
		MaxTransactionsPerBlock: 5000,
	}, ks, rd, GenesisContextData)
	if err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("kernel.Kernel", kn)

	ndcfg := &node.Config{
		ChainCoord: GenCoord,
		SeedNodes:  cfg.SeedNodes,
		Router: router.Config{
			Network: "tcp",
			Port:    cfg.Port,
			EvilNodeConfig: evilnode.Config{
				StorePath: cfg.StoreRoot + "/router",
			},
		},
		Peer: peer.Config{
			StorePath: cfg.StoreRoot + "/peers",
		},
	}
	nd, err := node.NewNode(ndcfg, kn)
	if err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("cmd.Node", nd)

	go nd.Run()
	rm := rpc.NewManager()
	cm.RemoveAll()
	cm.Add("rpc.Manager", rm)
	cm.Add("cmd.Node", nd)
	kn.AddEventHandler(rm)

	defer func() {
		cm.CloseAll()
		if err := recover(); err != nil {
			kn.DebugLog("Panic", err)
			panic(err)
		}
	}()

	// Chain
	rm.Add("Version", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		return kn.Provider().Version(), nil
	})
	rm.Add("Height", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		return kn.Provider().Height(), nil
	})
	rm.Add("LastHash", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		return kn.Provider().LastHash(), nil
	})
	rm.Add("Hash", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		if arg.Len() < 1 {
			return nil, rpc.ErrInvalidArgument
		}
		height, err := arg.Uint32(0)
		if err != nil {
			return nil, err
		}
		h, err := kn.Provider().Hash(height)
		if err != nil {
			return nil, err
		}
		return h, nil
	})
	rm.Add("Header", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		if arg.Len() < 1 {
			return nil, rpc.ErrInvalidArgument
		}
		height, err := arg.Uint32(0)
		if err != nil {
			return nil, err
		}
		h, err := kn.Provider().Header(height)
		if err != nil {
			return nil, err
		}
		return h, nil
	})
	rm.Add("Block", func(kn *kernel.Kernel, ID interface{}, arg *rpc.Argument) (interface{}, error) {
		if arg.Len() < 1 {
			return nil, rpc.ErrInvalidArgument
		}
		height, err := arg.Uint32(0)
		if err != nil {
			return nil, err
		}
		cd, err := kn.Provider().Data(height)
		if err != nil {
			return nil, err
		}
		b := &block.Block{
			Header: cd.Header.(*block.Header),
			Body:   cd.Body.(*block.Body),
		}
		return b, nil
	})

	GameKernel := kn

	var GameKey key.Key
	if bs, err := hex.DecodeString(cfg.KeyHex); err != nil {
		panic(err)
	} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
		panic(err)
	} else {
		GameKey = Key
	}

	var CreateAccountChannelID uint64
	var UsingChannelCount int64

	ew := NewEventWatcher()
	GameKernel.AddEventHandler(ew)

	e := echo.New()
	web := NewWebServer(e, "./webfiles")
	e.Renderer = web
	web.SetupStatic(e, "/public", "./webfiles/public")
	webChecker := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			web.CheckWatch()
			return next(c)
		}
	}
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		c.Logger().Error(err)
		c.HTML(code, err.Error())
	}

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

	e.GET("/", func(c echo.Context) error {
		args := make(map[string]interface{})
		return c.Render(http.StatusOK, "index.html", args)
	}, webChecker)

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
			utxos := []string{}
			for i := 0; i < citygame.GameCommandChannelSize; i++ {
				utxo := util.BytesToUint64(loader.AccountData(addr, []byte("utxo"+strconv.Itoa(i))))
				utxos = append(utxos, fmt.Sprintf("%v", utxo))
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

		var TxHash hash.Hash256
		loader := GameKernel.Loader()
		var rootAddress common.Address
		if bs := loader.AccountData(rootAddress, KeyHashID); len(bs) > 0 {
			return citygame.ErrExistKeyHash
		}
		if bs := loader.AccountData(rootAddress, UserIDHashID); len(bs) > 0 {
			return citygame.ErrExistUserID
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
		tx.Comment = req.Comment

		TxHash = tx.Hash()

		if sig, err := GameKey.Sign(TxHash); err != nil {
			return err
		} else if err := nd.CommitTransaction(tx, []common.Signature{sig}); err != nil { //TEMP
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
						utxos := []string{}
						for i := 0; i < citygame.GameCommandChannelSize; i++ {
							bs := util.BytesToUint64(loader.AccountData(addr, []byte("utxo"+strconv.Itoa(i))))
							utxos = append(utxos, fmt.Sprintf("%v", bs))
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
	gAPI.GET("/games/:address", func(c echo.Context) error {
		addrStr := c.Param("address")
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return err
		}

		var bs []byte
		var Height uint32
		var loader data.Loader
		if e := ew.LastBlockEvent(); e != nil {
			bs = e.ctx.AccountData(addr, []byte("game"))
			Height = e.b.Header.Height()
			loader = e.ctx
		} else {
			loader = GameKernel.Loader()
			GameKernel.Lock()
			bs = loader.AccountData(addr, []byte("game"))
			Height = GameKernel.Provider().Height()
			GameKernel.Unlock()
		}

		if len(bs) == 0 {
			return citygame.ErrNotExistAccount
		}

		gd := citygame.NewGameData(Height + 1)
		// res := gd.Resource(ctx.TargetHeight())

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
			Txs:          []*UTXO{},
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
		if req.UTXO == "" {
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
		utxo, err := strconv.ParseUint(req.UTXO, 10, 64)
		if err != nil {
			return citygame.ErrInvalidUTXO
		}
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxo)}
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
	gAPI.POST("/games/:address/commands/construction", func(c echo.Context) error {
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

		var req WebConstructionReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		if req.UTXO == "" {
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

		loader := GameKernel.Loader()
		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		t, err := loader.Transactor().NewByType(ConstructionTransactionType)
		if err != nil {
			return err
		}
		tx := t.(*citygame.ConstructionTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		utxo, err := strconv.ParseUint(req.UTXO, 10, 64)
		if err != nil {
			return citygame.ErrInvalidUTXO
		}
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxo)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.AreaType = citygame.AreaType(req.AreaType)

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    int(ConstructionTransactionType),
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
		if req.UTXO == "" {
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
		if req.TargetLevel < 2 || req.TargetLevel > len(citygame.GBuildingDefine[citygame.AreaType(req.AreaType)]) {
			return citygame.ErrInvalidLevel
		}

		loader := GameKernel.Loader()
		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		t, err := loader.Transactor().NewByType(UpgradeTransactionType)
		if err != nil {
			return err
		}
		tx := t.(*citygame.UpgradeTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		utxo, err := strconv.ParseUint(req.UTXO, 10, 64)
		if err != nil {
			return citygame.ErrInvalidUTXO
		}
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxo)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.AreaType = citygame.AreaType(req.AreaType)
		tx.TargetLevel = uint8(req.TargetLevel)

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    int(UpgradeTransactionType),
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
		if req.UTXO == "" {
			return citygame.ErrInvalidUTXO
		}
		if req.X > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.Y > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}

		loader := GameKernel.Loader()
		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		t, err := loader.Transactor().NewByType(GetCoinTransactionType)
		if err != nil {
			return err
		}
		tx := t.(*citygame.GetCoinTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		utxo, err := strconv.ParseUint(req.UTXO, 10, 64)
		if err != nil {
			return citygame.ErrInvalidUTXO
		}
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxo)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.Index = uint8(req.Index)

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
	gAPI.POST("/games/:address/commands/getexp", func(c echo.Context) error {
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

		var req WebGetExpReq
		if err := json.Unmarshal(body, &req); err != nil {
			return err
		}
		if req.UTXO == "" {
			return citygame.ErrInvalidUTXO
		}
		if req.X > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}
		if req.Y > citygame.GTileSize {
			return citygame.ErrInvalidPosition
		}

		loader := GameKernel.Loader()
		if is, err := loader.IsExistAccount(addr); err != nil {
			return err
		} else if !is {
			return citygame.ErrNotExistAccount
		}

		t, err := loader.Transactor().NewByType(GetExpTransactionType)
		if err != nil {
			return err
		}
		tx := t.(*citygame.GetExpTx)
		tx.Timestamp_ = uint64(time.Now().UnixNano())
		utxo, err := strconv.ParseUint(req.UTXO, 10, 64)
		if err != nil {
			return citygame.ErrInvalidUTXO
		}
		tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxo)}
		tx.Address = addr
		tx.X = uint8(req.X)
		tx.Y = uint8(req.Y)
		tx.AreaType = citygame.AreaType(req.AreaType)
		tx.Level = uint8(req.Level)

		var buffer bytes.Buffer
		if _, err := tx.WriteTo(&buffer); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, &WebTxRes{
			Type:    int(GetExpTransactionType),
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

		if err := nd.CommitTransaction(tx, []common.Signature{sig}); err != nil { //TEMP
			return err
		}
		return c.NoContent(http.StatusOK)
	})

	go ew.Run()
	e.Start(":8080")
}

type BlockEvent struct {
	b   *block.Block
	ctx *data.Context
}

// EventWatcher TODO
type EventWatcher struct {
	sync.Mutex
	writerMap      map[common.Address]*websocket.Conn
	blockEventChan chan *BlockEvent
	lastBlockEvent *BlockEvent
}

// NewEventWatcher returns a EventWatcher
func NewEventWatcher() *EventWatcher {
	ew := &EventWatcher{
		writerMap:      map[common.Address]*websocket.Conn{},
		blockEventChan: make(chan *BlockEvent, 10000),
	}
	return ew
}

// LastBlockEvent returns the last block event
func (ew *EventWatcher) LastBlockEvent() *BlockEvent {
	ew.Lock()
	defer ew.Unlock()
	return ew.lastBlockEvent
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

// Run TODO
func (ew *EventWatcher) Run() {
	for {
		select {
		case e := <-ew.blockEventChan:
			ew.processBlock(e.b, e.ctx)
		}
	}
}

func (ew *EventWatcher) processBlock(b *block.Block, ctx *data.Context) {
	for i, t := range b.Body.Transactions {
		switch tx := t.(type) {
		case *citygame.DemolitionTx:
			wtn, _, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), uint16(i))
			if err != nil {
				continue
			}
			wtn.Type = int(DemolitionTransactionType)
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
		case *citygame.ConstructionTx:
			wtn, gd, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), uint16(i))
			if err != nil {
				continue
			}

			wtn.Type = int(ConstructionTransactionType)
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)
			wtn.AreaType = int(tx.AreaType)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())
			if len(wtn.Error) == 0 {
				for _, e := range gd.Exps {
					if e.X == tx.X && e.Y == tx.Y && e.AreaType == tx.AreaType && e.Level == 1 {
						wtn.Exp = e
						break
					}
				}
			}
			ew.Notify(tx.Address, wtn)
		case *citygame.UpgradeTx:
			wtn, gd, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), uint16(i))
			if err != nil {
				continue
			}

			wtn.Type = int(UpgradeTransactionType)
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)
			wtn.AreaType = int(tx.AreaType)
			wtn.Level = int(tx.TargetLevel)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())
			if len(wtn.Error) == 0 {
				for _, e := range gd.Exps {
					if e.X == tx.X && e.Y == tx.Y && e.AreaType == tx.AreaType && e.Level == tx.TargetLevel {
						wtn.Exp = e
						break
					}
				}
			}
			ew.Notify(tx.Address, wtn)
		case *citygame.GetCoinTx:
			wtn, gd, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), uint16(i))
			if err != nil {
				continue
			}
			wtn.Type = int(GetCoinTransactionType)
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())
			wtn.Coin = gd.Coins[tx.Index]
			ew.Notify(tx.Address, wtn)
		case *citygame.GetExpTx:
			wtn, _, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), uint16(i))
			if err != nil {
				continue
			}
			wtn.Type = int(GetExpTransactionType)
			wtn.X = int(tx.X)
			wtn.Y = int(tx.Y)

			wtn.Tx.X = int(tx.X)
			wtn.Tx.Y = int(tx.Y)
			wtn.Tx.Hash = t.Hash().String()
			wtn.Tx.Height = b.Header.Height()
			wtn.Tx.Type = int(t.Type())
			wtn.Exp = &citygame.FletaCityExp{
				X:        tx.X,
				Y:        tx.Y,
				AreaType: tx.AreaType,
				Level:    tx.Level,
			}
			ew.Notify(tx.Address, wtn)
		}
	}
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
	e := &BlockEvent{
		b:   b,
		ctx: ctx,
	}

	ew.Lock()
	ew.lastBlockEvent = e
	ew.Unlock()

	ew.blockEventChan <- e
}

func (ew *EventWatcher) AfterPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) {
}

// DoTransactionBroadcast called when a transaction need to be broadcast
func (ew *EventWatcher) DoTransactionBroadcast(kn *kernel.Kernel, msg *message_def.TransactionMessage) {
}

// DebugLog TEMP
func (ew *EventWatcher) DebugLog(kn *kernel.Kernel, args ...interface{}) {}

func getWebTileNotify(ctx *data.Context, addr common.Address, height uint32, index uint16) (*WebTileNotify, *citygame.GameData, error) {
	gd := citygame.NewGameData(height + 1)
	bs := ctx.AccountData(addr, []byte("game"))
	if len(bs) == 0 {
		return nil, nil, citygame.ErrNotExistGameData
	}
	if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
		return nil, nil, err
	}
	id := transaction.MarshalID(height, index, 0)
	var errorMsg string
	for i := 0; i < citygame.GameCommandChannelSize; i++ {
		bs := ctx.AccountData(addr, []byte("utxo"+strconv.Itoa(i)))
		if len(bs) < 8 {
			continue
		}
		newid := util.BytesToUint64(bs)
		if id == newid {
			bs := ctx.AccountData(addr, []byte("result"+strconv.Itoa(i)))
			if len(bs) > 0 {
				errorMsg = string(bs)
			}
			break
		}
	}

	return &WebTileNotify{
		Height:       int(height),
		PointHeight:  int(gd.PointHeight),
		PointBalance: int(gd.PointBalance),
		CoinCount:    int(gd.CoinCount),
		TotalExp:     int(gd.TotalExp),
		UTXO:         fmt.Sprintf("%v", id),
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
	Comment   string `json:"comment"`
}

type WebAccountRes struct {
	Address string   `json:"address"`
	UTXOs   []string `json:"utxos"`
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
	TotalExp     int                                              `json:"total_exp"`
	Coins        []*citygame.FletaCityCoin                        `json:"coins"`
	Exps         []*citygame.FletaCityExp                         `json:"exps"`
	Tiles        []*WebTile                                       `json:"tiles"`
	Txs          []*UTXO                                          `json:"txs"`
	DefineMap    map[citygame.AreaType][]*citygame.BuildingDefine `json:"define_map"`
	ExpDefines   []*citygame.ExpDefine                            `json:"exp_defines"`
}

type WebHeightRes struct {
	Height int `json:"height"`
}

type WebDemolitionReq struct {
	UTXO string `json:"utxo"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type WebConstructionReq struct {
	UTXO     string `json:"utxo"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	AreaType int    `json:"area_type"`
}

type WebUpgradeReq struct {
	UTXO        string `json:"utxo"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	AreaType    int    `json:"area_type"`
	TargetLevel int    `json:"target_level"`
}

type WebGetCoinReq struct {
	UTXO  string `json:"utxo"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Index int    `json:"index"`
}

type WebGetExpReq struct {
	UTXO     string `json:"utxo"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	AreaType int    `json:"area_type"`
	Level    int    `json:"level"`
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
	Type         int                     `json:"type"`
	X            int                     `json:"x"`
	Y            int                     `json:"y"`
	AreaType     int                     `json:"area_type"`
	Level        int                     `json:"level"`
	Height       int                     `json:"height"`
	PointHeight  int                     `json:"point_height"`
	PointBalance int                     `json:"point_balance"`
	CoinCount    int                     `json:"coin_count"`
	TotalExp     int                     `json:"total_exp"`
	UTXO         string                  `json:"utxo"`
	Tx           *UTXO                   `json:"tx"`
	Coin         *citygame.FletaCityCoin `json:"coin"`
	Exp          *citygame.FletaCityExp  `json:"exp"`
	Error        string                  `json:"error"`
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
