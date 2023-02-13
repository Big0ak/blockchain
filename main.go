package main

import (
	"fmt"
	"log"

	"github.com/Big0ak/blockchain/block"
	"github.com/Big0ak/blockchain/wallet"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	walletM := wallet.NewWallet()
	walletA := wallet.NewWallet()
	walletB := wallet.NewWallet()

	// Кошелек
	t := wallet.NewTransaction(walletA.PrivateKey(), walletA.PublicKey(), walletA.BlockchainAdress(), walletB.BlockchainAdress(), 1.0)

	// блокчейн
	blockchain := block.NewBlockhain(walletM.BlockchainAdress())

	isAdded := blockchain.AddTransaction(walletA.BlockchainAdress(), walletB.BlockchainAdress(), 1.0,
		walletA.PublicKey(), t.GenerateSignature())
	fmt.Println(isAdded)

	blockchain.Mining()
	blockchain.Print()

	fmt.Printf("A %.1f\n", blockchain.CalculateTotalAmount(walletA.BlockchainAdress()))
	fmt.Printf("B %.1f\n", blockchain.CalculateTotalAmount(walletB.BlockchainAdress()))
	fmt.Printf("M %.1f\n", blockchain.CalculateTotalAmount(walletM.BlockchainAdress()))


	// myAdress := "my_Adress"
	// blockchain := NewBlockhain(myAdress)
	// blockchain.Print()

	// blockchain.AddTransaction("A", "B", 1.0)
	// blockchain.Mining()
	// blockchain.Print()

	// blockchain.AddTransaction("C", "D", 2.0)
	// blockchain.AddTransaction("X", "Y", 3.0)
	// blockchain.Mining()
	// blockchain.Print()

	// fmt.Printf("my_Adress %.1f\n", blockchain.CalculateTotalAmount(myAdress))
	// fmt.Printf("C %.1f\n", blockchain.CalculateTotalAmount("C"))
	// fmt.Printf("B %.1f\n", blockchain.CalculateTotalAmount("B"))
}
