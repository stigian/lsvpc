package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

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
			"%v%v%v ",
			color.Green,
			aws.StringValue(vpc.Id),
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
				"%s%v%v%v %v %v%v%v\n",
				indent(4),
				color.Cyan,
				aws.StringValue(peer.Id),
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
				"%s%v%v%v  %v  %v %v-->%v%v %v\n",
				indent(4),
				color.Blue,
				aws.StringValue(subnet.Id),
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
					"%s%v%v%v interface--> %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(interfaceEndpoint.Id),
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
					"%s%v%v%v gateway--> %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(gatewayEndpoint.Id),
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
					"%s%v%v%v %v %v %v %v %v : %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(iface.Id),
					color.Reset,
					aws.StringValue(iface.Type),
					aws.StringValue(iface.MAC),
					aws.StringValue(iface.PublicIp),
					aws.StringValue(iface.PrivateIp),
					aws.StringValue(iface.DNS),
					aws.StringValue(iface.Description),
				)
			}
			// Print EC2 Instance
			instanceKeys := []string{}
			for k := range subnet.EC2s {
				instanceKeys = append(instanceKeys, k)
			}
			sort.Strings(instanceKeys)
			for _, instanceId := range instanceKeys {
				instance := subnet.EC2s[instanceId]

				// Print Instance Info
				fmt.Printf(
					"%s%v%s%v -- %v -- %v -- %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(instance.Id),
					color.Reset,
					aws.StringValue(instance.State),
					aws.StringValue(instance.PublicIP),
					aws.StringValue(instance.PrivateIP),
				)

				// Print Instance Volumes
				for _, iface := range instance.Interfaces {
					fmt.Printf(
						"%s%v  %v  %v  %v\n",
						indent(12),
						aws.StringValue(iface.Id),
						aws.StringValue(iface.MAC),
						aws.StringValue(iface.PrivateIp),
						aws.StringValue(iface.DNS),
					)
				}

				// Print Instance Volumes
				for _, volume := range instance.Volumes {
					fmt.Printf(
						"%s%v  %v  %v  %v GiB\n",
						indent(12),
						aws.StringValue(volume.Id),
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
					"%s%v%v%v  %v  %v  %v  %v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(natGateway.Id),
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
					"%s%v%v%v ---> %v%v%v\n",
					indent(8),
					color.Cyan,
					aws.StringValue(tgw.AttachmentId),
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
