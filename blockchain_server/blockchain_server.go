package main

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/Big0ak/blockchain/block"
	"github.com/Big0ak/blockchain/wallet"
)

var cache map[string]*block.Blockchain = make(map[string]*block.Blockchain)

type BlockchainServer struct {
	port uint16
}

func NewBlockhainServer(port uint16) *BlockchainServer {
	return &BlockchainServer{
		port: port,
	}
}

// Получение информации о блокчейне
func (bsc *BlockchainServer) GetBlockchain() *block.Blockchain {
	bc, ok := cache["blockchain"]
	if !ok {
		minerWallet := wallet.NewWallet()
		bc = block.NewBlockhain(minerWallet.BlockchainAddress(), bsc.Port())
		cache["blockchain"] = bc
		log.Printf("private_key %v", minerWallet.PrivateKeyStr())
		log.Printf("publick_key %v", minerWallet.PublicKeyStr())
		log.Printf("blockchain_address %v", minerWallet.BlockchainAddress())
	}
	return bc
}

func (bcs *BlockchainServer) Port() uint16 {
	return bcs.port
}

func (bcs *BlockchainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockchain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))

	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}

}

func (bcs *BlockchainServer) Run() {
	http.HandleFunc("/", bcs.GetChain)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bcs.Port())), nil))
}