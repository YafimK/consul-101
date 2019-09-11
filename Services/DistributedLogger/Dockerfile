FROM golang:latest

MAINTAINER yafim.kazak@gmail.com

# ENV
ENV NAME="consul-101-logger"
ENV WORKDIR="/app"

WORKDIR $WORKDIR

# Copying config
COPY . $WORKDIR

RUN go build -o logger .

EXPOSE 8080

CMD ["./logger"]