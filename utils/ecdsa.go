package utils

import (
	"fmt"
	"math/big"
)

type Signature struct {
	R *big.Int // координата X открытых ключей
	S *big.Int // вычисляется как хэш транзакции и открытый ключ
}

func (s *Signature) String() string {
	return fmt.Sprintf("%x%x", s.R, s.S)
}
