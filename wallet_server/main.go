//Запускает сервер кошелька, который позволяет управлять кошельком пользователя
//Сервер запускается на порту 8080 
//Порт можно изменить go run . -port 8081
//Шлюз можно изменить go run . -gateway http://127.0.0.1:5001

package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Wallet Server: ")
}

func main() {
	port := flag.Uint("port", 8080, "TCP Port Number for Wallet Server")
	gateway := flag.String("gateway", "http://127.0.0.1:5000", "Blockchain Gateway")
	flag.Parse()

	app := NewWalletServer(uint16(*port), *gateway)
	app.Run()
}
