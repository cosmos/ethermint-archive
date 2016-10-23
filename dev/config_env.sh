export GOPATH=$(readlink --canonicalize $HOME/tendermint/gopath)
export PATH=$GOPATH/bin:$PATH
echo "Changing GOPATH to $GOPATH"
