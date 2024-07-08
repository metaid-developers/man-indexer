FROM ubuntu:20.04
ARG VER
WORKDIR /man

RUN echo "deb http://mirrors.aliyun.com/ubuntu/ jammy main \n" >> /etc/apt/sources.list
RUN apt-get update
RUN apt-get -y --no-install-recommends install wget
COPY ./get-app.sh /man/get-app.sh
RUN chmod +x get-app.sh
#RUN tar -zxvf manindex-linux.tar.gz
#RUN cp /man/releases/linux/config.toml ./config.toml
RUN apt-get -y --no-install-recommends install libc6
RUN apt-get -y --no-install-recommends install libzmq3-dev

