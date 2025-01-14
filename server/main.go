package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

var fileCounter int = 0

func main() {
	if err := os.MkdirAll("./data", 0755); err != nil {
		log.Fatal("Erreur création dossier received:", err)
	}
	Init_Tcp()
}

func Init_Tcp() {
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
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		if err := handleIncomingFile(conn); err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Client déconnecté")
				return
			}
			log.Printf("Erreur de traitement du fichier: %v", err)
			return
		}
	}
}

func handleIncomingFile(conn net.Conn) error {
	headerBuffer := make([]byte, 1024)

	n, err := conn.Read(headerBuffer)
	if err != nil {
		return err
	}
	if n != 1024 {
		return fmt.Errorf("taille header invalide: %d", n)
	}

	if headerBuffer[0] != byte(1) || headerBuffer[1023] != byte(0) {
		return fmt.Errorf("marqueurs header invalides")
	}

	reps := binary.BigEndian.Uint32(headerBuffer[1:5])
	lengthOfName := binary.BigEndian.Uint32(headerBuffer[5:9])
	fullPath := string(headerBuffer[9 : 9+lengthOfName])

	_, fileName := filepath.Split(fullPath)
	if fileName == "" {
		return fmt.Errorf("nom de fichier invalide")
	}

	newName := fmt.Sprintf("%d.jpg", fileCounter)
	fileCounter++

	fmt.Printf("Réception du fichier: %s\n", newName)

	if _, err := conn.Write([]byte("Header Received")); err != nil {
		return fmt.Errorf("erreur envoi confirmation header: %v", err)
	}

	filePath := filepath.Join("./received", newName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %v", err)
	}
	defer file.Close()

	dataBuffer := make([]byte, 1024)
	for i := 0; i < int(reps); i++ {
		n, err := conn.Read(dataBuffer)
		if err != nil {
			return fmt.Errorf("erreur lecture segment %d: %v", i, err)
		}
		if n != 1024 {
			return fmt.Errorf("taille segment %d invalide: %d", i, n)
		}

		segmentNumber := binary.BigEndian.Uint32(dataBuffer[1:5])
		length := binary.BigEndian.Uint32(dataBuffer[5:9])

		fmt.Printf("Fichier: %s, Segment: %d/%d\n", fileName, segmentNumber+1, reps)

		if length > 1014 {
			return fmt.Errorf("longueur données invalide segment %d: %d", i, length)
		}

		if _, err := file.Write(dataBuffer[9 : 9+length]); err != nil {
			return fmt.Errorf("erreur écriture fichier segment %d: %v", i, err)
		}

		if dataBuffer[0] != byte(0) || dataBuffer[1023] != byte(1) {
			return fmt.Errorf("marqueurs segment %d invalides", i)
		}

		if _, err := conn.Write([]byte("Segment Received")); err != nil {
			return fmt.Errorf("erreur envoi confirmation segment %d: %v", i, err)
		}
	}

	fmt.Printf("Fichier %s reçu avec succès\n", fileName)
	return nil
}
