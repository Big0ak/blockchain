//Запускает сервер блокчейна, который будет проверять транзакции пользователей и записывать их в блокчейн
//Сервер запускается на порту 5000 
//Порт можно изменить командой go run . -port 5001

package main

import (
	"flag"
	"log"
)

func init() {
	log.SetPrefix("Blockchain: ")
}

func main() {
	port := flag.Uint("port", 5000, "TCP Port Number for Blockchain Server")
	flag.Parse()
	app := NewBlockhainServer(uint16(*port))
	app.Run()    
}