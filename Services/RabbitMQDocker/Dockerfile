FROM rabbitmq:3.7.17-management

ENV RABBITMQ_ERLANG_COOKIE: "SWQOKODSQALRPCLNMEQG"
ENV RABBITMQ_DEFAULT_USER: "guest"
ENV RABBITMQ_DEFAULT_PASS: "guest"
ENV RABBITMQ_DEFAULT_VHOST: "/"

EXPOSE 15672:15672
EXPOSE 5672:5672

COPY rabbitmq.conf /etc/rabbitmq
RUN rabbitmq-plugins enable --offline rabbitmq_peer_discovery_consul
