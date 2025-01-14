package main

import (
	"fmt"
	"os"
	"encoding/binary"
    "log"
    "net"
    "time"
    "sync"
    "io/ioutil"
    "encoding/hex"
	"image"
	imgFct "workspace/IMAGE"
)

var wg2 sync.WaitGroup



func main() {
	var compteur int = compte_Images("./data")
	images := make([]image.Image, compteur)
	var buff string
	erreurs := make([]error, compteur)
	for i := 1; i<=compteur; i++ {
		buff = fmt.Sprintf("./data/img%d.jpg", i)
		fmt.Println("buff num ", i," = ", buff)
		images[i-1], _, erreurs[i-1] = imgFct.GetImageData(buff)
		if erreurs[i-1] != nil {
			fmt.Println("Failed to get image data:", erreurs[i])
			os.Exit(1)
			}
		}
	for i:=1; i<compteur; i++ {
		for j:=(i+1); j<=compteur; j++ {
			wg2.Add(1)
			go compare(images[i-1], images[j-1], i, j)
			}
		}
	wg2.Wait()
}

func compare(im1 image.Image, im2 image.Image, i int, j int) {
	var distance float64 = 0
	distance = imgFct.GetTotalDistance(im1, im2, wg2)
	fmt.Println("distance entre image",i," et ",j," est de ", distance)
}

func compte_Images (dossier string) int {
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



func Init_Tcp() {
    ln, err := net.Listen("tcp", ":8080")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer ln.Close()

    for {
    	fmt.Println("coucou")
        conn, err := ln.Accept()
        if err != nil {
            log.Fatal(err)
        }
        println("Hello")
        go handleIncomingRequests(conn)
    }
}

func handleIncomingRequests(conn net.Conn) {
    fmt.Println("Received a request: " + conn.RemoteAddr().String())
    headerBuffer := make([]byte, 1024)

    n, err := conn.Read(headerBuffer)
    if err != nil || n != 1024 {
        log.Fatal("Erreur de lecture du header ou taille incorrecte")
    }

    if headerBuffer[0] != byte(1) || headerBuffer[1023] != byte(0) {
        log.Fatal("Invalid header markers")
    }

    var name string
    var reps uint32

    reps = binary.BigEndian.Uint32(headerBuffer[1:5])
    lengthOfName := binary.BigEndian.Uint32(headerBuffer[5:9])
    name = string(headerBuffer[9:9+lengthOfName])

    conn.Write([]byte("Header Received"))

    dataBuffer := make([]byte, 1024)

    file, err := os.Create("./received/" + name)
    if err != nil {
        log.Fatal(err)
    }

    for i := 0; i < int(reps); i++ {
        _, err := conn.Read(dataBuffer)
        if err != nil {
            log.Fatal(err)
        }

        if len(dataBuffer) < 1024 {
            log.Fatal("Invalid Segment: Buffer size mismatch")
        }

        segmentNumber := binary.BigEndian.Uint32(dataBuffer[1:5])
        fmt.Printf("Segment Number: %d\n", segmentNumber)

        length := binary.BigEndian.Uint32(dataBuffer[5:9])
        fmt.Printf("File Data: %s\n", hex.EncodeToString(dataBuffer[9:9+length]))

        file.Write(dataBuffer[9 : 9+length])

        if dataBuffer[0] != byte(0) || dataBuffer[1023] != byte(1) {
            log.Fatal("Invalid Segment")
        }

        conn.Write([]byte("Segment Received"))
    }

    time := time.Now().UTC().Format("Monday, 02-Jan-06 15:04:05 MST")
    conn.Write([]byte(time))

    file.Close()
    conn.Close()
}
