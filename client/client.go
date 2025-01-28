package main

import (
	tcpHandle "client/handle"
	tcpFct "client/tcp"
	"log"
	"os"
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
	folder := "/home/lucaslhm/Téléchargements"
	clientPath = folder

	conn, err := tcpFct.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(clientPath+"/filtred", 0755); err != nil {
		log.Fatal("Erreur création dossier filtred:", err)
	}

	go tcpHandle.HandleConnection(conn, &fileCounter, clientPath)

	if err := tcpFct.SendPhoto(folder, conn); err != nil {
		log.Printf("Erreur envoi photos: %v", err)
		conn.Close()
		return
	}
}
