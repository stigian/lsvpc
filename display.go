// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

const nameTruncate = 20 //max size a Name tag can be before its truncated with a "..." at the end

func indent(num int) string {
	sb := strings.Builder{}

	for i := 0; i <= num; i++ {
		sb.WriteString(" ")
	}

	return sb.String()
}

type colorPalette struct {
	Reset  string
	Red    string
	Green  string
	Yellow string
	Blue   string
	Purple string
	Cyan   string
	White  string
}

func lineFeed() {
	if !Config.noSpace {
		fmt.Println()
	}
}

func formatName(name *string) string {
	if aws.StringValue(name) == "" {
		return ""
	}
	//Names can be up to 255 utf8 runes, we should truncate it
	runes := []rune(aws.StringValue(name))
	if len(runes) > nameTruncate {
		runes = runes[:(nameTruncate - 1 - 3)]
		runes = append(runes, []rune("...")...)
	}

	return fmt.Sprintf(" [%s]", string(runes))
}

func printSortedVPCs(vpcs []VPCSorted) {
	for _, vpc := range vpcs {
		fmt.Printf("%v [%v]\n", aws.StringValue(vpc.Id), aws.StringValue(vpc.Name))
		for _, subnet := range vpc.Subnets {
			fmt.Printf("    %v\n", aws.StringValue(subnet.Id))
			for _, nat := range subnet.NatGateways {
				fmt.Printf("      %v\n", aws.StringValue(nat.Id))
			}
			for _, tgw := range subnet.NatGateways {
				fmt.Printf("      %v\n", aws.StringValue(tgw.Id))
			}
			for _, eni := range subnet.ENIs {
				fmt.Printf("      %v\n", aws.StringValue(eni.Id))
			}
			for _, interfaceEndpoint := range subnet.InterfaceEndpoints {
				fmt.Printf("      %v\n", aws.StringValue(interfaceEndpoint.Id))
			}
			for _, gatewayEndpoint := range subnet.GatewayEndpoints {
				fmt.Printf("      %v\n", aws.StringValue(gatewayEndpoint.Id))
			}
			for _, instance := range subnet.Instances {
				fmt.Printf("      %v\n", aws.StringValue(instance.Id))
				for _, volume := range instance.Volumes {
					fmt.Printf("        %v\n", aws.StringValue(volume.Id))
				}
				for _, eni := range instance.Interfaces {
					fmt.Printf("        %v\n", aws.StringValue(eni.Id))
				}
			}
		}
	}
}

func printVPCs(in map[string]VPC) {
	vpcs := sortVPCs(in)
	color := colorPalette{}

	if !Config.noColor {
		color.Reset = "\033[0m"
		color.Red = "\033[31m"
		color.Green = "\033[32m"
		color.Yellow = "\033[33m"
		color.Blue = "\033[34m"
		color.Purple = "\033[35m"
		color.Cyan = "\033[36m"
		color.White = "\033[37m"
	}

	//sort the keys
	for _, vpc := range vpcs {
		// Print VPC
		fmt.Printf(
			"%v%v%v%v ",
			color.Green,
			aws.StringValue(vpc.Id),
			formatName(vpc.Name),
			color.Reset,
		)

		if vpc.IsDefault {
			fmt.Printf(
				"%v(default)%v ",
				color.Yellow,
				color.Reset,
			)
		}

		fmt.Printf(
			"%v %v -- ",
			aws.StringValue(vpc.CidrBlock),
			aws.StringValue(vpc.IPv6CidrBlock),
		)
		// Print gateways set to VPC
		for _, gateway := range vpc.Gateways {
			fmt.Printf(
				"%v%v%v ",
				color.Yellow,
				gateway,
				color.Reset,
			)
		}

		lineFeed()

		// Print Peers
		peersExist := false
		for _, peer := range vpc.Peers {
			direction := "peer-->"
			vpcOperand := aws.StringValue(peer.Accepter)
			if aws.StringValue(peer.Accepter) == aws.StringValue(vpc.Id) {
				direction = "<--peer"
				vpcOperand = aws.StringValue(peer.Requester)
			}
			fmt.Printf(
				"%s%v%v%v%v %v %v%v%v\n",
				indent(4),
				color.Cyan,
				aws.StringValue(peer.Id),
				formatName(peer.Name),
				color.Reset,
				direction,
				color.Green,
				vpcOperand,
				color.Reset,
			)
			peersExist = true
		}
		if peersExist {
			lineFeed()
		}

		// Print Subnets
		for _, subnet := range vpc.Subnets {

			// Print Subnet Info
			public := "Private"
			if subnet.Public {
				public = "Public"
			}
			fmt.Printf(
				"%s%v%v%v%v  %v  %v %v-->%v%v %v\n",
				indent(4),
				color.Blue,
				aws.StringValue(subnet.Id),
				formatName(subnet.Name),
				color.Reset,
				aws.StringValue(subnet.AvailabilityZone),
				aws.StringValue(subnet.CidrBlock),
				color.Yellow,
				aws.StringValue(subnet.RouteTable.Default),
				color.Reset,
				public,
			)

			//Print Endpoints
			for _, interfaceEndpoint := range subnet.InterfaceEndpoints {
				fmt.Printf(
					"%s%v%v%v%v interface--> %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(interfaceEndpoint.Id),
					formatName(interfaceEndpoint.Name),
					color.Reset,
					aws.StringValue(interfaceEndpoint.ServiceName),
				)
			}

			for _, gatewayEndpoint := range subnet.GatewayEndpoints {
				fmt.Printf(
					"%s%v%v%v%v gateway--> %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(gatewayEndpoint.Id),
					formatName(gatewayEndpoint.Name),
					color.Reset,
					aws.StringValue(gatewayEndpoint.ServiceName),
				)
			}

			// Print Interfaces
			for _, iface := range subnet.ENIs {
				fmt.Printf(
					"%s%v%v%v%v %v %v %v %v %v : %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(iface.Id),
					formatName(iface.Name),
					color.Reset,
					aws.StringValue(iface.Type),
					aws.StringValue(iface.MAC),
					aws.StringValue(iface.PublicIp),
					aws.StringValue(iface.PrivateIp),
					aws.StringValue(iface.DNS),
					aws.StringValue(iface.Description),
				)
			}
			// Print EC2 Instances
			for _, instance := range subnet.Instances {
				// Its too clunky to directly report SystemStatus and InstanceStatus, lets do it like the console does
				status := 0
				if aws.StringValue(instance.SystemStatus) == "ok" {
					status = status + 1
				}
				if aws.StringValue(instance.InstanceStatus) == "ok" {
					status = status + 1
				}

				// Print Instance Info
				fmt.Printf(
					"%s%v%s%v%v %v -- %v (%v/2) -- %v -- %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(instance.Id),
					formatName(instance.Name),
					color.Reset,
					aws.StringValue(instance.Type),
					aws.StringValue(instance.State),
					status,
					aws.StringValue(instance.PublicIP),
					aws.StringValue(instance.PrivateIP),
				)

				// Print Instance Interfaces
				for _, iface := range instance.Interfaces {
					fmt.Printf(
						"%s%v%v  %v  %v  %v\n",
						indent(12),
						aws.StringValue(iface.Id),
						formatName(iface.Name),
						aws.StringValue(iface.MAC),
						aws.StringValue(iface.PrivateIp),
						aws.StringValue(iface.DNS),
					)
				}

				// Print Instance Volumes
				for _, volume := range instance.Volumes {
					fmt.Printf(
						"%s%v%v  %v  %v  %v GiB\n",
						indent(12),
						aws.StringValue(volume.Id),
						formatName(volume.Name),
						aws.StringValue(volume.VolumeType),
						aws.StringValue(volume.DeviceName),
						aws.Int64Value(volume.Size),
					)
				}
			}

			//Print Nat Gateways
			for _, natGateway := range subnet.NatGateways {
				fmt.Printf(
					"%s%v%v%v%v  %v  %v  %v  %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(natGateway.Id),
					formatName(natGateway.Name),
					color.Reset,
					aws.StringValue(natGateway.Type),
					aws.StringValue(natGateway.State),
					aws.StringValue(natGateway.PublicIP),
					aws.StringValue(natGateway.PrivateIP),
				)
			}

			//Print Transit Gateway Attachments
			for _, tgw := range subnet.TGWs {
				fmt.Printf(
					"%s%v%v%v%v ---> %v%v%v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(tgw.AttachmentId),
					formatName(tgw.Name),
					color.Reset,
					color.Yellow,
					aws.StringValue(tgw.TransitGatewayId),
					color.Reset,
				)
			}

			lineFeed()
		}
	}
}
