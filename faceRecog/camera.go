package main

import (
	"bytes"
	"log"
	"net/http"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
)

func main(){
	imgFile, err := os.Open("images/Test/Angela_Merkel_0004.jpg")
	if err != nil {
		log.Println("Error 1:")
		log.Fatal(err)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Println("Error 2:")
		log.Fatal(err)
	}

	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, img, nil)
	if err != nil {
		log.Println("Error 3:")
		log.Fatal(err)
	}

	sendBuf := bytes.NewReader(buf.Bytes())
	resp, err := http.Post("http://127.0.0.1:8080/upload", "image/jpeg", sendBuf)
	if err != nil {
		log.Println("Error 4:")
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error 5:")
		log.Fatal(err)
	}

	log.Println(string(body))
}