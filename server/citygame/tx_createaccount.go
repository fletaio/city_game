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

// prefix values
var (
	PrefixKeyHash = []byte("k:")
	PrefixUserID  = []byte("i:")
)

var allowedPublicHashMap = map[uint64]common.PublicHash{}

// RegisterAllowedPublicHash TODO
func RegisterAllowedPublicHash(ChainCoord *common.Coordinate, pubhash common.PublicHash) {
	allowedPublicHashMap[ChainCoord.ID()] = pubhash
}

// UnregisterAllowedPublicHash  TODO
func UnregisterAllowedPublicHash(ChainCoord *common.Coordinate) {
	delete(allowedPublicHashMap, ChainCoord.ID())
}

func init() {
	data.RegisterTransaction("fletacity.CreateAccount", func(t transaction.Type) transaction.Transaction {
		return &CreateAccountTx{
			Base: utxo_tx.Base{
				Base: transaction.Base{
					Type_: t,
				},
				Vin: []*transaction.TxIn{},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*CreateAccountTx)
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
			pubhash := allowedPublicHashMap[loader.ChainCoord().ID()]
			if !utxo.PublicHash.Equal(pubhash) {
				return ErrNotAllowed
			}
		}
		return nil
	}, func(ctx *data.Context, Fee *amount.Amount, t transaction.Transaction, coord *common.Coordinate) (interface{}, error) {
		tx := t.(*CreateAccountTx)
		sn := ctx.Snapshot()
		defer ctx.Revert(sn)

		utxo, err := ctx.UTXO(tx.Vin[0].ID())
		if err != nil {
			return nil, err
		}

		if err := ctx.DeleteUTXO(utxo.ID()); err != nil {
			return nil, err
		}

		func() error {
			sn := ctx.Snapshot()
			defer ctx.Revert(sn)

			addr := common.NewAddress(coord, 0)
			if is, err := ctx.IsExistAccount(addr); err != nil {
				return err
			} else if is {
				return ErrExistAddress
			} else if isn, err := ctx.IsExistAccountName(tx.UserID); err != nil {
				return err
			} else if isn {
				return ErrExistAccountName
			}

			KeyHashID := append(PrefixKeyHash, tx.KeyHash[:]...)
			UserIDHashID := append(PrefixUserID, []byte(tx.UserID)...)

			var rootAddress common.Address
			if bs := ctx.AccountData(rootAddress, KeyHashID); len(bs) > 0 {
				return ErrExistKeyHash
			}
			if bs := ctx.AccountData(rootAddress, UserIDHashID); len(bs) > 0 {
				return ErrExistKeyHash
			}

			a, err := ctx.Accounter().NewByTypeName("fletacity.Account")
			if err != nil {
				return err
			}
			acc := a.(*Account)
			acc.Address_ = addr
			acc.Name_ = tx.UserID
			acc.KeyHash = tx.KeyHash
			acc.Height = ctx.TargetHeight()
			ctx.CreateAccount(acc)

			gd := NewGameData(ctx.TargetHeight())
			h := hash.Hash(util.Uint32ToBytes(ctx.TargetHeight()))
			for i := 0; i < 5; i++ {
				x := uint8(util.BytesToUint16([]byte(h[i*2:i*2+2]))) % GTileSize
				y := uint8(util.BytesToUint16([]byte(h[i*2+2:i*2+4]))) % GTileSize
				gd.Coins = append(gd.Coins, &FletaCityCoin{
					X:      x,
					Y:      y,
					Index:  uint8(i),
					Height: ctx.TargetHeight() + TimeCoinGenTime*2,
				})
			}
			var buffer bytes.Buffer
			if _, err := gd.WriteTo(&buffer); err != nil {
				return err
			}
			ctx.SetAccountData(addr, []byte("game"), buffer.Bytes())
			ctx.SetAccountData(addr, []byte("comment"), []byte(tx.Comment))

			ctx.SetAccountData(rootAddress, KeyHashID, addr[:])
			ctx.SetAccountData(rootAddress, UserIDHashID, addr[:])

			for i := 0; i < GameCommandChannelSize; i++ {
				id := transaction.MarshalID(coord.Height, coord.Index, uint16(i+1))
				ctx.CreateUTXO(id, &transaction.TxOut{
					Amount:     amount.NewCoinAmount(0, 0),
					PublicHash: tx.KeyHash,
				})
				ctx.SetAccountData(addr, []byte("utxo"+strconv.Itoa(i)), util.Uint64ToBytes(id))
			}

			e, err := ctx.Eventer().NewByTypeName("fletacity.CreateAccount")
			if err != nil {
				return err
			}
			ev := e.(*CreateAccountEvent)
			ev.Base.Coord_ = coord
			ev.Address = addr
			ev.KeyHash = tx.KeyHash
			ev.UserID = tx.UserID
			ev.Reward = tx.Reward
			ev.Comment = tx.Comment

			ctx.EmitEvent(e)

			ctx.Commit(sn)
			return nil
		}()

		var rootAddress common.Address
		for i := 0; i < CreateAccountChannelSize; i++ {
			did := []byte("utxo" + strconv.Itoa(i))
			oldid := util.BytesToUint64(ctx.AccountData(rootAddress, did))
			if oldid == tx.Vin[0].ID() {
				id := transaction.MarshalID(coord.Height, coord.Index, 0)
				ctx.CreateUTXO(id, &transaction.TxOut{
					Amount:     amount.NewCoinAmount(0, 0),
					PublicHash: utxo.PublicHash,
				})
				ctx.SetAccountData(rootAddress, did, util.Uint64ToBytes(id))
				break
			}
		}

		ctx.Commit(sn)
		return nil, nil
	})
}

// CreateAccountTx is a fleta.CreateAccountTx
// It is used to make a single account
type CreateAccountTx struct {
	utxo_tx.Base
	KeyHash common.PublicHash
	UserID  string `json:"user_id"`
	Reward  string `json:"reward"`
	Comment string `json:"comment"`
}

// Hash returns the hash value of it
func (tx *CreateAccountTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *CreateAccountTx) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := tx.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := tx.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, tx.UserID); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, tx.Reward); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, tx.Comment); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (tx *CreateAccountTx) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := tx.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := tx.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		tx.UserID = v
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Reward = v
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		tx.Comment = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (tx *CreateAccountTx) MarshalJSON() ([]byte, error) {
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

	buffer.WriteString(`"key_hash":`)
	if bs, err := json.Marshal(tx.KeyHash); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)

	buffer.WriteString(`"user_id":`)
	if bs, err := json.Marshal(tx.UserID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)

	buffer.WriteString(`"reward":`)
	if bs, err := json.Marshal(tx.Reward); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)

	buffer.WriteString(`"comment":`)
	if bs, err := json.Marshal(tx.Comment); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
