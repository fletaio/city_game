package citygame

import (
	"io"

	"git.fleta.io/fleta/common/util"
)

//FletaCityCoin is FletaCityCoin
type FletaCityCoin struct {
	X        int      `json:"x"`
	Y        int      `json:"y"`
	Hash     string   `json:"hash"`
	Height   uint32   `json:"height"`
	CoinType CoinType `json:"coin_type"`
}

// WriteTo is a serialization function
func (f *FletaCityCoin) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint32(w, uint32(f.X)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, uint32(f.Y)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteString(w, f.Hash); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, f.Height); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint8(w, uint8(f.CoinType)); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (f *FletaCityCoin) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		f.X = int(v)
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		f.Y = int(v)
	}

	if s, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		f.Hash = s
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		f.Height = v
	}
	if v, n, err := util.ReadUint8(r); err != nil {
		return read, err
	} else {
		read += n
		f.CoinType = CoinType(v)
	}
	return read, nil
}
