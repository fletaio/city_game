package citygame

import (
	_ "github.com/fletaio/extension/account_tx"
	_ "github.com/fletaio/extension/utxo_tx"
	_ "github.com/fletaio/solidity"
)

// consts
const (
	CreateAccountChannelSize = 100
	GameCommandChannelSize   = 8
)
