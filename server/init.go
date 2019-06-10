package main

import (
	"log"
	"strconv"

	"github.com/fletaio/citygame/server/citygame"
	"github.com/fletaio/common"
	"github.com/fletaio/common/util"
	"github.com/fletaio/core/account"
	"github.com/fletaio/core/amount"
	"github.com/fletaio/core/consensus"
	"github.com/fletaio/core/data"
	"github.com/fletaio/core/event"
	"github.com/fletaio/core/transaction"

	"github.com/fletaio/extension/account_def"
	_ "github.com/fletaio/extension/account_tx"
	_ "github.com/fletaio/extension/utxo_tx"
	_ "github.com/fletaio/solidity"
)

// consts
const (
	BlockchainVersion = 1
)

// transaction_type transaction types
const (
	// Game Transactions
	CreateAccountTransctionType = transaction.Type(1)
	DemolitionTransactionType   = transaction.Type(2)
	ConstructionTransactionType = transaction.Type(3)
	UpgradeTransactionType      = transaction.Type(4)
	GetCoinTransactionType      = transaction.Type(5)
	GetExpTransactionType       = transaction.Type(6)
	// Formulation Transactions
	CreateFormulationTransctionType = transaction.Type(60)
	RevokeFormulationTransctionType = transaction.Type(61)
)

// account_type account types
const (
	// Game Accounts
	AccountType = account.Type(1)
	// Formulation Accounts
	FormulationAccountType = account.Type(60)
)

// event_type event types
const (
	// Game Events
	CreateAccountEventType = event.Type(1)
	ConstructionEventType  = event.Type(2)
	UpgradeEventType       = event.Type(3)
	DemolitionEventType    = event.Type(4)
	GetCoinEventType       = event.Type(5)
	GetExpEventType        = event.Type(6)
)

func initChainComponent(act *data.Accounter, tran *data.Transactor, evt *data.Eventer) error {
	type txFee struct {
		Type transaction.Type
		Fee  *amount.Amount
	}

	TxFeeTable := map[string]*txFee{
		"fletacity.CreateAccount":     &txFee{CreateAccountTransctionType, amount.COIN.MulC(10)},
		"fletacity.Demolition":        &txFee{DemolitionTransactionType, amount.COIN.MulC(10)},
		"fletacity.Construction":      &txFee{ConstructionTransactionType, amount.COIN.MulC(10)},
		"fletacity.Upgrade":           &txFee{UpgradeTransactionType, amount.COIN.MulC(10)},
		"fletacity.GetCoin":           &txFee{GetCoinTransactionType, amount.COIN.MulC(10)},
		"fletacity.GetExp":            &txFee{GetExpTransactionType, amount.COIN.MulC(10)},
		"consensus.CreateFormulation": &txFee{CreateFormulationTransctionType, amount.COIN.DivC(10)},
		"consensus.RevokeFormulation": &txFee{RevokeFormulationTransctionType, amount.COIN.DivC(10)},
	}
	for name, item := range TxFeeTable {
		if err := tran.RegisterType(name, item.Type, item.Fee); err != nil {
			log.Println(name, item, err)
			return err
		}
	}

	AccTable := map[string]account.Type{
		"fletacity.Account":            AccountType,
		"consensus.FormulationAccount": FormulationAccountType,
	}
	for name, t := range AccTable {
		if err := act.RegisterType(name, t); err != nil {
			log.Println(name, t, err)
			return err
		}
	}

	EventTable := map[string]event.Type{
		"fletacity.CreateAccount": CreateAccountEventType,
		"fletacity.Construction":  ConstructionEventType,
		"fletacity.Upgrade":       UpgradeEventType,
		"fletacity.Demolition":    DemolitionEventType,
		"fletacity.GetCoin":       GetCoinEventType,
		"fletacity.GetExp":        GetExpEventType,
	}
	for name, t := range EventTable {
		if err := evt.RegisterType(name, t); err != nil {
			log.Println(name, t, err)
			return err
		}
	}
	return nil
}

func initGenesisContextData(act *data.Accounter, tran *data.Transactor, evt *data.Eventer) (*data.ContextData, error) {
	consensus.SetFormulatorPolicy(act.ChainCoord(), &consensus.FormulatorPolicy{
		CreateFormulationAmount: amount.NewCoinAmount(200000, 0),
		OmegaRequiredLockBlocks: 5184000,
		SigmaRequiredLockBlocks: 5184000,
	})

	loader := data.NewEmptyLoader(act.ChainCoord(), act, tran, evt)
	ctd := data.NewContextData(loader, nil)

	acg := &accCoordGenerator{}
	adminPubHash := common.MustParsePublicHash("Xy8u3LBXwKZ2F61UuzEcyBs11avYjkJQCGah2LvKNe")
	addUTXO(loader, ctd, adminPubHash, acg.Generate(), citygame.CreateAccountChannelSize)
	addFormulator(loader, ctd, common.MustParsePublicHash("2saGsDpsdZX6gNH5Zi3fWVBAWDuhNo4MuLbop1QeF2u"), common.MustParseAddress("3CUsUpvEK"), "citygame.fr00001")
	addFormulator(loader, ctd, common.MustParsePublicHash("4m6XsJbq6EFb5bqhZuKFc99SmF86ymcLcRPwrWyToHQ"), common.MustParseAddress("5PxjxeqTd"), "citygame.fr00002")
	addFormulator(loader, ctd, common.MustParsePublicHash("o1rVoXHFuz5EtwLwCLcrmHpqPdugAnWHEVVMtnCb32"), common.MustParseAddress("7bScSUkgw"), "citygame.fr00003")
	addFormulator(loader, ctd, common.MustParsePublicHash("47NZ8oadY4dCAM3ZrGFrENPn99L1SLSqzpR4DFPUpk5"), common.MustParseAddress("9nvUvJfvF"), "citygame.fr00004")
	addFormulator(loader, ctd, common.MustParsePublicHash("4TaHVFSzcrNPktRiNdpPitoUgLXtZzrVmkxE3GmcYjG"), common.MustParseAddress("BzQMQ8b9Z"), "citygame.fr00005")
	addFormulator(loader, ctd, common.MustParsePublicHash("2wqsb4J47T4JkNUp1Bma1HkjpCyei7sZinLmNprpdtY"), common.MustParseAddress("EBtDsxWNs"), "citygame.fr00006")
	addFormulator(loader, ctd, common.MustParsePublicHash("2a1CirwCHSYYpLqpbi1b7Rpr4BAJZvydbDA1bGjJ7FG"), common.MustParseAddress("GPN6MnRcB"), "citygame.fr00007")
	addFormulator(loader, ctd, common.MustParsePublicHash("2KnMHH973ZLicENxcsJbARdeTUiYZmN3WnBzbZqvvEx"), common.MustParseAddress("JaqxqcLqV"), "citygame.fr00008")
	addFormulator(loader, ctd, common.MustParsePublicHash("4fyTmraz8x3NKWnj4nWgPWKy8qCBF1hyqVJQeyupHAe"), common.MustParseAddress("LnKqKSG4o"), "citygame.fr00009")
	addFormulator(loader, ctd, common.MustParsePublicHash("2V1zboMnJbJdeLvRBRFVPvVqs8CCmjxToBpGJSNScu2"), common.MustParseAddress("NyohoGBJ7"), "citygame.fr00010")
	addFormulator(loader, ctd, common.MustParsePublicHash("3pEYkEgXoPUm4vdcGBXP46q1BpMj215uVQdAg6P4g74"), common.MustParseAddress("RBHaH66XR"), "citygame.fr00011")
	addFormulator(loader, ctd, common.MustParsePublicHash("rsUoPRfVgXJFuV6wYcy4M4kntvr3tooeXzcRhrjBq6"), common.MustParseAddress("TNmSkv1kj"), "citygame.fr00012")
	addFormulator(loader, ctd, common.MustParsePublicHash("4UMYzaBeXEKcm6hnDDEMqYRR5NLwGndCLksryVj98Fw"), common.MustParseAddress("VaFKEjvz3"), "citygame.fr00013")
	addFormulator(loader, ctd, common.MustParsePublicHash("3h2Lt2uYFMqVQKFgKszLJzwaLhQ5kt1nMcg8M758aLh"), common.MustParseAddress("XmjBiZrDM"), "citygame.fr00014")
	addFormulator(loader, ctd, common.MustParsePublicHash("4NkvvfPdHHvpo9YTkAQBrGxpnnML2pVRXHdLgzB2EYe"), common.MustParseAddress("ZyD4CPmSf"), "citygame.fr00015")
	addFormulator(loader, ctd, common.MustParsePublicHash("3ae9sCuM75vAheVLNp3DjQqDiD3TaxY5HYduHvsgzYZ"), common.MustParseAddress("cAgvgDgfy"), "citygame.fr00016")
	addFormulator(loader, ctd, common.MustParsePublicHash("2bR5L2ZSqKLUFQzdhzWV6e4BUupHPGDFtnZUNrZBZbZ"), common.MustParseAddress("eNAoA3buH"), "citygame.fr00017")
	addFormulator(loader, ctd, common.MustParsePublicHash("BPqzvcrYi364mm6GyraHHqJHrvEfqjwo1jEC8crTxZ"), common.MustParseAddress("gZefdsX8b"), "citygame.fr00018")
	addFormulator(loader, ctd, common.MustParsePublicHash("2vtYXNUAtBtt4fF6DEbVKNc7bGhA7yBbatTA6Ye9kMT"), common.MustParseAddress("im8Y7hSMu"), "citygame.fr00019")
	addFormulator(loader, ctd, common.MustParsePublicHash("42TUBLNb1natk7s7qsHNqxHwn7Pb3pNmTfTnd1sDQnb"), common.MustParseAddress("kxcQbXMbD"), "citygame.fr00020")
	addFormulator(loader, ctd, common.MustParsePublicHash("2yng1DwwBqMixjCnjx6Pdf9o5AkgEzkumxJySr8Qe6C"), common.MustParseAddress("oA6H5MGpX"), "citygame.fr00021")
	addFormulator(loader, ctd, common.MustParsePublicHash("3PNrAwb7FrvKeB1hCxYADwNxqWuYmaqoc8E8VjdBC"), common.MustParseAddress("qMa9ZBC3q"), "citygame.fr00022")
	addFormulator(loader, ctd, common.MustParsePublicHash("2eZAofvjk5AHUpaUyC7EDx3K8KAHUQNXMynHG7ZYFfn"), common.MustParseAddress("sZ42317H9"), "citygame.fr00023")
	addFormulator(loader, ctd, common.MustParsePublicHash("4QT4FGpoaFkPiRaZQCKDfrANWJ6EAqavqkQfGr6g4oG"), common.MustParseAddress("ukXtWq2WT"), "citygame.fr00024")
	addFormulator(loader, ctd, common.MustParsePublicHash("2nPZHDpFavW2VjnZGs7ZeQyFM19y517ZTQaTgqe3G69"), common.MustParseAddress("wx1kzewjm"), "citygame.fr00025")
	addFormulator(loader, ctd, common.MustParsePublicHash("bB88uMhpM4vjUHpV5WZqfQBh4kyi6wnnKCtVF4AE2D"), common.MustParseAddress("z9VdUUry5"), "citygame.fr00026")
	addFormulator(loader, ctd, common.MustParsePublicHash("2ZLEXwQ9pqvaATFttkkNWY2CGDHdJFa5V3GNapKeqtx"), common.MustParseAddress("22LyVxJnCP"), "citygame.fr00027")
	addFormulator(loader, ctd, common.MustParsePublicHash("4M2KFgmWSKu8JyjhkmVJ8U4hjtn9MX4rsch4ZoE1i32"), common.MustParseAddress("24YTNS8hRh"), "citygame.fr00028")
	addFormulator(loader, ctd, common.MustParsePublicHash("XG9nFJsdMpo6D6wYxYSyH5zAtnvsMjySFHp1XjCouY"), common.MustParseAddress("26jwEuxcf1"), "citygame.fr00029")
	addFormulator(loader, ctd, common.MustParsePublicHash("3uW4bb1kAx35ndj4ZVLMF8xWYercS2RfP7moxZvUm8Y"), common.MustParseAddress("28wR7PnXtK"), "citygame.fr00030")
	addFormulator(loader, ctd, common.MustParsePublicHash("4mY5G1BZuZaeHR5cH1K4sUNmccPa11JkHtjv5ctde3K"), common.MustParseAddress("2B8tyscT7d"), "citygame.fr00031")
	addFormulator(loader, ctd, common.MustParsePublicHash("3oocpeXtqUZeaut1A71fbCMBQefMFMCBt2BpamNZfA9"), common.MustParseAddress("2DLNrMSNLw"), "citygame.fr00032")
	addFormulator(loader, ctd, common.MustParsePublicHash("4wknRQ86rTcN1cQbXZfbCMkqXcS1FsYG8ihAYFhmxF"), common.MustParseAddress("2FXriqGHaF"), "citygame.fr00033")
	addFormulator(loader, ctd, common.MustParsePublicHash("3mT9SNvGscpwmDjHnojnVysd9pXUvg1fenVyiBFYTDs"), common.MustParseAddress("2HjLbK6CoZ"), "citygame.fr00034")
	addFormulator(loader, ctd, common.MustParsePublicHash("24zn1BgQBmMD8dWap9XbBHdZAivDppVhnYxzZ4ftZw4"), common.MustParseAddress("2KvpTnv82s"), "citygame.fr00035")
	addFormulator(loader, ctd, common.MustParsePublicHash("4TKCbNqM68vKmmXiMsjdb7qND8Qy1DCJKvFge7Dhw16"), common.MustParseAddress("2N8JLGk3GB"), "citygame.fr00036")
	citygame.RegisterAllowedPublicHash(loader.ChainCoord(), adminPubHash)
	return ctd, nil
}

func addUTXO(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, coord *common.Coordinate, count int) {
	var rootAddress common.Address
	for i := 0; i < count; i++ {
		id := transaction.MarshalID(coord.Height, coord.Index, uint16(i))
		ctd.CreateUTXO(id, &transaction.TxOut{Amount: amount.NewCoinAmount(0, 0), PublicHash: KeyHash})
		ctd.SetAccountData(rootAddress, []byte("utxo"+strconv.Itoa(i)), util.Uint64ToBytes(id))
	}
}

func addSingleAccount(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address, name string) {
	a, err := loader.Accounter().NewByTypeName("fleta.SingleAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*account_def.SingleAccount)
	acc.Address_ = addr
	acc.Name_ = name
	acc.Balance_ = amount.NewCoinAmount(10000000000, 0)
	acc.KeyHash = KeyHash
	ctd.CreatedAccountMap[acc.Address_] = acc
}

func addFormulator(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address, name string) {
	a, err := loader.Accounter().NewByTypeName("consensus.FormulationAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*consensus.FormulationAccount)
	acc.Address_ = addr
	acc.Name_ = name
	acc.Balance_ = amount.NewCoinAmount(0, 0)
	acc.KeyHash = KeyHash
	ctd.CreatedAccountMap[acc.Address_] = acc
}

type accCoordGenerator struct {
	idx uint16
}

func (acg *accCoordGenerator) Generate() *common.Coordinate {
	coord := common.NewCoordinate(0, acg.idx)
	acg.idx++
	return coord
}
