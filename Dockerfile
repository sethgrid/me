FROM google/debian:wheezy
MAINTAINER Seth Ammons <seth.ammons@gmail.com>

RUN apt-get update && apt-get install -y curl tar python python-pip && pip install Pygments

# grab the specifc version of hugo, and raname and move it, and append to path
RUN curl -L http://github.com/spf13/hugo/releases/download/v0.13/hugo_0.13_linux_amd64.tar.gz | tar xz
RUN mv hugo_0.13_linux_amd64 hugo &&  mv /hugo/hugo_0.13_linux_amd64 /hugo/hugo

ENV PATH $PATH:/hugo

WORKDIR /me/hugo

EXPOSE 1313

ADD . /me
