## Docs importantes sur les différentes technos


#### Go routine + Wait groupe

```go
package main

import (
	"fmt"
	"sync"
)

func myFunction(wg *sync.WaitGroup) {
	defer wg.Done() // Décrémente le compteur du WaitGroup lorsque la fonction termine
	fmt.Println("Goroutine terminée")
}

func main() {
	var wg sync.WaitGroup

	// Lancement de 3 goroutines
	wg.Add(3)

	go myFunction(&wg)
	go myFunction(&wg)
	go myFunction(&wg)

	wg.Wait() // Attend que toutes les goroutines aient fini
	fmt.Println("Toutes les goroutines sont terminées")
}
```
* les go routines s'éxecutent en parallèle
* les "Waitgroupe" (wg), regroupe les routines => typiquement pour attendre que le groupe finisse


#### How define a function properly
```go
func getImageData(path string) (image.Image, string, error) {
    ....
    ....
    ....
	return imageData, imageType, err
}
```

#### Comment récupérer une varibale sans pour autant l'utiliser ? 

```go
imageData1, _, err1 := getImageData("./img/img1.jpg")
```
(l'underscore)