package main

import (
	tcpHandle "client/handle"
	tcpFct "client/tcp"
	"fmt"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
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

func showFolderOpen(parent fyne.Window) {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			fmt.Println("Erreur lors de la sélection du dossier :", err)
			return
		}
		if folder != nil {
			clientPath = folder.Path()

			conn, err := tcpFct.CreateConnection()
			if err != nil {
				log.Fatal(err)
			}

			if err := os.MkdirAll(clientPath+"/filtred", 0755); err != nil {
				log.Fatal("Erreur création dossier filtred:", err)
			}

			go tcpHandle.HandleConnection(conn, fileCounter, clientPath)

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
