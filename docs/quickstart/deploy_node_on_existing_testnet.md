<!--
order: 8
-->

# Deploy node to existing testnet

Learn how to deploy a node to Digital Ocean and connect to public testnet

## Pre-requisite Readings

- [Install Ethermint](./installation.md) {prereq}
- [Start Testnet](./testnet.md) {prereq}
<!-- - [Deploy Testnet to DigitalOcean]() {prereq} -->


## Deploy node to Digital Ocean

### Create a Digital Ocean account and Generate personal access token

Head over to https://www.digitalocean.com/ and create an account.

Click the API navigation.

In the Personal Access Tokens section, click Generate new token.

Give your token a name, and click the Generate Token button.

Once it’s created, copy it and create an environment variable.

### Create a Droplet

We will use the docker-machine command, which should already be installed on your development machine if you have been working with Docker.

::: tip
 Type docker-machine into a terminal, if you get an error you may need to install it yourself.
:::

Run command 
```bash 
docker-machine create --digitalocean-size "s-2vcpu-4gb" --driver digitalocean --digitalocean-access-token PERSONAL_ACCESS_TOKEN machine-name
```
--digitalocean-size "s-2vcpu-4gb" creates an image with 2 CPUs and 4 GB of RAM — you can use any size you feel is appropriate

machine-name is the name of our droplet

You can verify if it was created successfully with `docker-machine ls`

You should see something like
```bash
machine-name * digitalocean Running tcp://152.63.254.5:2376 v18.05.0-ce
```

### Deploy Node
Use docker-machine use command to select your remote machine: `docker-machine use machine-name`

Now execute these commands to run a container on that machine:
```bash
# build the image
docker build -t node1 .
```

#### Copy the Genesis File

To connect to an existing testnet, fetch the testnet's `genesis.json` file and copy it into the docker images config directory (i.e `$HOME/.ethermintd/config/genesis.json`).

```bash
docker cp node1:$HOME/.ethermintd/config/genesis.json $HOME/.ethermintd/config/genesis.json `
```

```bash
# run the container on the remote machine
docker run -d node1 ethermintd start
```