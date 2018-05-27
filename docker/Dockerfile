# Use phusion/baseimage as base image. To make your builds
# reproducible, make sure you lock down to a specific version, not
# to `latest`! See
# https://github.com/phusion/baseimage-docker/blob/master/Changelog.md
# for a list of version numbers.
FROM phusion/baseimage:0.10.1

# Use baseimage-docker's init system.
CMD ["/sbin/my_init"]

# ...put your own build instructions here...

RUN apt-get update
RUN apt-get -y upgrade

##### INSTALL PACKAGES #####
RUN apt-get -y install git
RUN apt-get -y install wget
RUN apt-get -y install unzip
RUN apt-get -y install golang

##### SETUP GO #####
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN mkdir -p $GOPATH/src

##### INSTALL SIMPLE_SMTP #####
RUN wget -O /tmp/simpleSMTP-master.zip https://codeload.github.com/julianweise/simpleSMTP/zip/master
RUN unzip /tmp/simpleSMTP-master.zip -d $GOPATH/src
RUN mv $GOPATH/src/simpleSMTP-master $GOPATH/src/simpleSMTP

RUN mkdir -p $GOPATH/src/simpleSMTP/build

##### GENERATE SERVER-CRT/KEY #####
RUN openssl ecparam -genkey -name secp384r1 -out $GOPATH/src/simpleSMTP/build/server.key
RUN openssl req -new -x509 -sha256 -key $GOPATH/src/simpleSMTP/build/server.key -out $GOPATH/src/simpleSMTP/build/server.crt -days 3650 \
    -subj "/C=DE/ST=Brandenburg/L=Potsdam/O=HPI/OU=SRE 2018/CN=julian.weise@student.hpi.de"
    
##### BUILD/RUN SIMPLE_SMTP #####
RUN cd $GOPATH/src/simpleSMTP && \
    go get && \
    go build
EXPOSE 4431

##### ADD SMTP DEAMON #####
RUN mkdir -p /etc/service/simpleSMTP
RUN echo "#!/bin/sh\ncd $GOPATH/src/simpleSMTP && exec ./simpleSMTP" > /etc/service/simpleSMTP/run
RUN chmod +x /etc/service/simpleSMTP/run

# Clean up APT when done.
RUN apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*