package citygame

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"git.fleta.io/fleta/extension/utxo_tx"

	"git.fleta.io/fleta/core/amount"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fletacity.GetCoin", func(coord *common.Coordinate, t transaction.Type) transaction.Transaction {
		return &GetCoinTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					ChainCoord_: coord,
					Type_:       t,
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

			if tx.CoinType == ConstructCoinType {
				clbs := ctx.AccountData(tx.Address, []byte("CoinList"))
				bf := bytes.NewBuffer(clbs)
				if cl, err := CLReadFrom(bf); err != nil {
					return err
				} else {
					for i, c := range cl {
						if c.X == int(tx.X) && c.Y == int(tx.Y) && c.Hash == tx.TargetHash.String() && c.Height == tx.TargetHeight {
							delete(cl, i)

							bf := &bytes.Buffer{}
							_, err := CLWriteTo(bf, cl)
							if err != nil {
								return err
							}
							ctx.SetAccountData(tx.Address, []byte("CoinList"), bf.Bytes())
							bs := ctx.AccountData(tx.Address, []byte("GetCoinCount"))
							var coinCount uint32
							if len(bs) == 4 {
								coinCount = util.BytesToUint32(bs)
							} else {
								coinCount = 1
							}
							ctx.SetAccountData(tx.Address, []byte("GetCoinCount"), util.Uint32ToBytes(coinCount+1))

							ctx.Commit(sn)
							return nil
						}
					}
					return ErrTimeCoinNotExist
				}
			} else if tx.CoinType == TimeCoinType {
				clbs := ctx.AccountData(tx.Address, []byte("CoinList"))
				bf := bytes.NewBuffer(clbs)
				if cl, err := CLReadFrom(bf); err != nil {
					return err
				} else {
					// tl := CalcTargetCoinList(startHeight, cl)
					for i, c := range cl {
						if c.X == int(tx.X) && c.Y == int(tx.Y) && c.Hash == tx.TargetHash.String() && c.Height == tx.TargetHeight && c.Height < ctx.TargetHeight() {
							delete(cl, i)

							h := tx.Hash()
							x := int(util.BytesToUint16([]byte(h[0:2]))) % GTileSize
							y := int(util.BytesToUint16([]byte(h[2:4]))) % GTileSize

							if _, has := cl[h.String()]; has {
								panic("has")
							}

							cl[h.String()] = &FletaCityCoin{
								X:        int(x),
								Y:        int(y),
								Hash:     h.String(),
								Height:   ctx.TargetHeight() + TimeCoinGenTime,
								CoinType: TimeCoinType,
							}
							bf := &bytes.Buffer{}
							_, err := CLWriteTo(bf, cl)
							if err != nil {
								return err
							}
							ctx.SetAccountData(tx.Address, []byte("CoinList"), bf.Bytes())
							bs := ctx.AccountData(tx.Address, []byte("GetCoinCount"))
							var coinCount uint32
							if len(bs) == 4 {
								coinCount = util.BytesToUint32(bs)
							} else {
								coinCount = 1
							}
							ctx.SetAccountData(tx.Address, []byte("GetCoinCount"), util.Uint32ToBytes(coinCount+1))
							ctx.Commit(sn)
							return nil
						}
					}

					return ErrTimeCoinNotExist
				}
			}

			//TODO account 에서 fleta city coin이 있는지 검사

			return ErrTypeMissMatch
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

// GetCoinTx is a fleta.GetCoinTx
type GetCoinTx struct {
	utxo_tx.Base
	Address      common.Address
	X            uint8
	Y            uint8
	TargetHeight uint32
	TargetHash   hash.Hash256
	CoinType     CoinType
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
	if n, err := util.WriteUint32(w, tx.TargetHeight); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.TargetHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(tx.CoinType)); err != nil {
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
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		tx.TargetHeight = v
	}
	if n, err := tx.TargetHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		tx.CoinType = CoinType(v)
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *GetCoinTx) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"chain_coord":`)
	if bs, err := tx.ChainCoord_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
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

	buffer.WriteString(`"target_height":`)
	if bs, err := json.Marshal(tx.TargetHeight); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)

	buffer.WriteString(`"target_hash":`)
	if bs, err := json.Marshal(tx.TargetHash); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)

	buffer.WriteString(`"coin_type":`)
	if bs, err := json.Marshal(tx.CoinType); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
