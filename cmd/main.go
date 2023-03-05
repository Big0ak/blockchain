// Для поиска подключенных blockchain Node

package main

import (
	"fmt"

	"github.com/Big0ak/blockchain/utils"
)

func main() {
	// Поиск подлюченых адресов в заданных диапозонах
	//fmt.Println((utils.FindNeighbors("127.0.0.1", 0, 3, 5000, 5000, 5003)))

	fmt.Println(utils.GetHost())
}