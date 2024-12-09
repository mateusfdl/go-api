FROM golang:1.23 AS build
ENV CGO_ENABLED=0
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN go build -o main cmd/app/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=build /app/main /app/main
CMD ["/app/main"]

