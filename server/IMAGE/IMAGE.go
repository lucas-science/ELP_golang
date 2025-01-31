package IMAGE

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"sync"
	"gocv.io/x/gocv"
	"image/color"
	"math"
)

/*
type RGB struct {
	R, G, B int
}
*/

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

/*

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

*/

func euclidienne (desc1, desc2 gocv.Mat) float64 {
	if desc1.Cols() != desc2.Cols() || desc1.Rows() != desc2.Rows() {
		fmt.Println("Different descriptors length detected !")
		return math.Inf(1) // Distance infinie si incompatibles
	}
	
	var distance float64
	for i := 0; i < desc1.Rows(); i++ {
		for j := 0; j < desc1.Cols(); j++ {
			var diff float64
			diff = float64(desc1.GetFloatAt(i, j) - desc2.GetFloatAt(i, j))
			distance += diff * diff
		}
	}

	return math.Sqrt(distance)
}

func imgtogray (img image.Image) gocv.Mat {
	
	bounds := img.Bounds()
	mat := gocv.NewMatWithSize(bounds.Dy(), bounds.Dx(), gocv.MatTypeCV8U)
	
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			grayColor := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
			mat.SetUCharAt(y, x, grayColor.Y)
		}
	}
	return mat
}

func calcul (img gocv.Mat, sift gocv.SIFT, kpChan chan<- []gocv.KeyPoint, descChan chan<- gocv.Mat, wg7 *sync.WaitGroup) {
	
	defer wg7.Done()
	kp, desc := sift.DetectAndCompute(img, gocv.NewMat())
	kpChan <- kp
	descChan <- desc
}

func GetTotalDistance(imageData1 image.Image, imageData2 image.Image) float64 {
	
	img1 := imgtogray(imageData1)
	img2 := imgtogray(imageData2)
	defer img1.Close()
	defer img2.Close()
	
	sift := gocv.NewSIFT()
	defer sift.Close()
	
	kpChan1, descChan1 := make(chan []gocv.KeyPoint, 1), make(chan gocv.Mat, 1)
	kpChan2, descChan2 := make(chan []gocv.KeyPoint, 1), make(chan gocv.Mat, 1)
	
	var wg7 sync.WaitGroup
	wg7.Add(2)
	go func () {
		calcul(img1, sift, kpChan1, descChan1, &wg7)
		close(kpChan1)
		close(descChan1)
	}()
	go func () {
		calcul(img2, sift, kpChan2, descChan2, &wg7)
		close(kpChan2)
		close(descChan2)
	}()
	wg7.Wait()
	desc1 := <-descChan1
	desc2 := <-descChan2
	defer desc1.Close()
	defer desc2.Close()
	
	if desc1.Empty() || desc2.Empty() {
		fmt.Println("Error : Can't read descriptors'")
		os.Exit(1)
	}
	distancefinale := euclidienne(desc1.RowRange(0, 1), desc2.RowRange(0, 1))
	return distancefinale
}
