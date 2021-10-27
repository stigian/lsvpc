# vpc-ls
A simple AWS VPC listing tool to provide quick introspection on the makeup of a VPC


# Installation and running

## From Source

using your OS maintainer's version of go:

```
sudo make install
```

which will install the executable into `/usr/local/bin/`.To install in a different directory, such as `/usr/bin/`, simply override the `INSTALL` variable:

```
sudo INSTALL=/usr/bin/ make install
```

If go has been installed using the [tarball](https://golang.org/doc/install), the `go` binary is
probably not in the sudoers `secure_path`, and the path variable will need to be overridden when invoking sudo:

```
sudo env "PATH=$PATH" make install
```

which will install the `lsvpc` binary into `/usr/local/bin/`


## Fetching with `go get`

Alternatively, you don't need to actually clone this source, and with golang installed, you may simply call:

```bash
go get github.com/tjames-stig/lsvpc
sudo GOBIN=/usr/local/bin/ go install github.com/tjames-stig/lsvpc
```

to install `lsvpc` into `/usr/local/bin/`. `GOBIN` can be set to be the value of where you want the executable to be installed

Or you may simply tell go to run the binary using the repository path:

```
go run github.com/tjames-stig/lsvpc
```


# Usage

## Configuration and Permissions
This tool only uses [named profiles](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-profiles.html) for authentication.
Be sure to `export AWS_PROFILE=<profile_name>` before executing this tool so that it can access aws credentials.

Below are all of the SDK actions this tool uses, be sure that your aws credentials has IAM permissions for them:
```
ec2:DescribeEgressOnlyInternetGateways
ec2:DescribeInstances
ec2:DescribeInternetGateways
ec2:DescribeNatGateways
ec2:DescribeNetworkInterfaces
ec2:DescribeRegions
ec2:DescribeRouteTables
ec2:DescribeSubnets
ec2:DescribeTransitGatewayVpcAttachments
ec2:DescribeVolumes
ec2:DescribeVpcEndpoints
ec2:DescribeVpcPeeringConnections
ec2:DescribeVpcs
ec2:DescribeVpnGateways
```

## Execution

Executing `lsvpc` with no arguments produces a colored readout of vpc resources detected in the default region of your aws profile

### Parameters

`-a, -all`    - Prints data for all regions in account

`-nocolor`    - Suppresses color output. In general, lsvpc will also supress color if its output is piped

`-color`      - Force color output. Overrides nocolor, and will print color even if lsvpc's output is being sent through a pipe

`-nospace`    - Suppresses line spacing between entries.

`-r, -region` - Specify a region to print data for.