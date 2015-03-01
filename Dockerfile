FROM debian:wheezy

RUN echo "APT::Install-Recommends false;" >> /etc/apt/apt.conf.d/recommends.conf && \
    echo "APT::AutoRemove::RecommendsImportant false;" >> /etc/apt/apt.conf.d/recommends.conf && \
    echo "APT::AutoRemove::SuggestsImportant false;" >> /etc/apt/apt.conf.d/recommends.conf && \
    apt-get update && \
    apt-get upgrade -y

RUN apt-get install -y git-core build-essential automake autoconf libtool ca-certificates && \
    git clone https://github.com/machinezone/twemproxy.git -b lwalkin/config-reload /twemproxy && \
    cd /twemproxy && \
    autoreconf -fvi && \
    ./configure && \
    make && \
    cp src/nutcracker /bin/nutcracker && \
    rm -rf /twemproxy && \
    apt-get remove -y git-core build-essential automake autoconf libtool ca-certificates && \
    apt-get -y autoremove

COPY . /go/src/github.com/bobrik/ralph

RUN apt-get install -y git-core curl ca-certificates && \
    curl https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar xz -C /usr/local && \
    PATH=$PATH:/usr/local/go/bin GOPATH=/go go get github.com/bobrik/ralph/cmd/ralph && \
    cp /go/bin/ralph /bin/ralph && \
    rm -rf /go && \
    rm -rf /usr/local/go && \
    apt-get remove -y git-core curl ca-certificates && \
    apt-get autoremove -y

ENTRYPOINT ["/bin/ralph", "-c", "/etc/nutcracker.yml"]
