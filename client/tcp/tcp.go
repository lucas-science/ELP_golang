package tcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
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

func CreateConnection() (net.Conn, error) {
	conn, err := net.Dial("tcp", HOST+":"+PORT)
	if err != nil {
		return nil, fmt.Errorf("connection error: %v", err)
	}
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	return conn, nil
}

func sendEndMessage(conn net.Conn) error {
	buffer := []byte{END_CODE}
	if _, err := conn.Write(buffer); err != nil {
		return fmt.Errorf("erreur envoi END: %v", err)
	}
	return nil
}

func sendHeader(conn net.Conn, header FileMetaData) error {
	// Ajouter le code de fichier avant le header
	codeBuffer := []byte{FILE_CODE}
	if _, err := conn.Write(codeBuffer); err != nil {
		return fmt.Errorf("erreur envoi code fichier: %v", err)
	}

	// Envoyer le header
	headerBuffer := make([]byte, 1024)
	headerBuffer[0] = 1
	binary.BigEndian.PutUint32(headerBuffer[1:5], header.reps)
	binary.BigEndian.PutUint32(headerBuffer[5:9], uint32(len(header.name)))
	copy(headerBuffer[9:9+len(header.name)], []byte(header.name))
	headerBuffer[1023] = 0

	_, err := conn.Write(headerBuffer)
	return err
}

func sendFileSegments(conn net.Conn, file *os.File, header FileMetaData) error {
	dataBuffer := make([]byte, 1014)
	fmt.Printf("Envoi du fichier %s (%d segments)...\n", header.name, header.reps)

	for i := 0; i < int(header.reps); i++ {
		fmt.Printf("Envoi segment %d/%d...\r", i+1, header.reps)
		if err := sendSegment(conn, file, dataBuffer, i); err != nil {
			fmt.Println()
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

	_, err = conn.Write(segmentBuffer)
	return err
}

func SendPhoto(folderPath string, conn net.Conn) error {
	f, err := os.Open(folderPath)
	if err != nil {
		return fmt.Errorf("erreur ouverture dossier: %v", err)
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		return fmt.Errorf("erreur lecture dossier: %v", err)
	}

	fmt.Println("Nombre de fichiers:", len(files))
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		fmt.Println(v.Name())

		file, err := os.Open(folderPath + "/" + v.Name())
		if err != nil {
			return fmt.Errorf("erreur ouverture fichier: %v", err)
		}

		header := setMetaDataFile(file)
		if err := sendHeader(conn, header); err != nil {
			file.Close()
			return fmt.Errorf("erreur envoi header: %v", err)
		}

		if err := sendFileSegments(conn, file, header); err != nil {
			file.Close()
			return fmt.Errorf("erreur envoi segments: %v", err)
		}
		file.Close()
	}

	return sendEndMessage(conn)
}
