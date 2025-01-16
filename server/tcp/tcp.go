package tcp

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	HOST = "localhost"
	PORT = "8080"
	TYPE = "tcp"
)

const (
	PATH_CODE  byte = 1
	FILE_CODE  byte = 2
	END_CODE   byte = 3
	ACK_CODE   byte = 4
	ERROR_CODE byte = 5
)

type FileMetaData struct {
	name     string
	fileSize uint32
	reps     uint32
}

func setMetaDataFile(file *os.File) FileMetaData {
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return FileMetaData{}
	}

	size := fileInfo.Size()
	return FileMetaData{
		name:     file.Name(),
		fileSize: uint32(size),
		reps:     uint32(size/1014) + 1,
	}
}

func sendEndMessage(conn net.Conn) error {
	buffer := []byte{END_CODE}
	if _, err := conn.Write(buffer); err != nil {
		return fmt.Errorf("erreur envoi END: %v", err)
	}
	return waitForAck(conn)
}

func sendHeader(conn net.Conn, header FileMetaData) error {
	// Ajouter le code de fichier avant le header
	codeBuffer := []byte{FILE_CODE}
	if _, err := conn.Write(codeBuffer); err != nil {
		return fmt.Errorf("erreur envoi code fichier: %v", err)
	}

	// Envoyer le header comme avant
	headerBuffer := make([]byte, 1024)
	headerBuffer[0] = 1
	binary.BigEndian.PutUint32(headerBuffer[1:5], header.reps)
	binary.BigEndian.PutUint32(headerBuffer[5:9], uint32(len(header.name)))
	copy(headerBuffer[9:9+len(header.name)], []byte(header.name))
	headerBuffer[1023] = 0

	if _, err := conn.Write(headerBuffer); err != nil {
		return fmt.Errorf("erreur envoi header: %v", err)
	}

	return nil
}

func waitForAck(conn net.Conn) error {
	response := make([]byte, 2)
	_, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("erreur lecture réponse: %v", err)
	}

	if response[0] == ERROR_CODE {
		// Lire le message d'erreur
		msgLen := response[1]
		errMsg := make([]byte, msgLen)
		_, err := conn.Read(errMsg)
		if err != nil {
			return fmt.Errorf("erreur lecture message erreur: %v", err)
		}
		return fmt.Errorf("erreur serveur: %s", string(errMsg))
	}

	if response[0] != ACK_CODE || response[1] != 1 {
		return fmt.Errorf("réponse invalide du serveur")
	}

	return nil
}

func sendFileSegments(conn net.Conn, file *os.File, header FileMetaData) error {
	dataBuffer := make([]byte, 1014)
	fmt.Printf("Envoi du fichier %s (%d segments)...\n", header.name, header.reps)

	for i := 0; i < int(header.reps); i++ {
		fmt.Printf("Envoi segment %d/%d...\r", i+1, header.reps)
		if err := sendSegment(conn, file, dataBuffer, i); err != nil {
			fmt.Println() // Pour la nouvelle ligne après le \r
			return fmt.Errorf("erreur segment %d: %v", i, err)
		}
	}
	fmt.Println("\nFichier envoyé avec succès!")
	return nil
}

func sendSegment(conn net.Conn, file *os.File, dataBuffer []byte, segmentNum int) error {
	segmentBuffer := make([]byte, 1024)

	n, err := file.ReadAt(dataBuffer, int64(segmentNum*1014))
	if err != nil && err.Error() != "EOF" {
		return fmt.Errorf("erreur lecture fichier: %v", err)
	}

	segmentBuffer[0] = 0
	binary.BigEndian.PutUint32(segmentBuffer[1:5], uint32(segmentNum))
	binary.BigEndian.PutUint32(segmentBuffer[5:9], uint32(n))
	copy(segmentBuffer[9:9+n], dataBuffer[:n])
	segmentBuffer[1023] = 1
	//fmt.Printf("Envoi segment, taille: %d octets\n", len(segmentBuffer))

	if _, err := conn.Write(segmentBuffer); err != nil {
		return fmt.Errorf("erreur écriture segment: %v", err)
	}
	return nil
}

func SendPhoto(folderPath string, conn net.Conn) any {
	f, err := os.Open(folderPath)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	files, err := f.Readdir(0)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	fmt.Println("Nombre de fichiers:", len(files))
	for _, v := range files {
		fmt.Println(v.IsDir(), v.Name())
	}
	for i, v := range files {
		if v.IsDir() {
			continue
		}
		fmt.Println(i, v.Name())

		file, err := os.Open(folderPath + "/" + v.Name())
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		header := setMetaDataFile(file)
		fmt.Println("Envoi du fichier:", header.name)
		if err := sendHeader(conn, header); err != nil {
			fmt.Println("Erreur lors de l'envoi du header")
			log.Fatal(err)
		}

		fmt.Println("Envoi des segments...")

		if err := sendFileSegments(conn, file, header); err != nil {
			fmt.Println("Erreur lors de l'envoi du fichier")
			log.Fatal(err)
		}
	}
	sendEndMessage(conn)
	return nil
}
