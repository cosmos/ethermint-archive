FROM node

RUN apt-get update && apt-get -y dist-upgrade && apt-get install --no-install-recommends -y -q curl python build-essential git ca-certificates

ENV LC_ALL C.UTF-8
ENV LANG C.UTF-8

ADD . /contract
WORKDIR /contract

RUN npm install --save

CMD ["node", "main.js"]
