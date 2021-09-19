package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"gopkg.in/gographics/imagick.v2/imagick"
)

func heightToWidthRatio(im *imagick.MagickWand) float64 {
	return (float64)(im.GetImageHeight()) / (float64)(im.GetImageWidth())
}

func resize(mw *imagick.MagickWand, width int) (err error) {
	var height uint

	baseWidth := int(mw.GetImageWidth())
	if width > baseWidth {
		width = baseWidth
	}

	ratio := heightToWidthRatio(mw)
	height = uint((float64)(width) * ratio)

	err = mw.ResizeImage(uint(width), height, imagick.FILTER_LANCZOS, 1)
	if err != nil {
		return
	}

	return
}

func getImageHandler(w http.ResponseWriter, req *http.Request) {
	// read url params
	vars := mux.Vars(req)
	src := "./static/" + vars["src"]
	width, _ := strconv.Atoi(req.URL.Query().Get("w"))
	quality, _ := strconv.Atoi(req.URL.Query().Get("q"))
	log.Println(width, quality, src)

	// initialize libmagic
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	// read image from src
	if err := mw.ReadImage(src); err != nil {
		log.Fatal(err)
	}

	// resize
	if err := resize(mw, width); err != nil {
		log.Fatal(err)
	}

	// save output file
	mw.WriteImage("output.jpg")

	// send file as response
	file, _ := os.Open("output.jpg")
	defer file.Close()
	w.Header().Set("Content-Type", "image/jpeg")
	io.Copy(w, file)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/image/{src}", getImageHandler)

	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println("starting")
	log.Fatal(srv.ListenAndServe())
}
