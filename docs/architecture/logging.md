# Logging

Ethermint has to wrestle with two separate logging infrastructures. go-ethereum uses one type of logger
while the abci server uses a different logger. At the moment ethermint just instantiates two separate 
logger objects and passes them to their respective users. In the future it would be great to have an 
adapter that unifies both.
