package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	imgFile, err := os.Open("Test/Ben_Chandler_0001.jpg")
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

	data := url.Values{}
	data.Set("name", "Ben_Chandler")
	data.Set("img", string(buf.Bytes()))
	resp, err := http.PostForm("http://127.0.0.1:8090/upload", data)

	// sendBuf := bytes.NewReader(buf.Bytes())
	// resp, err := http.Post("http://127.0.0.1:8090/upload", "image/jpeg", sendBuf)
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
