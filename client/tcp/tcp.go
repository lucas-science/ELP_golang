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

type FileMetaData struct {
	name     string
	fileSize uint32
	reps     uint32
}

func GetMetaDataFile(file *os.File) FileMetaData {
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

	// Définir un timeout pour les opérations de lecture/écriture
	conn.SetDeadline(time.Now().Add(30 * time.Second))
	return conn, nil
}

func SendHeader(conn net.Conn, header FileMetaData) error {
	headerBuffer := make([]byte, 1024)
	headerBuffer[0] = 1
	binary.BigEndian.PutUint32(headerBuffer[1:5], header.reps)
	binary.BigEndian.PutUint32(headerBuffer[5:9], uint32(len(header.name)))
	copy(headerBuffer[9:9+len(header.name)], []byte(header.name))
	headerBuffer[1023] = 0

	_, err := conn.Write(headerBuffer)
	if err != nil {
		return fmt.Errorf("erreur d'envoi du header: %v", err)
	}

	return waitForResponse(conn)
}

func SendFileSegments(conn net.Conn, file *os.File, header FileMetaData) error {
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

	if _, err := conn.Write(segmentBuffer); err != nil {
		return fmt.Errorf("erreur écriture segment: %v", err)
	}

	return waitForResponse(conn)
}

func waitForResponse(conn net.Conn) error {
	response := make([]byte, 16)
	_, err := conn.Read(response)
	if err != nil {
		return fmt.Errorf("erreur attente réponse: %v", err)
	}
	return nil
}
