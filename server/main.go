package main

import (
	"fmt"
	"os"
	"encoding/binary"
    "log"
    "net"
    "time"
    "encoding/hex"
	_ "image"
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
