// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"
	"sort"
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
		}
	}
}

func printVPCs(vpcs map[string]VPC) {
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
	vpcKeys := []string{}
	for k := range vpcs {
		vpcKeys = append(vpcKeys, k)
	}
	sort.Strings(vpcKeys)
	for _, vpcId := range vpcKeys {
		vpc := vpcs[vpcId]

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
		peerKeys := []string{}
		for k := range vpc.Peers {
			peerKeys = append(peerKeys, k)
		}
		sort.Strings(peerKeys)
		for _, peerId := range peerKeys {
			peer := vpc.Peers[peerId]
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
		subnetKeys := []string{}
		for k := range vpc.Subnets {
			subnetKeys = append(subnetKeys, k)
		}
		sort.Strings(subnetKeys)
		for _, subnetId := range subnetKeys {

			subnet := vpc.Subnets[subnetId]

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
			interfaceEndpointKeys := []string{}
			for k := range subnet.InterfaceEndpoints {
				interfaceEndpointKeys = append(interfaceEndpointKeys, k)
			}
			sort.Strings(interfaceEndpointKeys)
			for _, interfaceEndpointId := range interfaceEndpointKeys {
				interfaceEndpoint := subnet.InterfaceEndpoints[interfaceEndpointId]
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

			gatewayKeys := []string{}
			for k := range subnet.GatewayEndpoints {
				gatewayKeys = append(gatewayKeys, k)
			}
			sort.Strings(gatewayKeys)
			for _, gatewayEndpointId := range gatewayKeys {
				gatewayEndpoint := subnet.GatewayEndpoints[gatewayEndpointId]
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
			interfaceKeys := []string{}
			for k := range subnet.ENIs {
				interfaceKeys = append(interfaceKeys, k)
			}
			sort.Strings(interfaceKeys)
			for _, interfaceId := range interfaceKeys {
				iface := subnet.ENIs[interfaceId]
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
			instanceKeys := []string{}
			for k := range subnet.Instances {
				instanceKeys = append(instanceKeys, k)
			}
			sort.Strings(instanceKeys)
			for _, instanceId := range instanceKeys {
				instance := subnet.Instances[instanceId]

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
				instanceInterfaceKeys := []string{}
				for k := range instance.Interfaces {
					instanceInterfaceKeys = append(instanceInterfaceKeys, k)
				}
				sort.Strings(instanceInterfaceKeys)
				for _, interfaceId := range instanceInterfaceKeys {
					iface := instance.Interfaces[interfaceId]
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
			natGatewayKeys := []string{}
			for k := range subnet.NatGateways {
				natGatewayKeys = append(natGatewayKeys, k)
			}
			sort.Strings(natGatewayKeys)
			for _, natGatewayId := range natGatewayKeys {
				natGateway := subnet.NatGateways[natGatewayId]
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
			tgwKeys := []string{}
			for k := range subnet.TGWs {
				tgwKeys = append(tgwKeys, k)
			}
			sort.Strings(tgwKeys)
			for _, tgwId := range tgwKeys {
				tgw := subnet.TGWs[tgwId]
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
