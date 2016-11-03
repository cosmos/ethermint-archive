FROM golang
MAINTAINER kobigurk@gmail.com

RUN curl https://glide.sh/get | sh
RUN mkdir -p /go/src/github.com/kobigurk
RUN cd /go/src/github.com/kobigurk && git clone http://github.com/kobigurk/tmsp-ethereum
RUN cd /go/src/github.com/kobigurk/tmsp-ethereum && glide install
RUN go get github.com/tendermint/tendermint/cmd/tendermint
RUN go get github.com/ethereum/go-ethereum/cmd/geth
RUN cd /go/src/github.com/kobigurk/tmsp-ethereum/vendor/github.com/tendermint/tendermint/cmd/tendermint && go install .
RUN cd /go/src/github.com/kobigurk/tmsp-ethereum/vendor/github.com/ethereum/go-ethereum/cmd/geth && go install .
RUN cd /go/src/github.com/kobigurk/tmsp-ethereum/cmd/tmsp-ethereum && go install .
RUN apt-get update && apt-get install -y supervisor
RUN mkdir -p /var/log/supervisor && chmod 777 /var/log/supervisor
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

RUN mkdir /tmsp-ethereum && chmod 777 /tmsp-ethereum
RUN useradd -ms /bin/bash tmsp-ethereum
RUN chown tmsp-ethereum /var/run
USER tmsp-ethereum

RUN mkdir -p /tmsp-ethereum/setup
COPY genesis.json /tmsp-ethereum/setup/
RUN geth -datadir /tmsp-ethereum/data init /tmsp-ethereum/setup/genesis.json
COPY keystore /tmsp-ethereum/data/keystore
EXPOSE 8545

RUN tendermint init

CMD ["/usr/bin/supervisord"]
