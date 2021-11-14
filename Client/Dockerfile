 
# syntax=docker/dockerfile:1

FROM ubuntu:20.04

#Install utils and dependencies
RUN apt-get update
RUN apt-get install -y openssh-server
RUN apt-get install -y golang
RUN apt-get install -y git
RUN go get github.com/gorilla/mux

#Configure and allow ssh access
RUN mkdir /var/run/sshd
RUN echo 'root:root' | chpasswd
RUN sed -i 's/#*PermitRootLogin prohibit-password/PermitRootLogin yes/g' /etc/ssh/sshd_config
RUN service ssh start

#Copy go file
COPY Client.go ./SDCC/
CMD cd SDCC && go run Client.go

EXPOSE 8080
EXPOSE 22
