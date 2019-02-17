package main

import (
	"git.fleta.io/fleta/core/amount"
	"git.fleta.io/fleta/core/data"

	"git.fleta.io/fleta/common"
)

type mockRewarder struct {
}

// ProcessReward gives a reward to the block generator address
func (rd *mockRewarder) ProcessReward(addr common.Address, ctx *data.Context) error {
	balance, err := ctx.AccountBalance(addr)
	if err != nil {
		return err
	}
	balance.AddBalance(ctx.ChainCoord(), amount.NewCoinAmount(1, 0))
	return nil
}
