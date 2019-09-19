FROM golang:latest

MAINTAINER yafim.kazak@gmail.com

# ENV
ENV NAME="consul-101-queue-producer"
ENV WORKDIR="/app"

WORKDIR $WORKDIR

# Copying config
COPY . $WORKDIR

RUN go build -o MessageQueueProducer .

EXPOSE 8080

CMD ["./MessageQueueProducer"]