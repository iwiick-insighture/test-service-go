FROM golang:1.20-alpine AS build
WORKDIR /go/src/app
COPY . .
RUN go build -o secret-service ./main.go

FROM golang:1.20-alpine 
RUN apk upgrade libssl3 libcrypto3 
WORKDIR /app
COPY --from=build /go/src/app/secret-service /app
EXPOSE 8080
CMD ["./secret-service"]