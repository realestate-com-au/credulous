# I need to be able to replicate the Travis build environment in order
# to troubleshoot some problems with it in a reasonable fast way.
# Their build systems are Ubuntu Server 12.04 LTS (aka Precise Pangolin)
FROM ubuntu:precise
MAINTAINER Colin.Panisset <colin.panisset@rea-group.com>
RUN apt-get update
RUN apt-get install -y python-software-properties
RUN add-apt-repository ppa:pdoes/ppa
RUN apt-get update
RUN apt-get install -y git mercurial subversion \
	curl wget clang gcc openssl rsync
ADD http://golang.org/dl/go1.2.2.linux-amd64.tar.gz /tmp/go1.2.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf /tmp/go1.2.2.linux-amd64.tar.gz
RUN ln -s /usr/local/go/bin/go /usr/bin/go
RUN ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt
RUN ln -s /usr/local/go/bin/godoc /usr/bin/godoc
