package main

import (
	"fmt"

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
			fmt.Println("Dossier sélectionné :", folder.Path())
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

	// Lancer l'application
	myWindow.ShowAndRun()
}
