FROM golang:1.26-bookworm AS build

WORKDIR /src

# Cache go modules
COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

# Download tailwind standalone
ADD https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64 /usr/local/bin/tailwindcss
RUN chmod +x /usr/local/bin/tailwindcss

# Copy source
COPY . .

# Build tailwind CSS
RUN tailwindcss -i styles/input.css -o cmd/web/static/css/tailwind.css --minify

# Generate templ files
RUN templ generate

# Build binary
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/shelf ./cmd/web

FROM gcr.io/distroless/static-debian12

COPY --from=build /bin/shelf /bin/shelf

ENTRYPOINT ["/bin/shelf"]
