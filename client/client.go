package main

import (
	tcpFct "client/tcp"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

var fileCounter int = 1

//var clientPath string = ""

const (
	PATH_CODE  byte = 1 // Code pour l'envoi du
	FILE_CODE  byte = 2 // Code pour l'envoi de fichier
	END_CODE   byte = 3 // Code pour le message de fin
	ACK_CODE   byte = 4 // Code pour les acquittements
	ERROR_CODE byte = 5 // Code pour les erreurs
)

func showFolderOpen(parent fyne.Window) {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			fmt.Println("Erreur lors de la sélection du dossier :", err)
			return
		}
		if folder != nil {
			conn, err := tcpFct.CreateConnection()
			if err != nil {
				log.Fatal(err)
			}

			if err := os.MkdirAll("./filtred", 0755); err != nil {
				log.Fatal("Erreur création dossier filtred:", err)
			}

			go handleConnection(conn)

			if err := tcpFct.SendPhoto(folder.Path(), conn); err != nil {
				log.Printf("Erreur envoi photos: %v", err)
				conn.Close()
				return
			}

		} else {
			fmt.Println("Aucun dossier sélectionné.")
		}
	}, parent)
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("File system filter")

	myWindow.Resize(fyne.NewSize(1000, 600))

	hello := widget.NewLabel("Choisi le dossier à trier")
	myWindow.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Quitter", func() {
			myApp.Quit()
		}),
		widget.NewButton("Choisir un dossier", func() {
			showFolderOpen(myWindow)
		}),
	))
	myWindow.ShowAndRun()
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Créer le dossier filtred s'il n'existe pas
	if err := os.MkdirAll("./filtred", 0755); err != nil {
		log.Printf("Erreur création dossier filtred: %v", err)
		return
	}

	fileCounter = 1 // Réinitialiser le compteur pour les fichiers reçus

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
			if err := handleIncomingFile(conn); err != nil {
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

	filePath := filepath.Join("./filtred", newName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("erreur création fichier: %v", err)
	}
	defer file.Close()
	fmt.Println("Reception du fichier", filePath)
	dataBuffer := make([]byte, 1024)
	for i := 0; i < int(reps); i++ {
		n, err := conn.Read(dataBuffer)
		if err != nil {
			return fmt.Errorf("erreur lecture segment %d: %v", i, err)
		}
		if n != 1024 {
			return fmt.Errorf("taille segment %d invalide: %d", i, n)
		}
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
