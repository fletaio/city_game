package main

import (
	"git.fleta.io/fleta/core/data"

	"git.fleta.io/fleta/common"
)

type mockRewarder struct {
}

// ProcessReward gives a reward to the block generator address
func (rd *mockRewarder) ProcessReward(addr common.Address, ctx *data.Context) error {
	return nil
}
