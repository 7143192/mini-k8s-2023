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
ARG DEBIAN_FRONTEND=noninteractive
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
ARG DEBIAN_FRONTEND=noninteractive
RUN sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
ARG DEBIAN_FRONTEND=noninteractive
RUN apt update && apt install -y docker-ce docker-ce-cli containerd.io
# try to install etcd.
ARG DEBIAN_FRONTEND=noninteractive
COPY ./etcd.tar.gz /
RUN cd / && tar -C /usr/local/bin -xvf etcd.tar.gz
# && echo "export PATH=$PATH:/usr/local/bin/etcd" >> /root/.bashrc && /bin/bash -c "source /root/.bashrc"
# try to install golang.
ARG DEBIAN_FRONTEND=noninteractive
COPY ./go.tar.gz /
RUN cd / && rm -rf /usr/local/go && tar -C /usr/local -xvf go.tar.gz && echo export PATH=$PATH:/usr/local/bin/etcd:/usr/local/go/bin >> /root/.bashrc && /bin/bash -c "source /root/.bashrc"
# try to install cni plugin.
ARG DEBIAN_FRONTEND=noninteractive
COPY ./cni.tgz /
RUN mkdir /opt/cni
RUN mkdir /opt/cni/bin
RUN cd / && tar -C /opt/cni/bin -xvf cni.tgz
# try to move some tools into docker.
COPY ./flanneld-amd64 /
COPY ./cadvisor /
COPY ./mk-docker-opts.sh /
CMD /bin/bash
