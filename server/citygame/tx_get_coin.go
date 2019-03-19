package citygame

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/fletaio/extension/utxo_tx"

	"github.com/fletaio/core/amount"

	"github.com/fletaio/common"
	"github.com/fletaio/common/hash"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/transaction"
)

func init() {
	data.RegisterTransaction("fletacity.GetCoin", func(t transaction.Type) transaction.Transaction {
		return &GetCoinTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*GetCoinTx)
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
		tx := t.(*GetCoinTx)
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
			if len(gd.Coins) <= int(tx.Index) {
				return ErrInvalidCoinIndex
			}
			c := gd.Coins[tx.Index]
			if c.X == tx.X && c.Y == tx.Y && c.Height < ctx.TargetHeight() {
				h := tx.Hash()
				x := uint8(util.BytesToUint16([]byte(h[0:2]))) % GTileSize
				y := uint8(util.BytesToUint16([]byte(h[2:4]))) % GTileSize

				gd.Coins[tx.Index] = &FletaCityCoin{
					X:      x,
					Y:      y,
					Index:  tx.Index,
					Height: ctx.TargetHeight() + TimeCoinGenTime,
				}
				gd.CoinCount++

				var buffer bytes.Buffer
				if _, err := gd.WriteTo(&buffer); err != nil {
					return err
				}
				ctx.SetAccountData(tx.Address, []byte("game"), buffer.Bytes())

				ctx.Commit(sn)
				return nil
			}
			return ErrCoinNotExist
		}()

		for i := 0; i < GameCommandChannelSize; i++ {
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

// GetCoinTx is a fleta.GetCoinTx
type GetCoinTx struct {
	utxo_tx.Base
	Address common.Address
	X       uint8
	Y       uint8
	Index   uint8
}

// Hash returns the hash value of it
func (tx *GetCoinTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *GetCoinTx) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := util.WriteUint8(w, tx.Index); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *GetCoinTx) ReadFrom(r io.Reader) (int64, error) {
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
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Index = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *GetCoinTx) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(tx.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"vin":`)
	buffer.WriteString(`[`)
	for i, vin := range tx.Vin {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := json.Marshal(vin.ID()); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`,`)
	buffer.WriteString(`"address":`)
	if bs, err := json.Marshal(tx.Address); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"x":`)
	if bs, err := json.Marshal(tx.X); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"y":`)
	if bs, err := json.Marshal(tx.Y); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"index":`)
	if bs, err := json.Marshal(tx.Index); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
