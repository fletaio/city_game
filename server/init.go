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
	"github.com/fletaio/core/transaction"

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

func initChainComponent(act *data.Accounter, tran *data.Transactor) error {
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
		"consensus.CreateFormulation": &txFee{CreateFormulationTransctionType, amount.COIN.MulC(50000)},
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
	return nil
}

func initGenesisContextData(act *data.Accounter, tran *data.Transactor) (*data.ContextData, error) {
	loader := data.NewEmptyLoader(act.ChainCoord(), act, tran)
	ctd := data.NewContextData(loader, nil)

	acg := &accCoordGenerator{}
	adminPubHash := common.MustParsePublicHash("3Zmc4bGPP7TuMYxZZdUhA9kVjukdsE2S8Xpbj4Laovv")
	addUTXO(loader, ctd, adminPubHash, acg.Generate(), citygame.CreateAccountChannelSize)
	addFormulator(loader, ctd, common.MustParsePublicHash("2xASBuEWw6LcQGjYxeGZH9w1DUsEDt7fvUh8p3auxyN"), common.NewAddress(acg.Generate(), 0))
	//addFormulator(loader, ctd, common.MustParsePublicHash("2VdGunZe8yZNm2mErqQqrFx2B7Mb4SBRPWviWnapahw"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
	/*
		addFormulator(loader, ctd, common.MustParsePublicHash("3eiovnNMgNCSkmxqwkjAabRTbNkkauMVk167Pgqon2Q"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("cNXbd7o43DkX48DaEy7hzuR6iy6DBxMAqNWmhxJLyA"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("3S7zbNCsAkHJns4Z3GP6RoQKcffHDxv8fPbk1tKD2Bb"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("39q6QQ9pfiP1yEAceCu11p5cmVhG8mHMiVayCD3UEa5"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("37pB69UiK7GX1sYcawoUq8c8yXS9WWbQnkmzoQjUmZB"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("2r9mQmdfvK62ELWezK8tUvDztettkUkEGrvMWUXL7D"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("2CQBhmtferf2qWDjqSnEE3f1ECimj4Lck2CxndgqEVq"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
		addFormulator(loader, ctd, common.MustParsePublicHash("4D5m6ssnsf3NxJmqKg7PpwoyG2PdMNPAuQjpB8ZKjDo"), common.NewAddress(acg.Generate(), st.ChainCoord(), 0))
	*/
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

func addSingleAccount(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := loader.Accounter().NewByTypeName("fletacity.Account")
	if err != nil {
		panic(err)
	}
	acc := a.(*citygame.Account)
	acc.Address_ = addr
	acc.Balance_ = amount.NewCoinAmount(0, 0)
	acc.KeyHash = KeyHash
	ctd.CreatedAccountMap[acc.Address_] = acc
}

func addFormulator(loader data.Loader, ctd *data.ContextData, KeyHash common.PublicHash, addr common.Address) {
	a, err := loader.Accounter().NewByTypeName("consensus.FormulationAccount")
	if err != nil {
		panic(err)
	}
	acc := a.(*consensus.FormulationAccount)
	acc.Address_ = addr
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
