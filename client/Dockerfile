FROM golang

RUN apt-get -y update

WORKDIR /app
ADD . /app/
RUN go build -o camera camera.go
ENTRYPOINT ["./camera"]