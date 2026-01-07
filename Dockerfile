# Gunakan Go versi terbaru (sesuai tahun 2026)
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency dulu biar cache efisien
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build menjadi file binary bernama 'server'
RUN go build -o server cmd/api/main.go

# --- Tahap 2: Image Kecil untuk Menjalankan ---
FROM alpine:latest

WORKDIR /root/

# Copy hasil build dari tahap sebelumnya
COPY --from=builder /app/server .
# Copy file .env agar settingan terbaca
COPY --from=builder /app/.env . 

# Buka port 8080
EXPOSE 8080

# Jalankan server
CMD ["./server"]