package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	tcpHandle "workspace/handle"
)

const (
	PATH_CODE  byte = 1
	FILE_CODE  byte = 2
	END_CODE   byte = 3
	ACK_CODE   byte = 4
	ERROR_CODE byte = 5
)

var fileCounter int = 1

func main() {
	var wg1 sync.WaitGroup

	wg1.Add(1)
	go Init_Tcp(&wg1)
	wg1.Wait()
}

func Init_Tcp(wg1 *sync.WaitGroup) {
	defer wg1.Done()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ln.Close()
	for {
		fmt.Println("En attente de connexion...")
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Erreur d'acceptation de connexion: %v", err)
			continue
		}
		go tcpHandle.HandleConnection(conn, &fileCounter)
	}
}
