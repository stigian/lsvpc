# vpc-ls
A simple AWS VPC listing tool to provide quick introspection on the makeup of a VPC


## Installation and running

### From Source

using your OS maintainer's version of go:

`sudo make install`

which will install the executable into `/usr/local/bin/`.To install in a different directory, such as `/usr/bin/`, simply override the `INSTALL` variable:

`sudo INSTALL=/usr/bin/ make install`

If go has been installed using the [tarball](https://golang.org/doc/install), the `go` binary is
probably not in the sudoers `secure_path`, and the path variable will need to be overridden when invoking sudo:

`sudo env "PATH=$PATH" make install`

which will install the `lsvpc` binary into `/usr/local/bin/`


### Fetching with go get

Alternatively, you don't need to actually clone this source, and with golang installed, you may simply call:

```bash
go get github.com/tjames-stig/lsvpc
sudo GOBIN=/usr/local/bin/ go install github.com/tjames-stig/lsvpc
```

to install `lsvpc` into `/usr/local/bin/`. `GOBIN` can be set to be the value of where you want the executable to be installed

Or you may simply tell go to run the binary using the repository path:

`go run github.com/tjames-stig/lsvpc`
