FROM golang:latest

MAINTAINER yafim.kazak@gmail.com

# ENV
ENV NAME="consul-101-queue-consumer"
ENV WORKDIR="/app"

WORKDIR $WORKDIR

# Copying config
COPY . $WORKDIR

RUN go build -o MessageQueueConsumer .

EXPOSE 8080

CMD ["./MessageQueueConsumer"]