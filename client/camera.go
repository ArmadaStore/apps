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
	"path/filepath"
	"strings"
	"time"
	"fmt"
)

func main() {
	taskip := os.Args[1]
	taskport := os.Args[2]
	pattern := "*.jpg"
	test := "Test"
	err := filepath.Walk(test, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            return nil
        }
        if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
            return err
        } else if matched {
            imgFile, err := os.Open(path)
			if err != nil {
				log.Fatal(err)
			}
			defer imgFile.Close()

			img, _, err := image.Decode(imgFile)
			if err != nil {
				log.Fatal(err)
			}

			buf := new(bytes.Buffer)
			err = jpeg.Encode(buf, img, nil)
			if err != nil {
				log.Println("Error 3:")
				log.Fatal(err)
			}

			// face name
			justfilename := filepath.Base(path)
			ext := filepath.Ext(justfilename)
			filename := strings.TrimSuffix(justfilename, ext)
			splitfilename := strings.Split(filename, "_")
			name := splitfilename[0] + "_" + splitfilename[1]


			data := url.Values{}
			data.Set("name", name)
			data.Set("img", string(buf.Bytes()))
			start := time.Now()
			resp, err := http.PostForm("http://" + taskip + ":" + taskport + "/upload", data)
			end := time.Since(start)
			fmt.Println("P ", end)

			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			_, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
        }
        return nil
    })
    if err != nil {
		log.Fatal(err)
	}
}