package main

import (
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/key"
	"github.com/fletaio/core/node"
	"github.com/fletaio/core/transaction"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

type cityGameCommand struct {
	Key                    key.Key
	GameKernel             *kernel.Kernel
	Node                   *node.Node
	ew                     *EventWatcher
	CreateAccountChannelID uint64
	UsingChannelCount      int64
}

func (cg *cityGameCommand) index(c echo.Context) error {
	urlPath := c.Request().URL.Path[1:]
	data, err := t.Route(c.Request(), urlPath)
	if err != nil {
		return err
	}
	return c.HTML(http.StatusOK, string(data))
}
func (cg *cityGameCommand) websocketAddress(c echo.Context) error {
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
	a, err := cg.GameKernel.Loader().Account(addr)
	if err != nil {
		return err
	}
	acc := a.(*citygame.Account)
	if !acc.KeyHash.Equal(pubhash) {
		return citygame.ErrInvalidAddress
	}

	cg.ew.AddWriter(addr, conn)
	defer cg.ew.RemoveWriter(addr, conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return nil
		}
	}
}

func (cg *cityGameCommand) chainHeight(c echo.Context) error {
	res := &WebHeightRes{
		Height: int(cg.GameKernel.Provider().Height()),
	}
	return c.JSON(http.StatusOK, res)
}

func (cg *cityGameCommand) accountsGet(c echo.Context) error {
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

	loader := cg.GameKernel.Loader()
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
}

func (cg *cityGameCommand) accountsPost(c echo.Context) error {
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
	loader := cg.GameKernel.Loader()
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

	defer atomic.AddInt64(&cg.UsingChannelCount, -1)
	if atomic.AddInt64(&cg.UsingChannelCount, 1) >= citygame.CreateAccountChannelSize {
		return citygame.ErrQueueFull
	}

	cnid := atomic.AddUint64(&cg.CreateAccountChannelID, 1) % citygame.CreateAccountChannelSize

	utxoid := util.BytesToUint64(loader.AccountData(rootAddress, []byte("utxo"+strconv.FormatUint(cnid, 10))))

	tx := t.(*citygame.CreateAccountTx)
	tx.Timestamp_ = uint64(time.Now().UnixNano())
	tx.Vin = []*transaction.TxIn{transaction.NewTxIn(utxoid)}
	tx.KeyHash = pubhash
	tx.UserID = req.UserID
	tx.Reward = req.Reward

	TxHash = tx.Hash()

	if sig, err := cg.Key.Sign(TxHash); err != nil {
		return err
	} else if err := cg.Node.CommitTransaction(tx, []common.Signature{sig}); err != nil {
		return err
	}

	timer := time.NewTimer(10 * time.Second)

	cp := cg.GameKernel.Provider()
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
}

func (cg *cityGameCommand) reportsAddress(c echo.Context) error {
	addrStr := c.Param("address")
	addr, err := common.ParseAddress(addrStr)
	if err != nil {
		return err
	}

	cg.GameKernel.Lock()
	bs := cg.GameKernel.Loader().AccountData(addr, []byte("game"))
	Height := cg.GameKernel.Provider().Height()
	cg.GameKernel.Unlock()

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
}

func (cg *cityGameCommand) gamesAddress(c echo.Context) error {
	addrStr := c.Param("address")
	addr, err := common.ParseAddress(addrStr)
	if err != nil {
		return err
	}

	cg.GameKernel.Lock()
	bs := cg.GameKernel.Loader().AccountData(addr, []byte("game"))
	Height := cg.GameKernel.Provider().Height()
	cg.GameKernel.Unlock()

	if len(bs) == 0 {
		return citygame.ErrNotExistAccount
	}

	gd := citygame.NewGameData(Height)

	txs := []*UTXO{}
	for i := 0; i < citygame.GameAccountChannelSize; i++ {
		utxoID := util.BytesToUint64(cg.GameKernel.Loader().AccountData(addr, []byte("utxo"+strconv.Itoa(i))))
		txIn := transaction.NewTxIn(utxoID)
		vin := []*transaction.TxIn{txIn}
		for len(vin) > 0 {
			txIn := vin[0]
			utxoID := transaction.MarshalID(txIn.Height, txIn.Index, txIn.N)
			b, err := cg.GameKernel.Block(txIn.Height)
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
	clbs := cg.GameKernel.Loader().AccountData(addr, []byte("CoinList"))
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

	ccbs := cg.GameKernel.Loader().AccountData(addr, []byte("GetCoinCount"))
	if len(ccbs) == 4 {
		coinCount := util.BytesToUint32(ccbs)
		res.CoinCount = int(coinCount)
	} else {
		res.CoinCount = 0
	}

	return c.JSON(http.StatusOK, res)
}

func (cg *cityGameCommand) gamesAddressDemolition(c echo.Context) error {
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

	loader := cg.GameKernel.Loader()

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
}

func (cg *cityGameCommand) gamesAddressUpgrade(c echo.Context) error {
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

	loader := cg.GameKernel.Loader()

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
}

func (cg *cityGameCommand) gamesAddressGetcoin(c echo.Context) error {
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

	loader := cg.GameKernel.Loader()

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
}

func (cg *cityGameCommand) gamesAddressCommit(c echo.Context) error {
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

	loader := cg.GameKernel.Loader()

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

	if err := cg.Node.CommitTransaction(tx, []common.Signature{sig}); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
