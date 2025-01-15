package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	imgFct "workspace/IMAGE"
)

const (
	PATH_CODE  byte = 1 // Code pour l'envoi du chemin
	FILE_CODE  byte = 2 // Code pour l'envoi de fichier
	END_CODE   byte = 3 // Code pour le message de fin
	ACK_CODE   byte = 4 // Code pour les acquittements
	ERROR_CODE byte = 5 // Code pour les erreurs
)

var fileCounter int = 1
var clientPath string = ""

func main() {
	var wg1 sync.WaitGroup

	wg1.Add(1)
	go Init_Tcp(&wg1)
	wg1.Wait()
}

func filtreImages() {
	var wg2 sync.WaitGroup

	var compteur int = compte_Images(clientPath + "/received")

	images := make([]image.Image, compteur)
	var buff string
	erreurs := make([]error, compteur)
	for i := 1; i <= compteur; i++ {
		buff = fmt.Sprintf(clientPath+"/received/%d.jpg", i)
		fmt.Println("buff num ", i, " = ", buff)
		images[i-1], _, erreurs[i-1] = imgFct.GetImageData(buff)
		if erreurs[i-1] != nil {
			fmt.Println("Failed to get image data:", erreurs[i])
			os.Exit(1)
		}
	}
	for i := 1; i < compteur; i++ {
		for j := (i + 1); j <= compteur; j++ {
			wg2.Add(1)
			go compare(images[i-1], images[j-1], i, j, &wg2)
		}
	}
	wg2.Wait()
}

func compare(im1 image.Image, im2 image.Image, i int, j int, wg *sync.WaitGroup) {
	defer wg.Done()
	var distance float64 = 0
	distance = imgFct.GetTotalDistance(im1, im2)
	fmt.Println("distance entre image", i, " et ", j, " est de ", distance)
}

func compte_Images(dossier string) int {
	fichiers, err := ioutil.ReadDir(dossier)
	if err != nil {
		os.Exit(1)
	}
	var compteur int = 0
	for _, fichier := range fichiers {
		if !fichier.IsDir() {
			compteur++
		}
	}
	return compteur
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
		go handleConnection(conn)
	}
}

func sendAck(conn net.Conn) error {
	response := make([]byte, 2)
	response[0] = ACK_CODE
	response[1] = 1 // 1 pour succès
	_, err := conn.Write(response)
	return err
}

func sendError(conn net.Conn, errMsg string) error {
	response := make([]byte, len(errMsg)+2)
	response[0] = ERROR_CODE
	response[1] = byte(len(errMsg))
	copy(response[2:], []byte(errMsg))
	_, err := conn.Write(response)
	return err
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	pathReceived := false

	for {
		codeBuffer := make([]byte, 1)
		_, err := conn.Read(codeBuffer)
		if err != nil {
			log.Printf("Erreur de lecture du code: %v", err)
			return
		}

		switch codeBuffer[0] {
		case PATH_CODE:
			sizeBuf := make([]byte, 4)
			_, err := conn.Read(sizeBuf)
			if err != nil {
				log.Printf("Erreur lecture taille chemin: %v", err)
				return
			}
			pathSize := binary.BigEndian.Uint32(sizeBuf)

			pathBuf := make([]byte, pathSize)
			_, err = conn.Read(pathBuf)
			if err != nil {
				log.Printf("Erreur lecture chemin: %v", err)
				return
			}

			clientPath = string(pathBuf)
			pathReceived = true
			fmt.Printf("Chemin reçu: %s\n", clientPath)

			if err := os.MkdirAll(clientPath+"/received", os.ModePerm); err != nil {
				log.Printf("Erreur création dossier: %v", err)
				sendError(conn, "Erreur création dossier")
				return
			}

			sendAck(conn)

		case FILE_CODE:
			if !pathReceived {
				sendError(conn, "Chemin non reçu")
				return
			}

			if err := handleIncomingFile(conn); err != nil {
				log.Printf("Erreur traitement fichier: %v", err)
				sendError(conn, fmt.Sprintf("Erreur fichier: %v", err))
				return
			}

		case END_CODE:
			sendAck(conn)
			fmt.Println("Message de fin reçu, lancement du filtrage...")
			fileCounter = 1
			filtreImages()
			return

		default:
			log.Printf("Code message inconnu: %d", codeBuffer[0])
			sendError(conn, "Code message inconnu")
			return
		}
	}
}
func handlePATH()

func handleIncomingFile(conn net.Conn) error {
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

	newName := fmt.Sprintf("%d.jpg", fileCounter)
	fileCounter++

	fmt.Printf("Réception du fichier: %s -> %s\n", fileName, newName)

	if err := sendAck(conn); err != nil {
		return fmt.Errorf("erreur envoi ack header: %v", err)
	}

	filePath := filepath.Join(clientPath+"/received", newName)
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

		length := binary.BigEndian.Uint32(dataBuffer[5:9])

		//segmentNumber := binary.BigEndian.Uint32(dataBuffer[1:5])
		//fmt.Printf("Fichier: %s, Segment: %d/%d\n", newName, segmentNumber+1, reps)

		if length > 1014 {
			return fmt.Errorf("longueur données invalide segment %d: %d", i, length)
		}

		if _, err := file.Write(dataBuffer[9 : 9+length]); err != nil {
			return fmt.Errorf("erreur écriture fichier segment %d: %v", i, err)
		}

		if err := sendAck(conn); err != nil {
			return fmt.Errorf("erreur envoi ack segment %d: %v", i, err)
		}
	}

	fmt.Printf("Fichier %s reçu avec succès\n", newName)
	return nil
}
