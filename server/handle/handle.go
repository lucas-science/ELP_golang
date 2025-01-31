package handle

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"workspace/filtre"
	tcpFct "workspace/tcp"
)

const (
	PATH_CODE  byte = 1
	FILE_CODE  byte = 2
	END_CODE   byte = 3
	ACK_CODE   byte = 4
	ERROR_CODE byte = 5
)

func HandleConnection(conn net.Conn, fileCounter *int) {
	defer conn.Close()
	fmt.Println(fileCounter)
	
	// supprime le dossier received
	if err := os.RemoveAll("./received"); err != nil {
		log.Printf("Failed to clean folder : %v", err)
		return
	}
	// Créer le dossier received s'il n'existe pas
	if err := os.MkdirAll("./received", 0755); err != nil {
		log.Printf("Failed to create folder : %v", err)
		return
	}

	for {
		codeBuffer := make([]byte, 1)
		_, err := conn.Read(codeBuffer)
		if err != nil {
			log.Printf("Erreur de lecture du code: %v", err)
			return
		}

		switch codeBuffer[0] {
		case FILE_CODE:
			if err := handleIncomingFile(conn, fileCounter); err != nil {
				log.Printf("Erreur traitement fichier: %v", err)
				return
			}

		case END_CODE:
			fmt.Println("Message de fin reçu, lancement du filtrage...")
			*fileCounter = 1

			// Exécuter le filtrage
			filtre.FiltreImages()

			// Renvoyer les images filtrées
			fmt.Println("Envoi des images filtrées au client...")
			if err := tcpFct.SendPhoto("./received", conn); err != nil {
				log.Printf("Erreur envoi images filtrées: %v", err)
			}
			return

		default:
			log.Printf("Code message inconnu: %d", codeBuffer[0])
			return
		}
	}
}
func handleIncomingFile(conn net.Conn, fileCounter *int) error {
	headerBuffer := make([]byte, 1024)
	_, err := conn.Read(headerBuffer)
	if err != nil {
		return fmt.Errorf("erreur lecture header: %v", err)
	}
	
	

	reps := binary.BigEndian.Uint32(headerBuffer[1:5])
	lengthOfName := binary.BigEndian.Uint32(headerBuffer[5:9])
	fullPath := string(headerBuffer[9 : 9+lengthOfName])

	_, fileName := filepath.Split(fullPath)
	if fileName == "" {
		return fmt.Errorf("nom de fichier invalide")
	}

	newName := fmt.Sprintf("%d.jpg", *fileCounter)
	*fileCounter++

	fmt.Printf("Réception du fichier: %s -> %s\n", fileName, newName)

	filePath := filepath.Join("./received", newName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %v", err)
	}
	defer file.Close()

	dataBuffer := make([]byte, 1024)
	for i := 0; i < int(reps); i++ {
		if err := readFull(conn, dataBuffer, 1024); err != nil {
			return fmt.Errorf("erreur lecture segment %d: %v", i, err)
		}

		length := binary.BigEndian.Uint32(dataBuffer[5:9])

		// debug info
		//segmentNumber := binary.BigEndian.Uint32(dataBuffer[1:5])
		//fmt.Printf("Fichier: %s, Segment: %d/%d\n", newName, segmentNumber+1, reps)

		if length > 1014 {
			return fmt.Errorf("longueur données invalide segment %d: %d", i, length)
		}

		if _, err := file.Write(dataBuffer[9 : 9+length]); err != nil {
			return fmt.Errorf("erreur écriture fichier segment %d: %v", i, err)
		}
	}

	fmt.Printf("Fichier %s reçu avec succès\n", newName)
	return nil
}

func readFull(conn net.Conn, buffer []byte, size int) error {
	totalRead := 0
	for totalRead < size {
		n, err := conn.Read(buffer[totalRead:])
		if err != nil {
			return fmt.Errorf("erreur lecture segment (lu %d/%d octets): %v", totalRead, size, err)
		}
		totalRead += n
	}
	return nil
}
