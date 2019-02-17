package citygame

import (
	_ "git.fleta.io/fleta/extension/account_tx"
	_ "git.fleta.io/fleta/extension/utxo_tx"
	_ "git.fleta.io/fleta/solidity"
)

// consts
const (
	CreateAccountChannelSize = 100
	GameAccountChannelSize   = 4
)
