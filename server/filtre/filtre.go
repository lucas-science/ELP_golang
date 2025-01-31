package filtre

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"sync"
	imgFct "workspace/IMAGE"
)

func FiltreImages() {
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
			fmt.Printf("lancement des goroutines\n")
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
