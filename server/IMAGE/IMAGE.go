package IMAGE

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"math"
	"os"
	"sync"
)

type RGB struct {
	R, G, B int
}

func GetImageData(path string) (image.Image, string, error) {
	fmt.Println("Récupération des données de ", path)
	imgFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Image introuvable")
		return nil, "", err
	}
	defer imgFile.Close()

	imageData, imageType, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Impossible de décoder l'image:", err)
		return nil, "", err
	}
	return imageData, imageType, nil
}

func distanceEuclidienneRGB(pixel1 RGB, pixel2 RGB) float64 {
	r1, g1, b1 := pixel1.R, pixel1.G, pixel1.B
	r2, g2, b2 := pixel2.R, pixel2.G, pixel2.B
	return math.Sqrt(math.Pow(float64(r1-r2), 2) + math.Pow(float64(g1-g2), 2) + math.Pow(float64(b1-b2), 2))
}

func rowProcessDistance(row1 []RGB, row2 []RGB) float64 {
	var distance float64
	for i := 0; i < len(row1); i++ {
		distance += distanceEuclidienneRGB(row1[i], row2[i])
	}
	return distance
}

func createRGBRow(imageData image.Image, y int) []RGB {
	widthmax := imageData.Bounds().Max.X
	row := make([]RGB, widthmax)
	for x := 0; x < widthmax; x++ {
		r, g, b, _ := imageData.At(x, y).RGBA()
		row[x] = RGB{int(r >> 8), int(g >> 8), int(b >> 8)}
	}
	return row
}

func GetTotalDistance(imageData1 image.Image, imageData2 image.Image) float64 {
	bounds1 := imageData1.Bounds()
	bounds2 := imageData2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		fmt.Println("Les images n'ont pas la même taille")
		return 1e9
	}

	// Création des matrices RGB
	matrix1 := make([][]RGB, bounds1.Dy())
	matrix2 := make([][]RGB, bounds2.Dy())
	var wg sync.WaitGroup

	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		wg.Add(1)
		go func(y int) {
			defer wg.Done()
			matrix1[y] = createRGBRow(imageData1, y)
			matrix2[y] = createRGBRow(imageData2, y)
		}(y)
	}
	wg.Wait()

	fmt.Println("Matrices remplies. Calcul de la distance totale...")

	totalDistance := 0.0
	for i := 0; i < len(matrix1); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			totalDistance += rowProcessDistance(matrix1[i], matrix2[i])
		}(i)
	}
	wg.Wait()

	return totalDistance
}
