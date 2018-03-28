FROM ubuntu:latest

EXPOSE 15567

ARG PROJECT=squawker
ARG VERSION=v0.4.0+3

RUN apt-get update && \
    apt-get -y install vlc

WORKDIR /home
COPY $PROJECT-$VERSION.linux.amd64 sq
RUN chmod +x sq

CMD ["./sq"]
