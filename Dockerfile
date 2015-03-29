FROM google/debian:wheezy
MAINTAINER Seth Ammons <seth.ammons@gmail.com>

RUN /bin/echo "grabbing required applications"
RUN apt-get update
RUN apt-get install curl -y
RUN apt-get install tar -y
RUN apt-get install python -y
RUN apt-get install python-pip -y
RUN apt-get install git -y

RUN pip install Pygments

RUN /bin/echo "grabbing and installing hugo tar.gz"
# grab the specifc version of hugo, and raname and move it, and append to path
RUN curl -L http://github.com/spf13/hugo/releases/download/v0.13/hugo_0.13_linux_amd64.tar.gz | tar xz
RUN mv hugo_0.13_linux_amd64 hugo
RUN mv /hugo/hugo_0.13_linux_amd64 /hugo/hugo

ENV PATH $PATH:/hugo

RUN /bin/echo "cloning down sendgrid/me"
RUN git clone https://github.com/sethgrid/me.git me
WORKDIR me/hugo
RUN mkdir -p themes
WORKDIR themes
RUN git clone https://github.com/gizak/nofancy.git

WORKDIR /me/hugo
RUN /bin/echo "starting hugo..."
EXPOSE 1313
#RUN hugo server --theme=nofancy --buildDrafts

