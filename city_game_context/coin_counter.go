package citygame

import (
	"io"
	"sort"

	"git.fleta.io/fleta/common/util"
)

// CLWriteTo is a serialization function
func CLWriteTo(w io.Writer, cl map[string]*FletaCityCoin) (int64, error) {
	var wrote int64
	if n, err := util.WriteUint16(w, uint16(len(cl))); err != nil {
		return 0, err
	} else {
		wrote += n
	}

	var keys []string
	for k := range cl {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// To perform the opertion you want
	for _, k := range keys {
		c := cl[k]
		if n, err := c.WriteTo(w); err != nil {
			return 0, err
		} else {
			wrote += n
		}
	}

	return wrote, nil
}

// CLReadFrom is a deserialization function
func CLReadFrom(r io.Reader) (map[string]*FletaCityCoin, error) {
	var read int64
	var length int
	if v, n, err := util.ReadUint16(r); err != nil {
		return nil, err
	} else {
		read += n
		length = int(v)
	}

	cl := map[string]*FletaCityCoin{}
	for i := 0; i < length; i++ {
		c := &FletaCityCoin{}
		if n, err := c.ReadFrom(r); err != nil {
			return nil, err
		} else {
			read += n
		}
		cl[c.Hash] = c
	}

	return cl, nil
}
