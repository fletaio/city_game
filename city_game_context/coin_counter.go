package citygame

import (
	"io"

	"git.fleta.io/fleta/common/util"
)

// CLWriteTo is a serialization function
func CLWriteTo(w io.Writer, cl []*FletaCityCoin) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint16(w, uint16(len(cl))); err != nil {
		return 0, err
	} else {
		wrote += n
	}
	for i := 0; i < len(cl); i++ {
		if n, err := cl[i].WriteTo(w); err != nil {
			return 0, err
		} else {
			wrote += n
		}
	}

	return wrote, nil
}

// CLReadFrom is a deserialization function
func CLReadFrom(r io.Reader) ([]*FletaCityCoin, error) {
	var read int64
	var length int
	if v, n, err := util.ReadUint16(r); err != nil {
		return nil, err
	} else {
		read += n
		length = int(v)
	}

	cl := []*FletaCityCoin{}
	for i := 0; i < length; i++ {
		c := &FletaCityCoin{}
		if n, err := c.ReadFrom(r); err != nil {
			return nil, err
		} else {
			read += n
		}
		cl = append(cl, c)
	}

	return cl, nil
}
