package handle

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

const (
	PATH_CODE  byte = 1
	FILE_CODE  byte = 2
	END_CODE   byte = 3
	ACK_CODE   byte = 4
	ERROR_CODE byte = 5
)

func HandleConnection(conn net.Conn, fileCounter *int, clientPath string, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()

	// Créer le dossier filtred s'il n'existe pas
	if err := os.MkdirAll(clientPath+"/filtred", 0755); err != nil {
		log.Printf("Erreur création dossier filtred: %v", err)
		return
	}

	for {
		codeBuffer := make([]byte, 1)
		_, err := conn.Read(codeBuffer)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Connexion terminée par le serveur")
				return
			}
			log.Printf("Erreur de lecture du code: %v", err)
			return
		}

		switch codeBuffer[0] {
		case FILE_CODE:
			if err := handleIncomingFile(conn, fileCounter, clientPath); err != nil {
				log.Printf("Erreur réception fichier: %v", err)
				return
			}

		case END_CODE:
			fmt.Println("Réception des fichiers terminée")
			return

		default:
			log.Printf("Code message inconnu: %d", codeBuffer[0])
			return
		}
	}
}
func handleIncomingFile(conn net.Conn, fileCounter *int, clientPath string) error {
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

	filePath := filepath.Join(clientPath+"/filtred", newName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %v", err)
	}
	defer file.Close()
	fmt.Println("Reception du fichier", filePath)
	dataBuffer := make([]byte, 1024)
	for i := 0; i < int(reps); i++ {
		if err := readFull(conn, dataBuffer, 1024); err != nil {
			return fmt.Errorf("erreur lecture segment %d: %v", i, err)
		}
		// debug info
		//segmentNumber := binary.BigEndian.Uint32(dataBuffer[1:5])
		//fmt.Printf("Fichier: %s, Segment: %d/%d\n", newName, segmentNumber+1, reps)

		length := binary.BigEndian.Uint32(dataBuffer[5:9])
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
