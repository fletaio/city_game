package citygame

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/fletaio/common/util"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/event"

	"github.com/fletaio/common"
)

func init() {
	data.RegisterEvent("fletacity.CreateAccount", func(t event.Type) event.Event {
		return &CreateAccountEvent{
			Base: event.Base{
				Type_: t,
			},
		}
	})
}

// CreateAccountEvent is a event of adding count to the account
type CreateAccountEvent struct {
	event.Base
	Address common.Address
	KeyHash common.PublicHash
	UserID  string
	Reward  string
	Comment string
}

// WriteTo is a serialization function
func (e *CreateAccountEvent) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := e.Base.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := e.Address.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := e.KeyHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, e.UserID); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, e.Reward); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, e.Comment); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (e *CreateAccountEvent) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := e.Base.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := e.Address.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := e.KeyHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		e.UserID = v
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		e.Reward = v
	}
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		e.Comment = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (e *CreateAccountEvent) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"coord":`)
	if bs, err := e.Coord_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"index":`)
	if bs, err := json.Marshal(e.Index_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"type":`)
	if bs, err := json.Marshal(e.Type_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"address":`)
	if bs, err := json.Marshal(e.Address); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := json.Marshal(e.KeyHash); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"user_id":`)
	if bs, err := json.Marshal(e.UserID); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"reward":`)
	if bs, err := json.Marshal(e.Reward); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"comment":`)
	if bs, err := json.Marshal(e.Comment); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
