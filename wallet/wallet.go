package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcutil/base58"

	"github.com/Big0ak/blockchain/utils"
)

type Wallet struct {
	privateKey       *ecdsa.PrivateKey
	publicKey        *ecdsa.PublicKey
	blockchainAdress string
}

func NewWallet() *Wallet {
	// Создание биткоин адреса
	// 1. Создание приватного ключа
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	// 2. Взять SHA-256 от публичного ключа
	h2 := sha256.New()
	h2.Write(privateKey.PublicKey.X.Bytes())
	h2.Write(privateKey.PublicKey.Y.Bytes())
	digest2 := h2.Sum(nil)

	// 3. RIPEMD-160 хэш (20 байт)
	h3 := sha256.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)

	// 4. Добавление в начале нулей 0x00 для того, чтобы это соответствовало основной сети
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])

	// 5. SHA-256 из 0x00 + RIPEMD-160
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)

	// 6. SHA-256 из предыдущего
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)

	// 7. Взять первые 4 байта из полученного хэша
	chsum := digest6[:4]

	// 8. Добавить чексумму в конец хэша
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])
	copy(dc8[21:], chsum[:])

	// 9. Конвертировать полученный хэш в base58
	address := base58.Encode(dc8)

	return &Wallet{
		privateKey:       privateKey,
		publicKey:        &privateKey.PublicKey,
		blockchainAdress: address,
	}
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.PrivateKey().D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%x%x", w.PublicKey().X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) BlockchainAdress() string {
	return w.blockchainAdress
}

// ----------------------------------------------------------------------------------------- //
// ---------------------------------- Transaction ------------------------------------------ //
// ----------------------------------------------------------------------------------------- //

type Transaction struct {
	senderPrivateKey          *ecdsa.PrivateKey
	senderPublicKey           *ecdsa.PublicKey
	senderBlockchainAddress   string
	recipientBlockchainAdress string
	value                     float32
}

func NewTransaction(privateKey *ecdsa.PrivateKey, publickKey *ecdsa.PublicKey,
	sender string, recipient string, value float32) *Transaction {
	return &Transaction{
		senderPrivateKey:          privateKey,
		senderPublicKey:           publickKey,
		senderBlockchainAddress:   sender,
		recipientBlockchainAdress: recipient,
		value:                     value,
	}
}

func (t *Transaction) GenerateSignature() *utils.Signature {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m)) // хэш от транзакции
	r, s, _ := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	return &utils.Signature{R: r, S: s}
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	// тут не нужна информация о закрытом и открытом ключе
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
