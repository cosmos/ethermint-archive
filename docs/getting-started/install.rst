Install Ethermint
=================

Via Package Manager
--------------------

Warning: Installation via package manager is still a work in progress;
the following may not yet work. If that's the case, install from source.

Note: these commands will also install ``tendermint`` as a required binary.

Linux
~~~~~

::

    apt-get install ethermint

MacOS
~~~~~

::

    brew install ethermint

Windows
~~~~~~~

::

    choco install ethermint


Download Binary
---------------

See the `releases page <https://github.com/tendermint/ethermint/releases>`__ to download binaries for each platform.


From Source
-----------

On all platforms, you'll need ``golang`` `installed <https://golang.org/doc/install>`__. Then, install ``ethermint``:

::

    go get -u -d github.com/tendermint/ethermint
    cd $GOPATH/src/github.com/tendermint/ethermint
    make get_vendor_deps
    make install

followed by ``tendermint``:

::

    go get -u -d github.com/tendermint/tendermint
    cd $GOPATH/src/github.com/tendermint/tendermint
    make get_vendor_deps
    make install


Hang tight, the build may take awhile. A tool called ``glide`` will also be installed to manage the dependencies.
