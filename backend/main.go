package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

const (
	maxUpload = 20 << 20
	maxPixels = 50_000_000
	DEFAULT_EFFORT = 8
)

var contentTypes = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"avif": "image/avif",
}

func writeToBuffer(imgBuf *bytes.Buffer, b []byte, err error) error {
	if err != nil {
		return err
	}
	_, err = imgBuf.Write(b)
	return err
}

func exportImage(imgBuf *bytes.Buffer, image *vips.ImageRef, format string, quality int, lossy bool) (error) {
	switch format {
	case "jpeg":
		bytes, _, error := image.ExportJpeg(&vips.JpegExportParams{Interlace: true, Quality: quality})
			return writeToBuffer(imgBuf, bytes, error)
	case "png":
		bytes, _, error := image.ExportPng(&vips.PngExportParams{Compression: 10, Interlace: true, Quality: quality})
			return writeToBuffer(imgBuf, bytes, error)
	case "gif":
		bytes, _, error := image.ExportGIF(&vips.GifExportParams{Effort: DEFAULT_EFFORT})
			return writeToBuffer(imgBuf, bytes, error)
	case "webp":
		bytes, _, error := image.ExportWebp(&vips.WebpExportParams{Quality: quality, Lossless: !lossy})
			return writeToBuffer(imgBuf, bytes, error)
	case "avif":
		bytes, _, error := image.ExportAvif(&vips.AvifExportParams{Effort: DEFAULT_EFFORT, Quality: int(quality)})
			return writeToBuffer(imgBuf, bytes, error)
	}
	return errors.New("unsupported format: " + format)
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	// Headroom for the multipart boundaries and other fields; the file itself is checked against maxUpload below.
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1<<20)

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Size > maxUpload {
		http.Error(w, "file must be 20 MB or smaller", http.StatusRequestEntityTooLarge)
		return
	}

	format := r.FormValue("format")
	contentType, ok := contentTypes[format]
	if !ok {
		http.Error(w, "format must be one of jpeg, png, gif, webp or avif", http.StatusBadRequest)
		return
	}

	quality := 80
	if q := r.FormValue("quality"); q != "" {
		quality, err = strconv.Atoi(q)
		if err != nil || quality < 1 || quality > 100 {
			http.Error(w, "quality must be 1-100", http.StatusBadRequest)
			return
		}
	}

	// Sequential access only reads the header here, so oversized dimensions are rejected before any pixels decode.
	params := &vips.ImportParams{}
	params.Access.Set(vips.AccessSequential)
	img, err := vips.LoadImageFromReader(file, params)
	if err != nil {
		http.Error(w, "could not read image: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer img.Close()
	if img.Width()*img.Height() > maxPixels {
		http.Error(w, "image dimensions too large", http.StatusBadRequest)
		return
	}
	// JPEG has no alpha, and libvips would otherwise flatten transparency onto black.
	if format == "jpeg" && img.HasAlpha() {
		if err := img.Flatten(&vips.Color{R: 255, G: 255, B: 255}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	lossy := r.FormValue("lossy") == "true"
	var buf bytes.Buffer

	if err := exportImage(&buf, img, format, quality, lossy); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(buf.Bytes())
}

func createHandler(mux *http.ServeMux) http.Handler {
	isDevelopment := os.Getenv("ENVIRONMENT") == "development"
	allowedOrigins := strings.Split(strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS")), ",")
	
	c := cors.New(cors.Options{    
		AllowedOrigins: allowedOrigins,
		AllowCredentials: true,
		Debug: isDevelopment,})

	return c.Handler(mux)
}

func main() {
	// .env is optional; in production, env vars come from the host
	_ = godotenv.Load()

	vips.Startup(nil)
	defer vips.Shutdown()

	mux := http.NewServeMux()
	
	mux.HandleFunc("POST /convert", convertHandler)

	handler := createHandler(mux)
	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
