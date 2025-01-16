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
	tcpFct "workspace/tcp"
)

const (
	PATH_CODE  byte = 1 // Code pour l'envoi du chemin
	FILE_CODE  byte = 2 // Code pour l'envoi de fichier
	END_CODE   byte = 3 // Code pour le message de fin
	ACK_CODE   byte = 4 // Code pour les acquittements
	ERROR_CODE byte = 5 // Code pour les erreurs
)

var fileCounter int = 1

func main() {
	var wg1 sync.WaitGroup

	wg1.Add(1)
	go Init_Tcp(&wg1)
	wg1.Wait()
}

func filtreImages() {
	var wg2 sync.WaitGroup

	var compteur int = compte_Images("./received")

	images := make([]image.Image, compteur)
	var buff string
	erreurs := make([]error, compteur)
	for i := 1; i <= compteur; i++ {
		buff = fmt.Sprintf("./received/%d.jpg", i)
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

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Créer le dossier received s'il n'existe pas
	if err := os.MkdirAll("./received", 0755); err != nil {
		log.Printf("Erreur création dossier received: %v", err)
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
			if err := handleIncomingFile(conn); err != nil {
				log.Printf("Erreur traitement fichier: %v", err)
				return
			}

		case END_CODE:
			fmt.Println("Message de fin reçu, lancement du filtrage...")
			fileCounter = 1

			// Exécuter le filtrage
			filtreImages()

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
