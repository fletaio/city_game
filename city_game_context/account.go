package citygame

import (
	"bytes"
	"encoding/json"
	"io"

	"git.fleta.io/fleta/common"
	"git.fleta.io/fleta/core/account"
	"git.fleta.io/fleta/core/data"
)

func init() {
	data.RegisterAccount("fletacity.Account", func(t account.Type) account.Account {
		return &Account{
			Base: account.Base{
				Type_: t,
			},
		}
	}, func(loader data.Loader, a account.Account, signers []common.PublicHash) error {
		acc := a.(*Account)
		if len(signers) != 1 {
			return ErrInvalidSignerCount
		}
		signer := signers[0]
		if !acc.KeyHash.Equal(signer) {
			return ErrInvalidAccountSigner
		}
		return nil
	})
}

// Account is a fleta.SingleAccount
// It is used as a basic account
type Account struct {
	account.Base
	KeyHash common.PublicHash
}

// Clone returns the clonend value of it
func (acc *Account) Clone() account.Account {
	return &Account{
		Base: account.Base{
			Type_:    acc.Type_,
			Address_: acc.Address_,
			Balance_: acc.Balance(),
		},
		KeyHash: acc.KeyHash.Clone(),
	}
}

// WriteTo is a serialization function
func (acc *Account) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := acc.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := acc.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (acc *Account) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := acc.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := acc.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (acc *Account) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"address":`)
	if bs, err := acc.Address_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(acc.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := acc.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
