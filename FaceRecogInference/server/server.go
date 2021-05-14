package main

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ArmadaStore/devices/cargo"
	face "github.com/nikhs247/goface"
)

const modelDir = "models"
const trainDir = "images/Train"

var fileNameCounter = 0

type faceRecogData struct {
	rec       *face.Recognizer
	labelmap  map[int32]string
	cargoInfo *cargo.CargoInfo
	mutex     *sync.Mutex
}

func logTime() {
	currTime := time.Now()
	fmt.Fprintf(os.Stderr, "%s", currTime.Format("2021-01-02 13:01:02 :: "))
}

func faceRecognitionSystem(frd *faceRecogData) {
	fmt.Fprintf(os.Stderr, "Facial Recognition System\n")
	rec, err := face.NewRecognizer(modelDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot initialize recognizer")
	}
	//defer rec.Close()
	frd.rec = rec
	fmt.Fprintf(os.Stderr, "Recognizer Initialized\n")

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

	rand.Seed(time.Now().UnixNano())
	for _, file := range files {
		trainImage := filepath.Join(trainDir, file)

		faces, err := frd.rec.RecognizeFile(trainImage)
		if err != nil {
			log.Fatalf("Can't recognize: %v", err)
		}

		for _, f := range faces {
			samples = append(samples, f.Descriptor)
			ID := rand.Int31n(math.MaxInt32)
			// Each face is unique on that image so goes to its own category.
			people = append(people, ID)
			name := strings.TrimSuffix(file, filepath.Ext(file))
			imageInfo := fmt.Sprintf("%d - %s - %v\n", ID, name, f.Descriptor)
			frd.cargoInfo.Write("id_label_desc.txt", imageInfo)
			frd.labelmap[ID] = name
		}
	}

	// Pass samples to the recognizer.
	frd.rec.SetSamples(samples, people)
}

func (frd *faceRecogData) uploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println("Error 1:")
		log.Fatal(err)
	}
	imgStr := r.Form.Get("img")
	imgByte := []byte(imgStr)
	imgData := bytes.NewReader(imgByte)
	img, err := jpeg.Decode(imgData)
	// img, err := jpeg.Decode(r.Body)
	if err != nil {
		log.Println("Error 1:")
		log.Fatal(err)
	}

	// fileName := fmt.Sprintf("face_%d.jpg", fileNameCounter)
	sendImgName := r.Form.Get("name")
	fileName := fmt.Sprintf("%s.jpg", sendImgName)
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

	///////////////////////////////////////////////////////////////////
	// Testimg with new images

	start := time.Now()
	testImage := fileName
	res, err := frd.rec.RecognizeSingleFile(testImage)
	if err != nil {
		log.Fatalf("Can't recognize: %v", err)
	}
	if res == nil {
		log.Fatalf("Not a single face on the image")
	}

	imageID := frd.rec.Classify(res.Descriptor)
	resName := ""
	if imageID > 0 {
		resName = frd.labelmap[int32(imageID)]
	}

	duration := time.Since(start)

	logTime()
	fmt.Println("Processing time: ", duration)

	if imageID < 0 || !strings.Contains(resName, sendImgName) {

		readData := frd.cargoInfo.Read("id_label_desc.txt")

		// Recognize and label faces
		var samples []face.Descriptor
		var people []int32

		records := strings.Split(readData, "\n")
		nRecs := len(records)

		for i := 0; i < nRecs-1; i++ {
			halfSplit := strings.Split(records[i], "[")

			// Descriptor
			descTemp := strings.TrimSuffix(halfSplit[1], "]")
			descStr := strings.Split(descTemp, " ")
			var desc face.Descriptor
			for j := 0; j < len(descStr); j++ {
				res, err := strconv.ParseFloat(descStr[j], 32)
				if err != nil {
					fmt.Println("Error converting string to float32")
				}
				desc[j] = float32(res)
			}

			// Name and ID
			nameIDTemp := strings.Split(halfSplit[0], "-")
			imgName := strings.TrimSuffix(nameIDTemp[1], " ")
			imgIDStr := strings.TrimSuffix(nameIDTemp[0], " ")
			id, err := strconv.ParseInt(imgIDStr, 10, 32)
			if err != nil {
				fmt.Println("imgIDStr ", imgIDStr)
				fmt.Println("Error converting string to int32")
			}
			imgID := int32(id)
			samples = append(samples, desc)
			people = append(people, imgID)
			frd.labelmap[imgID] = imgName
		}
		frd.rec.SetSamples(samples, people)
		imageID = frd.rec.Classify(res.Descriptor)
		if imageID > 0 {
			resName = frd.labelmap[int32(imageID)]
		}

		if !strings.Contains(resName, sendImgName) {
			resName = "Not found"
			ID := rand.Int31n(math.MaxInt32)
			imageInfo := fmt.Sprintf("%d - %s - %v\n", ID, sendImgName, res.Descriptor)
			frd.cargoInfo.Write("id_label_desc.txt", imageInfo)
			frd.labelmap[ID] = sendImgName
		}
	}
	fmt.Fprintf(w, resName)
}

func setupComm(frd *faceRecogData, IP string, Port string, AppID string, UserID string) {

	http.HandleFunc("/upload", frd.uploadImage)
	http.ListenAndServe(":8090", nil)
}

func main() {

	IP := os.Args[1]
	Port := os.Args[2]
	AppID := os.Args[3]
	UserID := os.Args[4]
	fmt.Println("Server starting")
	var frd faceRecogData
	frd.mutex = &sync.Mutex{}
	frd.labelmap = make(map[int32]string)
	frd.cargoInfo = cargo.InitCargo(IP, Port, AppID, UserID)
	faceRecognitionSystem(&frd)
	setupComm(&frd, IP, Port, AppID, UserID)
}
