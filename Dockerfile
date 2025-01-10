# use official Golang image
FROM golang:1.22.2
# set working directory
WORKDIR /app
# copy the source code
COPY . .
# download and install the dependencies
RUN go get -d -v ./...
# build the Go app
RUN go build -o api .
# expose the port
EXPOSE 8000
# tun the executable
CMD ["./api"]