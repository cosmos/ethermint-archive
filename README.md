# [Ethermint](https://github.com/tendermint/ethermint)
### Ethereum powered by Tendermint consensus

### [Download the latest release](https://github.com/tendermint/ethermint/releases/tag/v0.2.2)

[![GitHub release](https://img.shields.io/badge/release-latest-blue.svg)]() [![](https://circleci.com/gh/tendermint/ethermint/tree/master.svg?style=shield)](https://circleci.com/gh/tendermint/ethermint/tree/master) [![](https://tokei.rs/b1/github/tendermint/ethermint)](https://github.com/tendermint/ethermint) [![](https://img.shields.io/badge/go-1.8-blue.svg)](https://github.com/moovweb/gvm) [![](https://img.shields.io/badge/issues-7-yellow.svg)](https://github.com/tendermint/ethermint/issues) [![License](https://img.shields.io/badge/license-GPLv3.0%2B-blue.svg)](https://www.gnu.org/licenses/gpl-3.0.html)

### Join the chat!
[![](https://img.shields.io/badge/slack-join%20chat-brightgreen.svg)](http://forum.tendermint.com:3000/)

We have a friendly community of like-minded people which are always eager to help someone in need of advice or just
looking for casual banter.

### Code of Conduct
Please read, understand and adhere to our [code of conduct](https://github.com/tendermint/ethermint/blob/develop/CODE_OF_CONDUCT.md).

### Documentation
We are using `godoc` for our documentation. In order to browse the [interactive documentation](http://localhost:6060/pkg/github.com/tendermint/ethermint/) please open a terminal window and navigate
to the ethermint folder and run godoc.
```bash
cd $GOPATH/src/github.com/tendermint/ethermint

godoc -http :6060
```


----


## About Ethermint

Ethermint enables ethereum to run as an [ABCI](https://github.com/tendermint/abci) application on tendermint and the COSMOS hub. This application allows you to get all the benefits of ethereum without having to run your own miners.

This means running an Ethereum EVM-based network that uses Tendermint consensus instead of proof-of-work.
The way it's built makes it easy to use existing Ethereum tools (geth attach, web3) to interact with the node.


----


## Installation

### Binary Distribution
This installation involves the downloading of the binaries and subsequently moving them into your $PATH.

Download the [binaries](https://github.com/tendermint/ethermint/releases/tag/v0.2.2) for your platform on the release page. Open a terminal window and switch into the folder that you downloaded ethermint to.

```bash
unzip -d /usr/local/bin ${the name of the ethermint binary}.zip 
```

### Docker
We are currently building docker images for both [ethermint](https://hub.docker.com/r/adrianbrink/ethermint/) and [tendermint](https://hub.docker.com/r/adrianbrink/tendermint/). There are images for versioned releases, builds of master and builds of develop.

// TODO: Add example of how to get a node up and running using docker

### Source
Ethermint builds with go1.8.3 and hence go1.8.3 needs to be installed. In order to manage your go installation we recommend [GVM](https://github.com/moovweb/gvm).

Once go is installed and configured you can download ethermint.
```bash
go get -u -d github.com/tendermint/ethermint
```
This places the ethermint source code into the appropriate directory in your $GOPATH.

Next you need to switch into the ethermint directory and install it.
```bash
cd $GOPATH/src/github.com/tendermint/ethermint

make install
```
This will download all the necessary dependencies and install ethermint in your $PATH.


### Tendermint
Ethermint relies on the [Tendermint](https://github.com/tendermint/tendermint) executable to provide the networking and consensus engines.

```bash
go get -u -d github.com/tendermint/tendermint

cd $GOPATH/src/github.com/tendermint/tendermint

make install
```
This will download all the necessary dependencies and install tendermint in your $PATH.


----


## Starting Ethermint

### Initialisation
To get started, you need to initialise the genesis block for tendermint core and go-ethereum. We provide initialisation
files with reasonable defaults and money allocated into a predefined account. If you installed from binary or docker
please download these default files [here](https://github.com/tendermint/ethermint/tree/develop/setup).

You can choose where to store the ethermint files with `--datadir`. For this guide, we will use `~/.ethermint`, which is a reasonable default in most cases.

Before you can run ethermint you need to initialise tendermint and ethermint with their respective genesis states.
Please switch into the folder where you have the initialisation files. If you installed from source you can just follow
these instructions.
```bash
tendermint init --home ~/.ethermint/tendermint

cd $GOPATH/src/github.com/tendermint/ethermint

ethermint --datadir ~/.ethermint init setup/genesis.json

cp -r setup/keystore ~/.ethermint
```
In the last step we copy the private key from the initialisation folder into the actual ethereum folder. 


### Running
To execute ethermint we need to start two processes. The first one is for tendermint, which handles the P2P
communication as well as the consensus process, while the second one is actually ethermint, which provides the
go-ethereum functionality.

```bash
tendermint --home ~/.ethermint/tendermint node

ethermint --datadir ~/.ethermint --rpc --rpcaddr=0.0.0.0 --ws --wsaddr=0.0.0.0 --rpcapi eth,net,web3,personal,admin
```

The **password** for the default account is *1234*.


----


## Connecting to Ethermint
Ethermint is fully compatible with the standard go-ethereum tooling such as [geth](https://github.com/ethereum/go-ethereum/wiki/Geth), [mist](https://github.com/ethereum/mist) and [truffle](https://github.com/trufflesuite/truffle). Please
install whichever tooling suits you best. 

Below we are explaining how to connect these tools to ethermint. For installation instructions please visit the respective projects.

### Geth

```bash
geth attach http://localhost:8545
```
This will drop you into a web3 console.

### Mist
The mist tooling requires more setup, so be sure to read the explanations on the project site. Of course, you are 
welcome to ask for help in our [slack channel](https://img.shields.io/badge/slack-join%20chat-brightgreen.sv).
```bash
meteor --no-release-check

electron . --rpc http://localhost:8545
```

### Truffle
Truffle works without any extra arguments.
```bash
truffle init

truffle migrate
```

And you're off!


----


## Multi-Node Deployments
If you are interested in using ethermint in a multi-node setup (like a blockchain really, doh) you can use our kubernetes
scripts.

Please install [minikube and all its dependencies](https://github.com/kubernetes/minikube).

First you have to start minikube
```bash
minikube start
```

Starting a cluster of ethermint nodes goes like this
```bash
./scripts/kubernetes/start.sh 3 3 3
```
`start.sh N V S` will start *N* number of instances with *V* validator and *S* seed nodes.

To get an overview of all the running nodes use the minikube dashboard
```bash
minikube dashboard
```

In order to connect to a node using the tooling described above you have to expose it first
```bash
./scripts/kubernetes/expose.sh 1
```
`expose.sh N` exposes N number of nodes. N can't be greater than the number of nodes you started previously.
Minikube will assign a random port to every exposed node's services and hence you have to obtain that information as well.
```bash
minikube service tm-0 --format {{.IP}}:{{.Port}}
```
`tm-0` refers to the first ethermint node. The first IP:Port combination is for the go-ethereum RPC and the second one is for
the tendermint RPC.

To delete all nodes you can just run
```bash
./scripts/kubernetes/destroy.sh
```

Finally to shutdown kubernetes and minikube run
```bash
minikube stop
```

You might have wondered why you don't have to worry about starting tendermint and ethermint separately. Our scripts start
both processes at the same time and expose it as one node. That is also the reason why the expose commands exposes both processes.


---


## Run Ethereum Network Stats

The Ethereum network stats system consist of a frontend and a backend. In order for both of them to securely communicate
with each other you have to set the environment variable `WS_SECRET`. You can just use any random string and in this
example I will use `ethermint` as my `WS_SECRET`.

First you have to start Ethermint and Tendermint as described above and second you have to install two global dependencies.
```bash
npm install -g grunt-cli
```

To start the frontend you have to clone another repository, so please navigate to a suitable destination folder.
```bash
git clone https://github.com/cubedro/eth-netstats

cd eth-netstats

npm install

grunt

WS_SECRET=ethermint npm start
```
Now you can see the frontend running at [http://localhost:3000](http://localhost:3000), but no information will be present
since we haven't started the backend.

To start the backend you have to clone another repository again, so please navigate to a suitable destination folder.
```bash
git clone https://github.com/tendermint/eth-net-intelligence-api.git

cd eth-net-intelligence-api

npm install

WS_SECRET=ethermint node app.js
```

Now you should be able to navigate to [http://localhost:3000](http://localhost:3000) and see constant information about your
ethermint node streaming by.


----


## Contributing to Ethermint

**Thanks for considering making contributions to Ethermint!**

Please follow standard github best practices: fork the repo, **branch from the
tip of develop**, make some commits, test your code changes with `make test`,
and submit a pull request to develop.

See the [open issues](https://github.com/tendermint/ethermint/issues) for
things we need help with!

Please make sure to use `gofmt` before every commit - the easiest way to do
this is have your editor run it for you upon saving a file.
