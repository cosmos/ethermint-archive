Kuberenetes
===

### Minikube
`./start.sh N V S` will start N number of nodes, with V validators and S seed nodes.

if you want to expose nodes, you can use `./expose.sh N` which will expose N number of nodes from running pods.  
It will assign random ports to each services.
You can get these ports by running `minikube service tm-0 --url`, it will list links or `minikube service tm-0 --format {{.IP}}:{{.Port}}` where `tm-0` is first node. First IP/Port will be for GethRPC and second for Tendermint RPC
