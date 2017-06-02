FROM alpine:3.6

# This is the release of ethermint to pull in.
ENV EM_VERSION 0.2.0
ENV EM_SHA256SUM d9d003eec74e05102f10b40be611445ec3179a943c58dd27d5c97a4a43f65f15

ENV DATA_ROOT /ethermint

# Set user right away for determinism
RUN addgroup emuser && \
    adduser -S -G emuser emuser

# Create directory for persistence and give our user ownership
RUN mkdir -p $DATA_ROOT && \
    chown -R emuser:emuser $DATA_ROOT

# jq and curl used for extracting `pub_key` from private validator while
# deploying tendermint with Kubernetes. It is nice to have bash so the users
# could execute bash commands.
RUN apk add --no-cache bash

RUN apk add --no-cache openssl && \
    wget https://s3.eu-central-1.amazonaws.com/ethermint/${EM_VERSION}/ethermint_${EM_VERSION}_ethermint-linux-amd64.zip && \
    echo "${EM_SHA256SUM}  ethermint_${EM_VERSION}_ethermint-linux-amd64.zip" | sha256sum -c && \
    unzip -d /bin ethermint_${EM_VERSION}_ethermint-linux-amd64.zip && \
    mv /bin/ethermint-linux-amd64 /bin/ethermint && \
    rm -f ethermint_${EM_VERSION}_ethermint-linux-amd64.zip

# Expose the data directory as a volume since there's mutable state in there
VOLUME $DATA_ROOT

EXPOSE 46658

ENTRYPOINT ["ethermint"]
