package cityexplorer

import (
	"bytes"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/core/block"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/kernel"
	"github.com/fletaio/core/message_def"
	"github.com/fletaio/core/transaction"
)

// EventWatcher TODO
type EventWatcher struct {
	ce *CityExplorer
}

// NewEventWatcher returns a EventWatcher
func NewEventWatcher(ce *CityExplorer) *EventWatcher {
	ew := &EventWatcher{
		ce: ce,
	}
	return ew
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
		var addr common.Address
		var userID string
		switch tx := t.(type) {
		case *citygame.CreateAccountTx:
			coord := &common.Coordinate{Height: b.Header.Height(), Index: uint16(i)}
			addr = common.NewAddress(coord, 0)
			userID = tx.UserID
			ew.ce.CreatAddr(addr, tx)
		case *citygame.DemolitionTx:
			addr = tx.Address
		case *citygame.ConstructionTx:
			addr = tx.Address
		case *citygame.UpgradeTx:
			addr = tx.Address
		case *citygame.GetCoinTx:
			addr = tx.Address
		default:
			continue
		}

		gd, err := getWebTileNotify(ctx, addr, b.Header.Height(), uint16(i))
		if err != nil {
			if err != citygame.ErrNotExistGameData {
				continue
			}
		}
		ew.ce.UpdateScore(gd, b.Header.Height(), addr, userID, kn.Loader())
	}
}

func (ew *EventWatcher) AfterPushTransaction(kn *kernel.Kernel, tx transaction.Transaction, sigs []common.Signature) {
}

// DoTransactionBroadcast called when a transaction need to be broadcast
func (ew *EventWatcher) DoTransactionBroadcast(kn *kernel.Kernel, msg *message_def.TransactionMessage) {
}

// DebugLog TEMP
func (ew *EventWatcher) DebugLog(kn *kernel.Kernel, args ...interface{}) {}

func getWebTileNotify(ctx *data.Context, addr common.Address, height uint32, index uint16) (*citygame.GameData, error) {
	gd := citygame.NewGameData(height + 1)
	bs := ctx.AccountData(addr, []byte("game"))
	if len(bs) == 0 {
		return nil, citygame.ErrNotExistGameData
	}
	_, err := gd.ReadFrom(bytes.NewReader(bs))
	return gd, err
}
