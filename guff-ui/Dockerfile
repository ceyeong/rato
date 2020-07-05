FROM golang:1.14.1-alpine3.11

#set working dir
WORKDIR /app

#copy go mod and go sum file to workdir
COPY go.mod go.sum ./

#download dependencies
RUN go mod download

#copy source to workdir
COPY . .

#build the app
RUN go build -o app .

#run the executable
CMD ["./app"]