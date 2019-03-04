package blockexplorer

import (
	"io"

	"github.com/fletaio/common/util"
)

//ScoreType is define score type
type ScoreType uint8

// score type
const (
	Level         ScoreType = 1
	Balance       ScoreType = 2
	ManProvided   ScoreType = 3
	PowerProvided ScoreType = 4
	CoinCount     ScoreType = 5
)

//ScoreCase is score struct
type ScoreCase struct {
	UserID        string
	Level         uint64
	Balance       uint64
	ManProvided   uint64
	PowerProvided uint64
	CoinCount     uint64
}

func (c *ScoreCase) getValue(s ScoreType) uint64 {
	switch s {
	case Level:
		return c.Level
	case Balance:
		return c.Balance
	case ManProvided:
		return c.ManProvided
	case PowerProvided:
		return c.PowerProvided
	case CoinCount:
		return c.CoinCount
	}
	return 0
}

func getType(s ScoreType) string {
	switch s {
	case Level:
		return "GameScore"
	case Balance:
		return "GameBalance"
	case ManProvided:
		return "GameMan"
	case PowerProvided:
		return "GamePower"
	case CoinCount:
		return "GameCoin"
	}
	return ""
}

func getTypeFromString(s string) ScoreType {
	switch s {
	case "GameScore":
		return Level
	case "GameBalance":
		return Balance
	case "GameMan":
		return ManProvided
	case "GamePower":
		return PowerProvided
	case "GameCoin":
		return CoinCount
	}
	return Level
}

// WriteTo is a serialization function
func (c *ScoreCase) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := util.WriteString(w, c.UserID); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, c.Level); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, c.Balance); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, c.ManProvided); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, c.PowerProvided); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint64(w, c.CoinCount); err != nil {
		return wrote, err
	} else {
		wrote += n
	}

	return wrote, nil
}

// ReadFrom is a deserialization function
func (c *ScoreCase) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if v, n, err := util.ReadString(r); err != nil {
		return read, err
	} else {
		read += n
		c.UserID = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		c.Level = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		c.Balance = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		c.ManProvided = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		c.PowerProvided = v
	}
	if v, n, err := util.ReadUint64(r); err != nil {
		return read, err
	} else {
		read += n
		c.CoinCount = v
	}

	return read, nil
}
