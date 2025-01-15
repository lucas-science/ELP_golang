package main

import (
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

// Fonction pour afficher le sélecteur de dossier
func showFolderOpen(parent fyne.Window) {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			fmt.Println("Erreur lors de la sélection du dossier :", err)
			return
		}
		if folder != nil {
			sendPhoto(folder.Path())
		} else {
			fmt.Println("Aucun dossier sélectionné.")
		}
	}, parent)
}

func sendPhoto(folderPath string) any {
	conn, err := tcpFct.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

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
	if err := tcpFct.SendPath(conn, folderPath); err != nil {
		log.Fatal("Erreur lors de l'envoi du chemin:", err)
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

		header := tcpFct.GetMetaDataFile(file)

		if err := tcpFct.SendHeader(conn, header); err != nil {
			fmt.Println("Erreur lors de l'envoi du header")
			log.Fatal(err)
		}

		if err := tcpFct.SendFileSegments(conn, file, header); err != nil {
			fmt.Println("Erreur lors de l'envoi du fichier")
			log.Fatal(err)
		}
	}
	tcpFct.SendEndMessage(conn)
	return nil
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
