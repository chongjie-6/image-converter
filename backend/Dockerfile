FROM golang:1.27-trixie AS build
RUN apt-get update && apt-get install -y --no-install-recommends libvips-dev && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /server .

FROM debian:trixie-slim
# libheif-plugin-aomenc provides the AV1 encoder that AVIF export needs
RUN apt-get update && apt-get install -y --no-install-recommends libvips42t64 libheif-plugin-aomenc ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /server /server
EXPOSE 8080
CMD ["/server"]
