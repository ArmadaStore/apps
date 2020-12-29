package main

import (
	"net/http"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"strings"
	"path/filepath"

	"github.com/nikhs247/goface"

)

const modelDir = "models"
const trainDir = "images/Train"

type faceRecogData struct {
	rec *face.Recognizer
	labels []string
}

func faceRecognitionSystem(frd *faceRecogData) {
	fmt.Println("Facial Recognition System")
	rec, err := face.NewRecognizer(modelDir)
	if err != nil {
		fmt.Println("Cannot initialize recognizer")
	}
	//defer rec.Close()
	frd.rec = rec
	fmt.Println("Recognizer Initialized")

	///////////////////////////////////////////////////////////////////

	// traverse the train folder and label images by their filename

	var files []string
	err = filepath.Walk(trainDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir(){
			return nil
		}
		_, file := filepath.Split(path)
        files = append(files, file)
        return nil
    })
    if err != nil {
        panic(err)
    }

    // Recognize and label faces
	var samples []face.Descriptor
	var people []int32
	ID := 0
	for _, file := range files {
        trainImage := filepath.Join(trainDir, file)

		faces, err := frd.rec.RecognizeFile(trainImage)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}

		
		for _, f := range faces {
			samples = append(samples, f.Descriptor)
			// Each face is unique on that image so goes to its own category.
			people = append(people, int32(ID))
		}

		ID = ID + 1

		frd.labels = append(frd.labels, strings.TrimSuffix(file, filepath.Ext(file)))
    }    
    // Pass samples to the recognizer.
	frd.rec.SetSamples(samples, people)
}


func (frd *faceRecogData)uploadImage(w http.ResponseWriter, r *http.Request) {
	img, err := jpeg.Decode(r.Body)
	if err != nil {
		log.Println("Error 1:")
		log.Fatal(err)
	}

	imgFile, err := os.Create("receivedImag.jpg")
	if err != nil {
		log.Println("Error 2:")
		log.Fatal(err)
	}

	err = jpeg.Encode(imgFile, img, nil)
	if err != nil {
		imgFile.Close()
		log.Println("Error 3:")
		log.Fatal(err)	
	}
	imgFile.Close()

	///////////////////////////////////////////////////////////////////
	// Testimg with new images
	
	testImage := "receivedImag.jpg"
	res, err := frd.rec.RecognizeSingleFile(testImage)
	if err != nil {
		log.Fatalf("Can't recognize: %v", err)
	}
	if res == nil {
		log.Fatalf("Not a single face on the image")
	}

	imageID := frd.rec.Classify(res.Descriptor)
	if imageID < 0 {
		log.Fatalf("Can't classify")
	}

	fmt.Println(frd.labels[imageID])
	fmt.Fprintf(w, frd.labels[imageID])
}

func setupHandles(frd *faceRecogData){
	http.HandleFunc("/upload", frd.uploadImage)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Server starting")
	var frd faceRecogData
	faceRecognitionSystem(&frd)
	setupHandles(&frd)
}