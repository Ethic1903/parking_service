FROM golang:1.24.3-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -o /out/parking-api ./cmd/parking-api

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/parking-api /app/parking-api

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/parking-api"]
