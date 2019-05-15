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
	data.RegisterEvent("fletacity.GetCoin", func(t event.Type) event.Event {
		return &GetCoinEvent{
			Base: event.Base{
				Type_: t,
			},
		}
	})
}

// GetCoinEvent is a event of adding count to the account
type GetCoinEvent struct {
	event.Base
	Address   common.Address
	X         uint8
	Y         uint8
	CoinCount uint32
}

// WriteTo is a serialization function
func (e *GetCoinEvent) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := util.WriteUint8(w, e.X); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, e.Y); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, e.CoinCount); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (e *GetCoinEvent) ReadFrom(r io.Reader) (int64, error) {
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
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		e.X = v
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		e.Y = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		e.CoinCount = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (e *GetCoinEvent) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"x":`)
	if bs, err := json.Marshal(e.X); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"y":`)
	if bs, err := json.Marshal(e.Y); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"coin_count":`)
	if bs, err := json.Marshal(e.CoinCount); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
