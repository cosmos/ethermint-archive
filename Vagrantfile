# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  version = "1.8"
  locale = "en_GB.UTF.8"

  config.vm.box = "bento/ubuntu-16.04"

  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  # setup
  config.vm.provision :shell, :inline => "apt-get update --fix-missing"
  config.vm.provision :shell, :inline => "apt-get install -y --no-install-recommends curl"

  # go-lang
  config.vm.provision :shell, :inline => "curl -O https://storage.googleapis.com/golang/go#{version}.linux-amd64.tar.gz"
  config.vm.provision :shell, :inline => "tar -xvf go#{version}.linux-amd64.tar.gz"
  config.vm.provision :shell, :inline => "rm -rf /usr/local/go"
  config.vm.provision :shell, :inline => "mv go /usr/local"
  config.vm.provision :shell, :inline => "rm -f go#{version}.linux-amd64.tar.gz"
  config.vm.provision :shell, :inline => "mkdir -p /home/vagrant/go/bin"
  config.vm.provision :shell, :inline => "echo 'export PATH=$PATH:/usr/local/go/bin:/home/vagrant/go/bin' >> /home/vagrant/.bash_profile"
  config.vm.provision :shell, :inline => "echo 'export GOPATH=/home/vagrant/go' >> /home/vagrant/.bash_profile"
  config.vm.provision :shell, :inline => "echo 'export LC_ALL=#{locale}' >> /home/vagrant/.bash_profile"

  # setup
  config.vm.provision :shell, :inline => "apt-get autoremove -y"

  # ethermint
  config.vm.provision :shell, :inline => "mkdir -p /home/vagrant/go/src/github.com/tendermint"
  config.vm.provision :shell, :inline => "ln -s /vagrant /home/vagrant/go/src/github.com/tendermint/ethermint"
  config.vm.provision :shell, :inline => "chown -R vagrant:vagrant /home/vagrant/go"
  config.vm.provision :shell, :inline => "su - vagrant -c 'cd /home/vagrant/go/src/github.com/tendermint/ethermint && make get_vendor_deps'"
end