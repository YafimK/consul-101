FROM golang:latest

MAINTAINER yafim.kazak@gmail.com

# ENV
ENV NAME="consul-101"
ENV WORKDIR="app"

WORKDIR $WORKDIR

# Copying config
COPY . $WORKDIR

RUN go mod download
RUN go build -o producer .

EXPOSE 8080

CMD ["./producer"]