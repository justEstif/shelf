# ─── Build ───
FROM golang:1.26-bookworm AS build

WORKDIR /src

# Install build tools
ADD https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 /usr/bin/tailwindcss
RUN chmod +x /usr/bin/tailwindcss
RUN go install github.com/a-h/templ/cmd/templ@latest

# Cache Go module downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN mkdir -p cmd/web/static/css && \
    templ generate && \
    tailwindcss -i styles/input.css -o cmd/web/static/css/tailwind.css --minify && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /shelf cmd/web/main.go

# ─── Runtime ───
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /shelf /shelf

ENV SHELF_PORT=3000
ENV SHELF_PAGES_DIR=/shelf/pages
ENV SHELF_DATA_DIR=/shelf/data
EXPOSE 3000

VOLUME /shelf/pages /shelf/data

ENTRYPOINT ["/shelf"]
