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
	data.RegisterTransaction("fletacity.GetExp", func(t transaction.Type) transaction.Transaction {
		return &GetExpTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*GetExpTx)
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
		tx := t.(*GetExpTx)
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
			for i, e := range gd.Exps {
				if e.X == tx.X && e.Y == tx.Y && e.AreaType == tx.AreaType && e.Level == tx.Level {
					bds, has := GBuildingDefine[e.AreaType]
					if !has {
						return ErrInvalidAreaType
					}
					if e.Level == 0 || int(e.Level) >= len(bds)+1 {
						return ErrInvalidLevel
					}
					gd.TotalExp += uint64(bds[e.Level-1].Exp)

					ed := GExpDefine[0]
					for _, v := range GExpDefine {
						if v.AccExp <= gd.TotalExp {
							ed = v
						} else {
							break
						}
					}
					h := hash.Hash(util.Uint32ToBytes(ctx.TargetHeight()))
					for i := len(gd.Coins); i < int(ed.CoinGen); i++ {
						x := uint8(util.BytesToUint16([]byte(h[i*2:i*2+2]))) % GTileSize
						y := uint8(util.BytesToUint16([]byte(h[i*2+2:i*2+4]))) % GTileSize
						gd.Coins = append(gd.Coins, &FletaCityCoin{
							X:      x,
							Y:      y,
							Index:  uint8(i),
							Height: ctx.TargetHeight() + TimeCoinGenTime*2,
						})
					}

					if len(gd.Exps) == 1 {
						gd.Exps = []*FletaCityExp{}
					} else {
						if i == 0 {
							gd.Exps = gd.Exps[1:]
						} else if i == len(gd.Exps)-1 {
							gd.Exps = gd.Exps[:len(gd.Exps)-1]
						} else {
							gd.Exps = append(gd.Exps[:i], gd.Exps[i+1:]...)
						}
					}

					var buffer bytes.Buffer
					if _, err := gd.WriteTo(&buffer); err != nil {
						return err
					}
					ctx.SetAccountData(tx.Address, []byte("game"), buffer.Bytes())

					e, err := ctx.Eventer().NewByTypeName("fletacity.GetExp")
					if err != nil {
						return err
					}
					ev := e.(*GetExpEvent)
					ev.Base.Coord_ = coord
					ev.Address = tx.Address
					ev.X = tx.X
					ev.Y = tx.Y
					ev.TotalExp = gd.TotalExp

					ctx.EmitEvent(e)

					ctx.Commit(sn)
					return nil
				}
			}
			return ErrExpNotExist
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

// GetExpTx is a fleta.GetExpTx
type GetExpTx struct {
	utxo_tx.Base
	Address  common.Address
	X        uint8
	Y        uint8
	AreaType AreaType
	Level    uint8
}

// Hash returns the hash value of it
func (tx *GetExpTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *GetExpTx) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := util.WriteUint8(w, uint8(tx.AreaType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, tx.Level); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *GetExpTx) ReadFrom(r io.Reader) (int64, error) {
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
		tx.AreaType = AreaType(v)
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Level = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *GetExpTx) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"area_type":`)
	if bs, err := json.Marshal(tx.AreaType); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"level":`)
	if bs, err := json.Marshal(tx.Level); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
