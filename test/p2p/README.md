Docker Compose based load runner for Ethereum / Ethermint

### Pre Requirements
docker images are created

in `root` directory
```
  docker build -t docker.ethermint -f .travis.dockerfile .
```

in `test/contract`
```
  docker build -t docker.contract .
```

folders are inititalized
```
mkdir dummy
docker run -it --rm -v `pwd`/dummy:/data docker.ethermint init /go/src/github.com/tendermint/ethermint/dev/genesis.json
cp -r dummy/chaindata mach1/
cp -r dummy/chaindata mach2/
cp -r dummy/chaindata mach3/
cp -r dummy/chaindata mach4/
cp -r dummy/chaindata gateway/
```

### Run
```
docker-compose up -d mach1 mach2 mach3 mach4 gateway
docker-compose logs --tail 100
docker-compose run contract node deploy.js
docker-compose run contract node main.js
docker-compose stop
docker-compose logs --tail 100
docker-compose rm -f
```
