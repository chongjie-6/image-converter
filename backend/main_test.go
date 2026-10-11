package main

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func post(src []byte, format string) *httptest.ResponseRecorder {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, _ := mw.CreateFormFile("file", "in.png")
	fw.Write(src)
	mw.WriteField("format", format)
	mw.WriteField("quality", "50")
	mw.Close()

	req := httptest.NewRequest("POST", "/convert", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	convertHandler(rec, req)
	return rec
}

func TestConvert(t *testing.T) {
	var src bytes.Buffer
	png.Encode(&src, image.NewNRGBA(image.Rect(0, 0, 16, 16)))

	for format := range contentTypes {
		rec := post(src.Bytes(), format)
		if rec.Code != 200 {
			t.Fatalf("%s: status %d: %s", format, rec.Code, rec.Body)
		}
		if _, got, err := image.DecodeConfig(rec.Body); err != nil || got != format {
			t.Errorf("%s: decoded as %q, err %v", format, got, err)
		}
	}

	if rec := post(src.Bytes(), "bmp"); rec.Code != 400 {
		t.Errorf("bmp: want 400, got %d", rec.Code)
	}
	if rec := post([]byte("not an image"), "png"); rec.Code != 400 {
		t.Errorf("garbage: want 400, got %d", rec.Code)
	}

	// The PNG decoder stops at IEND, so trailing padding still decodes.
	atLimit := append(src.Bytes(), make([]byte, maxUpload-src.Len())...)
	if rec := post(atLimit, "png"); rec.Code != 200 {
		t.Errorf("file at limit: want 200, got %d: %s", rec.Code, rec.Body)
	}
	if rec := post(append(atLimit, 0), "png"); rec.Code != 413 {
		t.Errorf("file over limit: want 413, got %d", rec.Code)
	}
}
