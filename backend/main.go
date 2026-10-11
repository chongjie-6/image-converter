package main

import (
	"bytes"
	"errors"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/chai2010/webp"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/vegidio/avif-go"
)

const (
	maxUpload = 20 << 20
	maxPixels = 50_000_000
	AVIF_SPEED = 8
)

var contentTypes = map[string]string{
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
	"webp": "image/webp",
	"avif": "image/avif",
}

func decode(r io.ReadSeeker) (image.Image, error) {
	// Check dimensions from the header first so a tiny file can't decode into gigabytes of pixels.
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return nil, err
	}
	if cfg.Width*cfg.Height > maxPixels {
		return nil, errors.New("image dimensions too large")
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	img, _, err := image.Decode(r)
	return img, err
}

// flatten composites onto white, since JPEG and the Plan9 GIF palette would otherwise turn transparency black.
func flatten(img image.Image) image.Image {
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), img, img.Bounds().Min, draw.Over)
	return dst
}

func encode(w io.Writer, img image.Image, format string, quality int, lossy bool) error {
	switch format {
	case "jpeg":
		return jpeg.Encode(w, flatten(img), &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(w, img)
	case "gif":
		return gif.Encode(w, flatten(img), nil)
	case "webp":
		return webp.Encode(w, img, &webp.Options{Lossless: !lossy, Quality: float32(quality)})
	case "avif":
		return avif.Encode(w, img, &avif.Options{Speed: AVIF_SPEED, ColorQuality: int(quality)})
	}
	return errors.New("unsupported format")
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

	img, err := decode(file)
	if err != nil {
		http.Error(w, "could not read image: "+err.Error(), http.StatusBadRequest)
		return
	}

	lossy := r.FormValue("lossy") == "true"
	var buf bytes.Buffer
	if err := encode(&buf, img, format, quality, lossy); err != nil {
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

	mux := http.NewServeMux()
	
	mux.HandleFunc("POST /convert", convertHandler)

	handler := createHandler(mux)
	log.Println("listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
