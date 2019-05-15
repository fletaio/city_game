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
	data.RegisterEvent("fletacity.GetExp", func(t event.Type) event.Event {
		return &GetExpEvent{
			Base: event.Base{
				Type_: t,
			},
		}
	})
}

// GetExpEvent is a event of adding count to the account
type GetExpEvent struct {
	event.Base
	Address  common.Address
	X        uint8
	Y        uint8
	TotalExp uint64
}

// WriteTo is a serialization function
func (e *GetExpEvent) WriteTo(w io.Writer) (int64, error) {
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
	if n, err := util.WriteUint64(w, e.TotalExp); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (e *GetExpEvent) ReadFrom(r io.Reader) (int64, error) {
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
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		e.TotalExp = v
	}
	return read, nil
}

// MarshalJSON is a marshaler function
func (e *GetExpEvent) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"total_exp":`)
	if bs, err := json.Marshal(e.TotalExp); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
