package main

import (
	"messaggio/internal/server"
)

//	@title			Messaggio
//	@version		1.0

// @host		62.109.17.207:8080

func main() {
	err := server.StartServer()
	if err != nil {
		panic(err)
	}
}
