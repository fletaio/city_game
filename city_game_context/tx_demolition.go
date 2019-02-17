package citygame

import (
	"bytes"
	"io"
	"strconv"

	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/extension/utxo_tx"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fletacity.Demolition", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &DemolitionTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*DemolitionTx)
		if len(tx.Vin) != 1 {
			return ErrInvalidTxInCount
		}
		if len(signers) != 1 {
			return ErrInvalidSignerCount
		}

		if utxo, err := loader.UTXO(tx.Vin[0].ID()); err != nil {
			return err
		} else {
			if !utxo.PublicHash.Equal(signers[0]) {
				return ErrInvalidTransactionSignature
			}
		}

		fromAcc, err := loader.Account(tx.Address)
		if err != nil {
			return err
		}

		if err := loader.Accounter().Validate(loader, fromAcc, signers); err != nil {
			return err
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*DemolitionTx)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		utxo, err := ctx.UTXO(tx.Vin[0].ID())
		if err != nil {
			return nil, err
		}

		if err := ctx.DeleteUTXO(utxo.ID()); err != nil {
			return nil, err
		}

		gameErr := func() error {
			sn := ctx.Snapshot()
			defer ctx.Revert(sn)

			gd := NewGameData(ctx.TargetHeight())
			bs := ctx.AccountData(tx.Address, []byte("game"))
			if _, err := gd.ReadFrom(bytes.NewReader(bs)); err != nil {
				return err
			}

			idx := tx.X + GTileSize*tx.Y
			tile := gd.Tiles[idx]
			if tile == nil {
				return ErrNotExistTile
			}

			res := gd.Resource(ctx.TargetHeight())
			bd := GBuildingDefine[tile.AreaType][tile.Level-1]
			switch tile.AreaType {
			case IndustrialAreaType:
				if res.PowerRemained < bd.Output {
					return ErrInvalidDemolition
				}
			case ResidentialAreaType:
				if res.ManRemained < bd.Output {
					return ErrInvalidDemolition
				}
			}
			gd.Tiles[idx] = nil

			gd.UpdatePoint(ctx.TargetHeight(), gd.Resource(ctx.TargetHeight()).Balance)

			var buffer bytes.Buffer
			if _, err := gd.WriteTo(&buffer); err != nil {
				return err
			}
			ctx.SetAccountData(tx.Address, []byte("game"), buffer.Bytes())

			ctx.Commit(sn)
			return nil
		}()

		for i := 0; i < GameAccountChannelSize; i++ {
			did := []byte("utxo" + strconv.Itoa(i))
			oldid := util.BytesToUint64(ctx.AccountData(tx.Address, did))
			if oldid == tx.Vin[0].ID() {
				id := transaction.MarshalID(coord.Height, coord.Index, 0)
				ctx.CreateUTXO(id, &transaction.TxOut{
					Amount:     amount.NewCoinAmount(0, 0),
					PublicHash: utxo.PublicHash,
				})
				ctx.SetAccountData(tx.Address, did, util.Uint64ToBytes(id))
				if gameErr != nil {
					ctx.SetAccountData(tx.Address, []byte("result"+strconv.Itoa(i)), []byte(gameErr.Error()))
				} else {
					ctx.SetAccountData(tx.Address, []byte("result"+strconv.Itoa(i)), nil)
				}
				break
			}
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// DemolitionTx is a fleta.DemolitionTx
// It is engraved dapp on main chain
type DemolitionTx struct {
	utxo_tx.Base
	Address common.Address
	X       uint8
	Y       uint8
}

// Hash returns the hash value of it
func (tx *DemolitionTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *DemolitionTx) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.Address.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, tx.X); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, tx.Y); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *DemolitionTx) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.Address.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.X = v
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Y = v
	}
	return read, nil
}
