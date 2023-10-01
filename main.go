package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

const (
	rpcURL    = "https://eth.llamarpc.com"
	outputDir = "./output/"
)

var (
	concurrentLimit      int
	startBlock, endBlock int64
)

type TransactionEntry struct {
	TransactionHash string `json:"hash"`
	From            string `json:"from"`
	Input           string `json:"input"`
	Nonce           string `json:"nonce"`
}

func init() {
	flag.IntVar(&concurrentLimit, "routines", 5, "set go routines for processing.")
	flag.Int64Var(&startBlock, "start", 12000000, "set starting block.")
	flag.Int64Var(&endBlock, "end", 14000000, "set ending block.")
	flag.Parse()
}

func main() {
	client, err := rpc.Dial(rpcURL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	sBlock := big.NewInt(startBlock)
	eBlock := big.NewInt(endBlock)

	blockCh := make(chan *big.Int)
	dataCh := make(chan []TransactionEntry)
	errorCh := make(chan error)
	semaphore := make(chan struct{}, concurrentLimit)

	for i := 0; i < concurrentLimit; i++ {
		go blockWorker(client, blockCh, dataCh, errorCh, semaphore)
	}

	fmt.Printf("startBlock: %s, endBlock: %s\n", sBlock.String(), eBlock.String())

	go func() {
		for i := new(big.Int).Set(sBlock); i.Cmp(eBlock) <= 0; i.Add(i, big.NewInt(1)) {
			blockCh <- i
		}
		close(blockCh)
	}()
	for i := new(big.Int).Set(sBlock); i.Cmp(eBlock) <= 0; i.Add(i, big.NewInt(1)) {
		if _, err := os.Stat(outputDir); os.IsNotExist(err) {
			os.Mkdir(outputDir, os.ModePerm)
		}
		select {
		case data := <-dataCh:
			blockNum := i.String()
			transactionsJSON, err := json.MarshalIndent(data, "", "  ")
			if err != nil {
				log.Fatalf("Failed to marshal transactions into JSON: %v", err)
			}

			fileName := fmt.Sprintf("%s/tx_block_%s.json", outputDir, blockNum)
			file, err := os.Create(fileName)
			if err != nil {
				log.Fatalf("Failed to create JSON file for block %s: %v", blockNum, err)
			}

			_, err = file.Write(transactionsJSON)
			file.Close()
			if err != nil {
				log.Fatalf("Failed to write JSON to file for block %s: %v", blockNum, err)
			}

			fmt.Printf("Transactions for block %s saved to %s\n", blockNum, fileName)
		case err := <-errorCh:
			log.Fatalf("Error processing block: %v", err)
		}
	}
}

func getBlock(client *rpc.Client, blockNumber *big.Int) ([]TransactionEntry, error) {
	// fmt.Printf("Now getting all transactions from block number %s\n", blockNumber.String())
	var block map[string]interface{}
	err := client.Call(&block, "eth_getBlockByNumber", blockNumber, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get block: %v", err)
	}

	txs := block["transactions"].([]interface{})
	var transactionsArray []TransactionEntry
	for _, tx := range txs {
		txHash := tx.(map[string]interface{})["hash"].(string)
		transaction, err := getTransaction(client, txHash)
		if err != nil {
			return nil, err
		}

		if len(transaction.TransactionHash) > 0 {
			transactionsArray = append(transactionsArray, transaction)
		}
	}
	return transactionsArray, nil
}

func getTransaction(client *rpc.Client, txHash string) (TransactionEntry, error) {
	fmt.Printf("Now getting details from transaction ID %s\n", txHash)
	var tx map[string]interface{}
	err := client.Call(&tx, "eth_getTransactionByHash", txHash)
	if err != nil {
		return TransactionEntry{}, fmt.Errorf("failed to get transaction: %v", err)
	}

	txValue := new(big.Int)
	txValue.SetString(tx["value"].(string)[2:], 16)
	if txValue.Cmp(big.NewInt(0)) > 0 {
		transaction := TransactionEntry{
			TransactionHash: txHash,
			From:            tx["from"].(string),
			Input:           tx["input"].(string),
			Nonce:           tx["nonce"].(string),
		}
		return transaction, nil
	}
	return TransactionEntry{}, nil
}

func blockWorker(client *rpc.Client, blockCh chan *big.Int, dataCh chan []TransactionEntry, errorCh chan error, semaphore chan struct{}) {
	for block := range blockCh {
		semaphore <- struct{}{}
		txs, err := getBlock(client, block)
		if err != nil {
			errorCh <- err
		} else {
			go txWorker(client, txs, dataCh, errorCh)
		}
		<-semaphore
	}
}

func txWorker(client *rpc.Client, txs []TransactionEntry, dataCh chan []TransactionEntry, errorCh chan error) {
	var result []TransactionEntry
	for _, txHash := range txs {
		tx, err := getTransaction(client, txHash.TransactionHash)
		if err != nil {
			errorCh <- err
			return
		}
		result = append(result, tx)
	}
	dataCh <- result
}
