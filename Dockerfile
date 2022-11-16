FROM golang:latest as builder

WORKDIR /app
ENV CGO_ENABLED=0
COPY . .

RUN go build -o . cmd/service/main.go

FROM alpine:latest

COPY --from=builder /app /app

EXPOSE 8080
RUN apk --no-cache add ca-certificates

ENTRYPOINT ["app/main"]