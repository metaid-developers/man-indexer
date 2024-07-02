FROM ubuntu:20.04
ARG VER
WORKDIR /man

RUN echo "deb http://mirrors.aliyun.com/ubuntu/ jammy main \n" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get -y --no-install-recommends install wget
RUN wget --no-check-certificate https://github.com/metaid-developers/man-indexer/releases/download/$VER/manindex-linux.tar.gz
#COPY ./manindex-linux.tar.gz ./
RUN tar -zxvf manindex-linux.tar.gz
RUN cp /man/releases/linux/config.toml ./config.toml
#COPY ./manindexer ./
RUN apt-get -y --no-install-recommends install libc6
RUN apt-get -y --no-install-recommends install libzmq3-dev

