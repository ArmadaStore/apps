package main

import (
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/ArmadaStore/devices/cargo"
	face "github.com/nikhs247/goface"
)

const modelDir = "models"
const trainDir = "images/Train"

var fileNameCounter = 0

type faceRecogData struct {
	rec       *face.Recognizer
	labels    []string
	cargoInfo *cargo.CargoInfo
}

func faceRecognitionSystem(frd *faceRecogData) {
	fmt.Fprintf(os.Stderr, "Facial Recognition System")
	rec, err := face.NewRecognizer(modelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot initialize recognizer")
	}
	//defer rec.Close()
	frd.rec = rec
	fmt.Fprintf(os.Stderr, "Recognizer Initialized")

	///////////////////////////////////////////////////////////////////

	// traverse the train folder and label images by their filename

	var files []string
	err = filepath.Walk(trainDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
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
			fmt.Println("Type and size of descriptor: ", reflect.TypeOf(f.Descriptor), " ", len(f.Descriptor))
		}

		ID = ID + 1

		frd.labels = append(frd.labels, strings.TrimSuffix(file, filepath.Ext(file)))
	}
	// Pass samples to the recognizer.
	frd.rec.SetSamples(samples, people)
}

func (frd *faceRecogData) uploadImage(w http.ResponseWriter, r *http.Request) {
	img, err := jpeg.Decode(r.Body)
	if err != nil {
		log.Println("Error 1:")
		log.Fatal(err)
	}

	fileName := fmt.Sprintf("face_%d.jpg", fileNameCounter)
	imgFile, err := os.Create(fileName)
	if err != nil {
		log.Println("Error 2:")
		log.Fatal(err)
	}
	fileNameCounter++

	err = jpeg.Encode(imgFile, img, nil)
	if err != nil {
		imgFile.Close()
		log.Println("Error 3:")
		log.Fatal(err)
	}
	imgFile.Close()

	// frd.cargoInfo.Send(fileName)

	///////////////////////////////////////////////////////////////////
	// Testimg with new images

	testImage := fileName
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

func setupComm(frd *faceRecogData, IP string, Port string, AppID string, UserID string) {
	// frd.cargoInfo = cargo.InitCargo(IP, Port, AppID, UserID)
	http.HandleFunc("/upload", frd.uploadImage)
	http.ListenAndServe(":8080", nil)
}

func main() {

	IP := os.Args[1]
	Port := os.Args[2]
	AppID := os.Args[3]
	UserID := os.Args[4]
	fmt.Println("Server starting")
	var frd faceRecogData
	faceRecognitionSystem(&frd)
	setupComm(&frd, IP, Port, AppID, UserID)
	// frd.cargoInfo.CleanUp()
}
