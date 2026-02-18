FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /confighub-demobridge .

FROM --platform=linux/arm64 alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /confighub-demobridge /usr/local/bin/confighub-demobridge
ENTRYPOINT ["confighub-demobridge"]
