package blockexplorer

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.fleta.io/fleta/common"

	"git.fleta.io/fleta/core/block"
	"github.com/dgraph-io/badger"
)

func (e *BlockExplorer) formulators() []countInfo {
	return e.formulatorCountList
}
func (e *BlockExplorer) transactions() []countInfo {
	return e.transactionCountList
}
func (e *BlockExplorer) chainInfo() currentChainInfo {
	return e.CurrentChainInfo
}

type typePerBlock struct {
	BlockTime uint64 `json:"blockTime"`
	Symbol    string `json:"symbol"`
	TxCount   string `json:"txCount"`
	// Types     map[string]int `json:"types"`
}

type blockInfos struct {
	BlockHeight uint32   `json:"Block Height"`
	BlockHash   string   `json:"Block Hash"`
	Time        string   `json:"Time"`
	Status      string   `json:"Status"`
	Txs         string   `json:"Txs"`
	Formulator  string   `json:"Formulator"`
	Msg         string   `json:"Msg"`
	Signs       []string `json:"Signs"`
	BlockCount  uint32   `json:"BlockCount"`
}
type blockInfosCase struct {
	ITotalRecords        int          `json:"iTotalRecords"`
	ITotalDisplayRecords int          `json:"iTotalDisplayRecords"`
	SEcho                int          `json:"sEcho"`
	SColumns             string       `json:"sColumns"`
	AaData               []blockInfos `json:"aaData"`
}

func (e *BlockExplorer) lastestBlocks() (result blockInfosCase) {
	currHeight := e.Kernel.Provider().Height()

	result.AaData = []blockInfos{}

	for i := currHeight; i > 0 && i > currHeight-8; i-- {
		b, err := e.Kernel.Block(i)
		if err != nil {
			continue
		}

		b.Header.Hash().String()

		cd, err := e.Kernel.Provider().Data(i)
		if err != nil {
			continue
		}
		status := 1
		if b.Header.TimeoutCount > 0 {
			status = 2
		}

		bs := block.Signed{
			HeaderHash:         cd.Header.Hash(),
			GeneratorSignature: cd.Signatures[0],
		}
		Signs := []string{
			cd.Signatures[1].String(),
			cd.Signatures[2].String(),
			cd.Signatures[3].String(),
		}

		tm := time.Unix(int64(cd.Header.Timestamp()/uint64(time.Second)), 0)

		result.AaData = append(result.AaData, blockInfos{
			BlockHeight: i,
			BlockHash:   cd.Header.Hash().String(),
			Time:        tm.Format("2006-01-02 15:04:05"),
			Status:      strconv.Itoa(status),
			Txs:         strconv.Itoa(len(b.Body.Transactions)),
			Formulator:  b.Header.Formulator.String(),
			Msg:         bs.Hash().String(),
			Signs:       Signs,
			BlockCount:  e.GetBlockCount(b.Header.Formulator.String()),
		})
	}

	result.ITotalRecords = len(result.AaData)
	result.ITotalDisplayRecords = len(result.AaData)

	return
}

type txInfos struct {
	TxHash    string `json:"TxHash"`
	BlockHash string `json:"BlockHash"`
	ChainID   string `json:"ChainID"`
	Time      uint64 `json:"Time"`
	TxType    string `json:"TxType"`
}

func (e *BlockExplorer) lastestTransactions() []txInfos {
	if len(e.lastestTransactionList) < 8 {
		return e.lastestTransactionList[0:len(e.lastestTransactionList)]
	}
	return e.lastestTransactionList[0:8]
}

func (e *BlockExplorer) blocks(start int, currHeight uint32) []blockInfos {
	length := 10
	aaData := []blockInfos{}

	for i, j := currHeight-uint32(start), 0; i > 0 && j < length; i, j = i-1, j+1 {
		b, err := e.Kernel.Block(i)
		if err != nil {
			continue
		}
		cd, err := e.Kernel.Provider().Data(i)
		if err != nil {
			continue
		}
		status := 1
		if b.Header.TimeoutCount > 0 {
			status = 2
		}

		tm := time.Unix(int64(cd.Header.Timestamp()/uint64(time.Second)), 0)

		aaData = append(aaData, blockInfos{
			BlockHeight: i,
			BlockHash:   cd.Header.Hash().String(),
			Time:        tm.Format("2006-01-02 15:04:05"),
			Status:      strconv.Itoa(status),
			Txs:         strconv.Itoa(len(b.Body.Transactions)),
		})
	}

	return aaData
}

func (e *BlockExplorer) paginationBlocks(r *http.Request) (result blockInfosCase) {
	param := r.URL.Query()
	startStr := param.Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		return
	}
	currHeight := e.Kernel.Provider().Height()

	result.ITotalRecords = int(currHeight)
	result.ITotalDisplayRecords = int(currHeight)

	result.AaData = e.blocks(start, currHeight)

	return
}

type txInfosCase struct {
	ITotalRecords        int       `json:"iTotalRecords"`
	ITotalDisplayRecords int       `json:"iTotalDisplayRecords"`
	SEcho                int       `json:"sEcho"`
	SColumns             string    `json:"sColumns"`
	AaData               []txInfos `json:"aaData"`
}

func (e *BlockExplorer) txs(start int, length int) []txInfos {
	max := start + length
	if max > len(e.lastestTransactionList) {
		max = len(e.lastestTransactionList)
	}

	return e.lastestTransactionList[start:max]
}

func (e *BlockExplorer) paginationTxs(r *http.Request) (result txInfosCase) {
	param := r.URL.Query()
	startStr := param.Get("start")

	start, err := strconv.Atoi(startStr)
	if err != nil {
		return
	}
	length := 10

	result.ITotalRecords = len(e.lastestTransactionList)
	result.ITotalDisplayRecords = len(e.lastestTransactionList)

	result.AaData = e.txs(start, length)

	return
}

type score struct {
	Rank     uint32
	Addr     string
	Score    string
	AllScore *ScoreCase
}

type allScore struct {
	Total       []score
	Gold        []score
	Population  []score
	Electricity []score
	CoinCount   []score
}

func (e *BlockExplorer) allScore(r *http.Request) (result []score) {
	param := r.URL.Query()
	sort := param.Get("sort")
	keyword := param.Get("keyword")

	e.db.View(func(txn *badger.Txn) error {
		st := getTypeFromString("Game" + sort)

		var prefix []byte
		if keyword != "" {
			item, err := txn.Get([]byte("GameId" + keyword))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return err
				}
			} else {
				value, err := item.ValueCopy(nil)
				if err != nil {
					return err
				}

				addr := common.Address{}
				copy(addr[:], value)

				gameScoreAddr := []byte(fmt.Sprintf("%v:Addr:%v", getType(st), addr.String()))
				item, err := txn.Get(gameScoreAddr)
				if err != nil {
					if err != badger.ErrKeyNotFound {
						return err
					}
				} else {
					var err error
					prefix, err = item.ValueCopy(nil)
					if err != nil {
						return err
					}
				}
			}
		}

		result, _ = getScore(txn, prefix, st, 20)
		return nil
	})

	return
}

func (e *BlockExplorer) totalScore(r *http.Request) (result *allScore) {
	result = &allScore{
		Total:       []score{},
		Gold:        []score{},
		Population:  []score{},
		Electricity: []score{},
		CoinCount:   []score{},
	}
	e.db.View(func(txn *badger.Txn) error {
		result.Total, _ = getScore(txn, nil, Level, 10)
		result.CoinCount, _ = getScore(txn, nil, CoinCount, 10)
		result.Gold, _ = getScore(txn, nil, Balance, 5)
		result.Population, _ = getScore(txn, nil, ManProvided, 5)
		result.Electricity, _ = getScore(txn, nil, PowerProvided, 5)
		return nil
	})

	return
}

func getScore(txn *badger.Txn, prefixKey []byte, sType ScoreType, limit uint32) ([]score, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	opts.Reverse = true
	it := txn.NewIterator(opts)
	defer it.Close()

	prefix := []byte(getType(sType) + ":Score:")
	var rank uint32
	if len(prefixKey) == 0 {
		prefixKey = append([]byte(getType(sType)+":Score:"), 0xFF)
	} else {
		opts2 := badger.DefaultIteratorOptions
		opts2.PrefetchValues = false
		opts2.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()
		startkey := append([]byte(getType(sType)+":Score:"), 0xFF)
		for it.Seek(startkey); !it.ValidForPrefix(prefixKey); it.Next() {
			rank++
		}
	}
	s := []score{}
	for it.Seek(prefixKey); it.ValidForPrefix(prefix); it.Next() {
		rank++
		item := it.Item()
		k := item.Key()
		err := item.Value(func(v []byte) error {
			buf := bytes.NewBuffer(v)
			sc := &ScoreCase{}
			sc.ReadFrom(buf)

			value := strings.TrimPrefix(string(k), getType(sType)+":Score:")
			num := value[:15]
			Addr := value[15:]
			s = append(s, score{
				Rank:     rank,
				Addr:     Addr,
				Score:    num,
				AllScore: sc,
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
		if rank >= limit {
			break
		}
	}

	return s, nil

}
