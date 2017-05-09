FROM alpine:latest

MAINTAINER Andrey

WORKDIR "/opt"

ADD .docker_build/go_player /opt/bin/go_player
ADD ./templates /opt/templates
ADD ./static /opt/static

CMD ["/opt/bin/go_player"]

