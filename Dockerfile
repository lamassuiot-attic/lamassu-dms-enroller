FROM golang:1.16
WORKDIR /app
COPY . .
WORKDIR /app/cmd/enroller
RUN go mod tidy
WORKDIR /app
RUN CGO_ENABLED=0 go build -o enroller ./cmd/enroller/main.go

FROM scratch
COPY --from=0 /app/enroller /
CMD ["/enroller"]
