package main

import (
	"github.com/fletaio/core/data"

	"github.com/fletaio/common"
)

type mockRewarder struct {
}

// ProcessReward gives a reward to the block generator address
func (rd *mockRewarder) ProcessReward(addr common.Address, ctx *data.Context) error {
	return nil
}
