package main

import (
	tcpHandle "client/handle"
	tcpFct "client/tcp"
	"log"
	"os"
	"sync"
)

var fileCounter int = 1

var clientPath string = ""

const (
	PATH_CODE  byte = 1
	FILE_CODE  byte = 2
	END_CODE   byte = 3
	ACK_CODE   byte = 4
	ERROR_CODE byte = 5
)

func main() {
	folder := "/home/lucaslhm/Documents/test_golang"
	clientPath = folder

	conn, err := tcpFct.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(clientPath+"/filtred", 0755); err != nil {
		log.Fatal("Erreur création dossier filtred:", err)
	}

	if err := tcpFct.SendPhoto(folder, conn); err != nil {
		log.Printf("Erreur envoi photos: %v", err)
		conn.Close()
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go tcpHandle.HandleConnection(conn, &fileCounter, clientPath, &wg)
	wg.Wait()
}
