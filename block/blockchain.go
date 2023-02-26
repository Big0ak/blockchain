package block

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Big0ak/blockchain/utils"
)

const (
	MINING_DIFFICULTY = 3 // Количество нулей в хэш функции
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWARD     = 1.0
)

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Block ------------------------------------------------ //
// ----------------------------------------------------------------------------------------- //

type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	return &Block{
		timestamp:    time.Now().UnixNano(),
		nonce:        nonce,
		previousHash: previousHash,
		transactions: transactions,
	}
}

func (b *Block) Print() {
	fmt.Printf("timestamp 		%d\n", b.timestamp)
	fmt.Printf("nonce 		 	%d\n", b.nonce)
	fmt.Printf("previous_hash 	%x\n", b.previousHash)
	for _, t := range b.transactions {
		t.Print()
	}
}

// Получение хэш блока
func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

// Custom marshal for Block
func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Blockchain ------------------------------------------- //
// ----------------------------------------------------------------------------------------- //

type Blockchain struct {
	transactionPool  []*Transaction
	chain            []*Block
	blockchainAdress string
	port             uint16
}

func NewBlockhain(blockchainAdress string, port uint16) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.CreatBlock(0, b.Hash())
	bc.blockchainAdress = blockchainAdress
	bc.port = port
	return bc
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chains"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) CreatBlock(nonce int, previousHash [32]byte) *Block {
	// Все транзакции из пула записываем в блок
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	// Удаление всех транзакций из пула
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("=", 59))
}

// Добавление транзакций в пул
func (bc *Blockchain) AddTransaction(sender, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	// Если это транзакция вознаграждение майнера
	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		// if bc.CalculateTotalAmount(sender) < value {
		// 	log.Print("ERROR: Not enough balance in a wallet")
		// 	return false
		// }
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: Verify Transaction")
	}
	return false
}

// Проверка транзакции на подпись
func (bc *Blockchain) VerifyTransactionSignature(
	senderPublickKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublickKey, h[:], s.R, s.S)
}

// Копировние пула транзакций
func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transaction := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transaction = append(transaction,
			NewTransaction(t.senderBlockchainAddress, t.recipientBlockchainAdress, t.value))
	}
	return transaction
}

// Проверка блока.
// Необходимые данные для проверки: nonce, хэш предыдущей транзакции, все транзакции для этого блока
func (bc *Blockchain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	zeros := strings.Repeat("0", difficulty) // возможно сделать константой
	guessBlock := Block{0, nonce, previousHash, transactions}
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	return guessHashStr[:difficulty] == zeros
}

// Поиск nonce
func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool()
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for ; !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY); nonce++ {
	}
	return nonce
}

func (bc *Blockchain) Mining() bool {
	// транзакция о вознаграждении майнера
	bc.AddTransaction(MINING_SENDER, bc.blockchainAdress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreatBlock(nonce, previousHash)
	log.Println("action=mining, status=success")
	return true
}

func (bc *Blockchain) CalculateTotalAmount(blockchainAdress string) float32 {
	var totalAmount float32 = 0.0
	for _, b := range bc.chain {
		for _, t := range b.transactions {
			val := t.value
			if blockchainAdress == t.recipientBlockchainAdress {
				totalAmount += val
			}

			if blockchainAdress == t.senderBlockchainAddress {
				totalAmount -= val
			}
		}
	}
	return totalAmount
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Transaction ------------------------------------------ //
// ----------------------------------------------------------------------------------------- //

// транзакции сначала отображаюься в пуле, потом попадают в блок при его создании
type Transaction struct {
	senderBlockchainAddress   string
	recipientBlockchainAdress string
	value                     float32
}

func NewTransaction(sender, recipient string, value float32) *Transaction {
	return &Transaction{
		senderBlockchainAddress:   sender,
		recipientBlockchainAdress: recipient,
		value:                     value,
	}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender      %s\n", t.senderBlockchainAddress)
	fmt.Printf(" recipient   %s\n", t.recipientBlockchainAdress)
	fmt.Printf(" value   %.1f\n", t.value)
}

// Custom marshal for Transaction
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_adress"`
		Recipient string  `json:"recipient_adress"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAdress,
		Value:     t.value,
	})
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Transaction Request ---------------------------------- //
// ----------------------------------------------------------------------------------------- //

// транзакция отправленная от backend wallet to blockchain server
type TransactionRequest struct {
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublickKey           *string `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.Signature == nil || tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil || tr.SenderPublickKey == nil ||
		tr.Value == nil {
		return false
	}
	return true
}
