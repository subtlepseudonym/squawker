FROM ubuntu:latest

EXPOSE 15567
ENV BIN squawker-v0.1.0+1.linux.amd64
RUN apt-get update && \
    apt-get -y install vlc

WORKDIR /root
COPY $BIN sq
RUN chmod +x sq

CMD ["./sq"]
