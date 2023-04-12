package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Big0ak/blockchain/utils"
)

const (
	MINING_DIFFICULTY = 3 // Количество нулей в хэш функции
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWARD     = 1.0
	MINING_TIMER_SEC  = 20 // 20 сек. майнится новый блок

	BLOCKCHAIN_PORT_RANGE_START       = 5000
	BLOCKCHAIN_PORT_RANGE_END         = 5003
	NEIGHBOR_IP_RANGE_START           = 0
	NEIGHBOR_IP_RANGE_END             = 1
	BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC = 20 // проверять кадые 20 сек. подлючение blockchain Node
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

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Transaction() []*Transaction {
	return b.transactions
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

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	v := &struct {
		Timestamp *int64 `json:"timestamp"`
		Nonce *int `json:"nonce"`
		PreviousHash *string `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
	} {
		Timestamp: &b.timestamp,
		Nonce: &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	return nil
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Blockchain ------------------------------------------- //
// ----------------------------------------------------------------------------------------- //

type Blockchain struct {
	transactionPool  []*Transaction
	chain            []*Block
	blockchainAdress string
	port             uint16
	mux              sync.Mutex

	neighbors    []string
	muxNeighbors sync.Mutex
}

func NewBlockhain(blockchainAdress string, port uint16) *Blockchain {
	b := &Block{}
	bc := new(Blockchain)
	bc.CreatBlock(0, b.Hash())
	bc.blockchainAdress = blockchainAdress
	bc.port = port
	return bc
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

// ------------ Работа с Node --------------------
func (bc *Blockchain) Run() {
	bc.StartSyncNeighbors()
	// Перед запуском нового узла на него загружается блокчейн, который сейчас используется
	bc.ResolveConflicts()
	// Автоматический старт майнинга
	bc.StartMining()
}

// Поиск подлюченных Node
func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(
		"127.0.0.1", NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		bc.port, BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)
	log.Printf("%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

// Проверка каждые 20 секунд какие Node подключены
func (bc * Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second * BLOCKCHAIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}
// ----------------------------------------------

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

// очистка пула транзакций, в слуае... (не знаю как правильно написать)
func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Block *[]*Block `json:"chain"`
	} {
		Block: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) CreatBlock(nonce int, previousHash [32]byte) *Block {
	// Все транзакции из пула записываем в блок
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	// Удаление всех транзакций из пула
	bc.transactionPool = []*Transaction{}
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
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

func (bc *Blockchain) CreatTransaction(sender, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)

	// (синхронизация) обработка транзакций другими серверами
	if isTransacted {
		// отправляем подтвержденную транзакцию другим узлам
		for _, n := range bc.neighbors {
			publickKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(), senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &recipient, &publickKeyStr, &value, &signatureStr }
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("%v", resp)
		}
	}
	return isTransacted
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
		// Если недостаточный баланс, то не можем отправить
		if bc.CalculateTotalAmount(sender) < value {
			log.Print("ERROR: Not enough balance in a wallet")
			return false
		}
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
	bc.mux.Lock() // Блокируется
	// Предполагается, что майнинг закончится черзе 20 сек
	// Но если не так, то переходим к следующему майнингу блока потому что есть Lock()

	defer bc.mux.Unlock()

	// (Для тестовой сети) Если нет транзакций в блоке, то не запускать майнинг
	// if len(bc.transactionPool) == 0 {
	// 	return false
	// }

	// транзакция о вознаграждении майнера
	bc.AddTransaction(MINING_SENDER, bc.blockchainAdress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreatBlock(nonce, previousHash)
	log.Println("action=mining, status=success")

	// После того как нашли Nonce у блока, рассылаем всем узлам новый блок
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	return true
}

// Выполняется первая (Для майнинга)
func (bc *Blockchain) StartMining() {
	bc.Mining()
	// Если майнинг закончится быстрее чем 20 секунд,
	// то через 20 секунд будет запущен новый майнинг блока с помощью этой функции
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
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

// проверка корректности цепочки блоков
func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		// Проверка совпадания хэша предыдущего блока с получением хэша от самого предыдущего блока 
		if b.previousHash != preBlock.Hash() {
			return false
		}
		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transaction(), MINING_DIFFICULTY) {
			return false
		}
		
		preBlock = b
		currentIndex += 1
	}

	return true
}

// Если есть две цеппочки блоков, то выберается самая длинная
// Таким образом разрешается конфликт, когда появляется две разные копии блокчейна
func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLengh := len(bc.chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, _ := http.Get(endpoint)
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			_ =  decoder.Decode(&bcResp)

			chain := bcResp.Chain()

			if len(chain) > maxLengh && bc.ValidChain(chain) {
				maxLengh = len(chain)
				longestChain = chain
			}
		}
	}

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle conflicts replaced")
		return true
	}
	log.Printf("Resolve conflicts not replaced")
	return false
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

func (t *Transaction) UnmarshalJSON(data []byte) error {
	v := &struct {
		Sender    *string  `json:"sender_adress"`
		Recipient *string  `json:"recipient_adress"`
		Value     *float32 `json:"value"`
	}{
		Sender:    &t.senderBlockchainAddress,
		Recipient: &t.recipientBlockchainAdress,
		Value:     &t.value,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil 
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Transaction Request ---------------------------------- //
// ----------------------------------------------------------------------------------------- //

// транзакция отправленная от backend wallet to blockchain server
type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublickKey           *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string  `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.Signature == nil || tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil || tr.SenderPublickKey == nil ||
		tr.Value == nil {
		return false
	}
	return true
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Amount Response -------------------------------------- //
// ----------------------------------------------------------------------------------------- //

// Для получение баланса кошелька
type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}
