package main

import (
	"fmt"
	"os"
	imgFct "workspace/IMAGE"
)

func main() {
	fmt.Println("Hello, world.")

	imageData1, _, err1 := imgFct.GetImageData("./data/img1.jpg")
	imageData2, _, err2 := imgFct.GetImageData("./data/img3.jpg")

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
	totalDistance := imgFct.GetTotalDistance(imageData1, imageData2)

	fmt.Println("Total distance:", totalDistance)
}
