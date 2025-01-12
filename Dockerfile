FROM golang:1.20-alpine

# set default working dir as '/app'
WORKDIR /app

# copy file go.mod and go.sum into container
COPY go.mod go.sum ./
RUN go mod tidy

# download dependencies
RUN go mod download && go mod verify

# copy project into container
COPY . ./

# build app to binary
RUN go build -o /server ./server

# expose port 8080 for the service
EXPOSE 8080

CMD ["/server"]