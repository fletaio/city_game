package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/cmd/closer"
	"github.com/fletaio/framework/config"
	"github.com/fletaio/framework/router/evilnode"
	"github.com/fletaio/framework/rpc"

	"github.com/fletaio/citygame/server/blockexplorer"
	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/block"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/message_def"
	"github.com/fletaio/core/node"
	"github.com/fletaio/core/reward"
	"github.com/fletaio/core/transaction"
	"github.com/fletaio/framework/peer"
	"github.com/fletaio/framework/router"
	"github.com/fletaio/framework/template"

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

// Config is a configuration for the cmd
type Config struct {
	SeedNodes    []string
	ObserverKeys []string
	Port         int
	APIPort      int
	ServerPort   int
	ExplorerPort int
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
	if err := initChainComponent(act, tran); err != nil {
		panic(err)
	}

	GenesisContextData, err := initGenesisContextData(act, tran)
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
	if s, err := kernel.NewStore(cfg.StoreRoot+"/kernel", BlockchainVersion, act, tran, cfg.ForceRecover); err != nil {
		if cfg.ForceRecover || err != badger.ErrTruncateNeeded {
			panic(err)
		} else {
			fmt.Println(err)
			fmt.Println("Do you want to recover database(it can be failed)? [y/n]")
			var answer string
			fmt.Scanf("%s", &answer)
			if strings.ToLower(answer) == "y" {
				if s, err := kernel.NewStore(cfg.StoreRoot+"/kernel", BlockchainVersion, act, tran, true); err != nil {
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

	t = template.NewTemplate(&template.TemplateConfig{
		TemplatePath: libPath + "html/pages/",
		LayoutPath:   libPath + "html/layout/",
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

	basePath := "./test/"
	be, err := blockexplorer.NewBlockExplorer(basePath, GameKernel)
	if err != nil {
		panic(err)
	}

	ew := NewEventWatcher(be)
	GameKernel.AddEventHandler(ew)

	cg := &cityGameCommand{
		GameKernel: GameKernel,
		Key:        GameKey,
		ew:         ew,
	}

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

	e.GET("/", cg.index)
	e.GET("/websocket/:address", cg.websocketAddress)

	gAPI := e.Group("/api")
	gAPI.GET("/chain/height", cg.chainHeight)
	gAPI.GET("/accounts", cg.accountsGet)
	gAPI.POST("/accounts", cg.accountsPost)
	gAPI.GET("/reports/:address", cg.reportsAddress)
	gAPI.GET("/games/:address", cg.gamesAddress)
	gAPI.POST("/games/:address/commands/demolition", cg.gamesAddressDemolition)
	gAPI.POST("/games/:address/commands/upgrade", cg.gamesAddressUpgrade)
	gAPI.POST("/games/:address/commands/getcoin", cg.gamesAddressGetcoin)
	gAPI.POST("/games/:address/commands/commit", cg.gamesAddressCommit)

	go be.StartExplorer(cfg.ExplorerPort)
	go e.Start(":" + strconv.Itoa(cfg.ServerPort))

	go func() {
		if err := rm.Run(kn, ":"+strconv.Itoa(cfg.APIPort)); err != nil {
			if http.ErrServerClosed != err {
				panic(err)
			}
		}
	}()

	cm.Wait()
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
			clbs := ctx.AccountData(tx.Address, []byte("CoinList"))
			bf := bytes.NewBuffer(clbs)
			if cl, err := citygame.CLReadFrom(bf); err == nil {
				wtn.CoinList = cl
			}

			ew.Notify(tx.Address, wtn)
			ew.be.UpdateScore(gd, b.Header.Height(), tx.Address, "", wtn.CoinCount)
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

			ew.be.UpdateScore(gd, b.Header.Height(), utx.Address, "", wtn.CoinCount)
		case *citygame.GetCoinTx:
			wtn, gd, err := getWebTileNotify(ctx, tx.Address, b.Header.Height(), i)
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
			ew.be.UpdateScore(gd, b.Header.Height(), tx.Address, "", wtn.CoinCount)
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
