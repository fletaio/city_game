package citygame

import (
	"io"

	"git.fleta.io/fleta/extension/utxo_tx"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/common/hash"
	"git.fleta.io/fleta/common/util"
	"git.fleta.io/fleta/core/data"
	"git.fleta.io/fleta/core/transaction"
)

func init() {
	data.RegisterTransaction("fletacity.Construction", func(t transaction.Type) transaction.Transaction {
		return &ConstructionTx{
			UpgradeTx: &UpgradeTx{
				Base: utxo_tx.Base{
					Base: transaction.Base{
						Type_: t,
					},
					Vin: []*transaction.TxIn{},
				},
			},
		}
	}, func(loader data.Loader, t transaction.Transaction, signers []common.PublicHash) error {
		tx := t.(*ConstructionTx)
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
	}, UpgradeTxExecFunc)
}

// ConstructionTx is a fleta.ConstructionTx
// It is used to make a single account
type ConstructionTx struct {
	*UpgradeTx
	// utxo_tx.Base
	// Address     common.Address
	// X           uint8
	// Y           uint8
	// AreaType    AreaType
	// TargetLevel uint8
}

// Hash returns the hash value of it
func (tx *ConstructionTx) Hash() hash.Hash256 {
	return hash.DoubleHashByWriterTo(tx)
}

// WriteTo is a serialization function
func (tx *ConstructionTx) WriteTo(w io.Writer) (int64, error) {
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
func (tx *ConstructionTx) ReadFrom(r io.Reader) (int64, error) {
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
