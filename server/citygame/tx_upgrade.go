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
	data.RegisterTransaction("fletacity.Upgrade", func(t transaction.Type) transaction.Transaction {
		return &UpgradeTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*UpgradeTx)
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
		tx := t.(*UpgradeTx)

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

			bds, has := GBuildingDefine[tx.AreaType]
			if !has {
				return ErrInvalidAreaType
			}
			if tx.TargetLevel == 0 || int(tx.TargetLevel) >= len(bds)+1 {
				return ErrInvalidLevel
			}

			res := gd.Resource(ctx.TargetHeight())
			bd := bds[tx.TargetLevel-1]
			if bd.CostUsage > res.Balance {
				// fmt.Println("Cost need ", bd.CostUsage, " but has ", res.Balance)
				return ErrInsufficientResource
			}
			if bd.ManUsage > res.ManRemained {
				// fmt.Println("Man need ", bd.ManUsage, " but has ", res.ManRemained)
				return ErrInsufficientResource
			}
			if bd.PowerUsage > res.PowerRemained {
				// fmt.Println("Power need ", bd.PowerUsage, " but has ", res.PowerRemained)
				return ErrInsufficientResource
			}

			idx := int(tx.X) + GTileSize*int(tx.Y)
			tile := gd.Tiles[idx]
			if tile == nil {
				return ErrInvalidLevel
			}
			if tx.AreaType != tile.AreaType {
				return ErrInvalidAreaType
			}
			if tx.TargetLevel < 2 || tx.TargetLevel > 6 {
				return ErrInvalidLevel
			}
			if tx.TargetLevel != tile.Level+1 {
				return ErrInvalidLevel
			}
			if tx.TargetLevel < 6 {
				tile.Level++
				tile.BuildHeight = ctx.TargetHeight()
			} else {
				if tx.X < 1 || tx.Y < 1 {
					return ErrInvalidPosition
				}
				if nearTile := gd.Tiles[int(tx.X)-1+GTileSize*(int(tx.Y))]; nearTile == nil || nearTile.Level != 5 {
					return ErrInvalidPosition
				}
				if nearTile := gd.Tiles[int(tx.X)+GTileSize*(int(tx.Y)-1)]; nearTile == nil || nearTile.Level != 5 {
					return ErrInvalidPosition
				}
				if nearTile := gd.Tiles[int(tx.X)-1+GTileSize*(int(tx.Y)-1)]; nearTile == nil || nearTile.Level != 5 {
					return ErrInvalidPosition
				}
				gd.Tiles[int(tx.X)-1+GTileSize*(int(tx.Y))] = nil
				gd.Tiles[int(tx.X)+GTileSize*(int(tx.Y)-1)] = nil
				gd.Tiles[int(tx.X)-1+GTileSize*(int(tx.Y)-1)] = nil
				tile.Level++
				tile.BuildHeight = ctx.TargetHeight()
			}

			gd.UpdatePoint(ctx.TargetHeight(), res.Balance-bd.CostUsage)

			MaxLevel := gd.MaxLevels[int(tx.X)+int(tx.Y)*GTileSize]
			if MaxLevel < tx.TargetLevel {
				gd.Exps = append(gd.Exps, &FletaCityExp{
					X:        tx.X,
					Y:        tx.Y,
					AreaType: tx.AreaType,
					Level:    tx.TargetLevel,
				})
				gd.MaxLevels[int(tx.X)+int(tx.Y)*GTileSize] = tx.TargetLevel
			}

			var buffer bytes.Buffer
			if _, err := gd.WriteTo(&buffer); err != nil {
				return err
			}
			ctx.SetAccountData(tx.Address, []byte("game"), buffer.Bytes())
			ctx.Commit(sn)

			return nil
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

// UpgradeTx is a fleta.UpgradeTx
// It is used to make a single account
type UpgradeTx struct {
	utxo_tx.Base
	Address     common.Address
	X           uint8
	Y           uint8
	AreaType    AreaType
	TargetLevel uint8
}

// Hash returns the hash value of it
func (tx *UpgradeTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *UpgradeTx) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := util.WriteUint8(w, tx.TargetLevel); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *UpgradeTx) ReadFrom(r io.Reader) (int64, error) {
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
		tx.TargetLevel = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *UpgradeTx) MarshalJSON() ([]byte, error) {
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

	buffer.WriteString(`"target_level":`)
	if bs, err := json.Marshal(tx.TargetLevel); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
