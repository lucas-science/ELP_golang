package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"math"
	"os"
)

struct RGB {
	R, G, B int
} 

func getImageData(path string) (image.Image, string, error) {
	fmt.Println("Récupération des données de ", path)
	imgFile, err := os.Open(path)

	if err != nil {
		fmt.Println("image not found")
		return nil, "", err
	}
	defer imgFile.Close()

	imageData, imageType, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Failed to decode the image:", err)
		return nil, "", err
	}
	return imageData, imageType, err
}

func distanceEuclidienneRGB(pixel1 Array, pixel2 Array) float64 {
	r1, g1, b1 := pixel1[0], pixel1[1], pixel1[2]
	r2, g2, b2 := pixel2[0], pixel2[1], pixel2[2]
	return math.Sqrt(math.Pow(r1-r2, 2) + math.Pow(g1-g2, 2) + math.Pow(b1-b2, 2))
}

func rowProcessDistance(row1, row2 []Array) float64 {
	var distance float64
	for i := 0; i < len(row1); i++ {
		distance += distanceEuclidienneRGB(row1[i], row2[i])
	}
	return distance
}

func addRGBRow(imageData image.Image, y int, matrice *[][]RGB) []Array {
	vidthmax := imageData.Bounds().Max.X
	row := make([]color.Color, widthmax)
	for x := 0; x < widthmax; x++ {
		row[x] = imageData.At(x, y)
	}
	return row
}

func GetTotalDistance(imageData1 image.Image, imageData2 image.Image) float64 {
	bounds1 := imageData1.Bounds()
	bounds2 := imageData2.Bounds()

	if bounds1.Dx() != bounds2.Dx() || bounds1.Dy() != bounds2.Dy() {
		fmt.Println("Les images n'ont pas la même taille")
		return 1000000000
	}
	var matrix [][]RGB

	var wg sync.WaitGroup
	wg.Add(bounds1.MAX.Y)
	for y := bounds1.Min.Y; y < bounds1.Max.Y; y++ {
		go addRGBRow(imageData1, y, &matrix)
	}
	wg.Wait() 
	fmt.Println("Go routines terminés. La matrice remplie")

	totalDistance := 0.0
	// on fait les distances
	return totalDistance
}

func main() {
	fmt.Println("Hello, world.")

	// On ignore imageType en utilisant l'underscore
	imageData1, _, err1 := getImageData("./img/img1.jpg")
	imageData2, _, err2 := getImageData("./img/img2.jpg")

	if err1 != nil || err2 != nil {
		if err1 != nil {
			fmt.Println("Failed to get image data:", err1)
		} else if err2 != nil {
			fmt.Println("Failed to get image data:", err2)
		} else {
			fmt.Println("Failed to get these images data")
		}
		os.Exit(1)
	}

	fmt.Println(imageData1.At(10, 20))
	fmt.Println(imageData2.At(10, 20))
}
