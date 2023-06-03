FROM ubuntu:20.04
# privileged
USER root
# use aliyun images.
ARG DEBIAN_FRONTEND=noninteractive
RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list
ARG DEBIAN_FRONTEND=noninteractive
RUN apt clean
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update
# install basics
ARG DEBIAN_FRONTEND=noninteractive
RUN apt install -y sudo git cmake g++ gcc vim tar gdb openssh-server rsync python3.8 python3-pip dos2unix clang-format-10 apt-transport-https ca-certificates curl gnupg-agent software-properties-common
RUN mkdir /home/src && mkdir /home/compile && mkdir /home/result
COPY ./gpuserver /home/
CMD /home/gpuserver
