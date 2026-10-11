# Image Converter

Convert images between JPEG, PNG, GIF, WebP and AVIF in the browser. A Next.js frontend uploads the image to a Go backend, which decodes and re-encodes it with [libvips](https://www.libvips.org/) and returns the converted file.

```
frontend/   Next.js 16 + React 19 + Tailwind 4 UI
backend/    Go HTTP API using govips (libvips bindings)
```

## Features

- Output formats: JPEG, PNG, GIF, WebP (lossy or lossless) and AVIF
- Quality slider (1–100) for lossy output
- Shows the size before and after conversion, with a preview and a download link
- Transparent images converted to JPEG are flattened onto white instead of black
- Uploads are limited to 20 MB and 50 megapixels, and oversized images are rejected before their pixels are decoded

## Running locally

### Backend

You need Go 1.27+ and libvips with its development headers (`libvips-dev` on Debian/Ubuntu, `brew install vips` on macOS). The bindings use cgo, so the easiest path on Windows is Docker or WSL.

```bash
cd backend
echo "ALLOWED_ORIGINS=http://localhost:3000" > .env
go run .
```

Or with Docker:

```bash
cd backend
docker build -t image-converter-backend .
docker run -p 8080:8080 -e ALLOWED_ORIGINS=http://localhost:3000 image-converter-backend
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Open http://localhost:3000.

## Configuration

### Backend (`backend/.env` or host environment)

| Variable          | Description                                                        | Default |
| ----------------- | ------------------------------------------------------------------ | ------- |
| `ALLOWED_ORIGINS` | Comma-separated list of origins allowed by CORS, e.g. `http://localhost:3000` | none    |
| `ENVIRONMENT`     | Set to `development` to turn on CORS debug logging                 | none    |
| `PORT`            | Port the server listens on                                         | `8080`  |

`.env` is optional. In production, set these on the host.

### Frontend

| Variable              | Description          | Default                 |
| --------------------- | -------------------- | ----------------------- |
| `NEXT_PUBLIC_API_URL` | Base URL of the backend | `http://localhost:8080` |

## API

### `POST /convert`

`multipart/form-data` with:

| Field     | Required | Description                                             |
| --------- | -------- | ------------------------------------------------------- |
| `file`    | yes      | The image to convert (20 MB max)                         |
| `format`  | yes      | `jpeg`, `png`, `gif`, `webp` or `avif`                   |
| `quality` | no       | 1–100, default `80`                                      |
| `lossy`   | no       | `true` for lossy WebP; any other value gives lossless WebP |

Returns the converted image with the matching `Content-Type`. Errors come back as plain text with status `400` (bad input), `413` (file too large) or `500` (encoding failed).

```bash
curl -F file=@photo.png -F format=webp -F quality=75 -F lossy=true \
  http://localhost:8080/convert -o photo.webp
```

### `GET /ping`

Returns `pong`. The frontend calls it when the page loads so a scaled-to-zero backend starts warming up before the user submits an image.

## Tests

```bash
cd backend
go test ./...
```

The tests convert a small PNG to every supported format and check the rejection paths (unsupported format, invalid image, file over the size limit). They need libvips installed.
