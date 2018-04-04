FROM alpine:latest
MAINTAINER Toni Moreno <toni.moreno@gmail.com>

ADD ./pseriescollector-last.tar.gz /

VOLUME ["/opt/pseriescollector/conf", "/opt/pseriescollector/log"]

EXPOSE 8090

WORKDIR /opt/pseriescollector
COPY ./pseriescollector.toml ./conf/
COPY ./start.sh /

ENTRYPOINT ["/start.sh"]
