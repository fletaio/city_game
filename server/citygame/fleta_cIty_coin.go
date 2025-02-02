package citygame

import (
	"io"

	"github.com/fletaio/common/util"
)

//FletaCityCoin is FletaCityCoin
type FletaCityCoin struct {
	X      uint8  `json:"x"`
	Y      uint8  `json:"y"`
	Index  uint8  `json:"index"`
	Height uint32 `json:"height"`
}

// WriteTo is a serialization function
func (f *FletaCityCoin) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint8(w, f.X); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, f.Y); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, f.Index); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, f.Height); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (f *FletaCityCoin) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		f.X = v
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		f.Y = v
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		f.Index = v
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		f.Height = v
	}
	return read, nil
}
