package citygame

import (
	"errors"
)

// account_tx errors
var (
	ErrInvalidSequence             = errors.New("invalid sequence")
	ErrInvalidTransactionSignature = errors.New("invalid transaction signature")
	ErrInvalidSignerCount          = errors.New("invalid signer count")
	ErrInvalidAccountSigner        = errors.New("invalid account signer")
	ErrInvalidLevel                = errors.New("invalid level")
	ErrInvalidAreaType             = errors.New("invalid area type")
	ErrInvalidCoinType             = errors.New("invalid coin type")
	ErrInvalidDemolition           = errors.New("invalid demolition")
	ErrInvalidPosition             = errors.New("invalid position")
	ErrInvalidAddress              = errors.New("invalid address")
	ErrInvalidPublicKey            = errors.New("invalid public key")
	ErrInvalidReward               = errors.New("invalid reward")
	ErrInvalidTxInCount            = errors.New("invalid txin count")
	ErrInvalidUTXO                 = errors.New("invalid utxo")
	ErrShortUserID                 = errors.New("short userid")
	ErrNotAllowed                  = errors.New("not allowed")
	ErrNotExistTile                = errors.New("not exist tile")
	ErrNotExistAccount             = errors.New("not exist account")
	ErrExistAddress                = errors.New("exist address")
	ErrExistKeyHash                = errors.New("exist key hash")
	ErrExistUserID                 = errors.New("exist userid")
	ErrExistReward                 = errors.New("exist reward")
	ErrQueueFull                   = errors.New("queue full")
	ErrInsufficientResource        = errors.New("insufficient resource")
	ErrTimeCoinNotExist            = errors.New("time coin not exist")
	ErrTypeMissMatch               = errors.New("type miss match")
	ErrNotExistGameData            = errors.New("not exist game data")
)
