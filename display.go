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

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func printVPCs(vpcs map[string]VPC) {
	//sort the keys
	vpcKeys := []string{}
	for k, _ := range vpcs {
		vpcKeys = append(vpcKeys, k)
	}

	sort.Strings(vpcKeys)

	for _, vpcId := range vpcKeys {
		vpc := vpcs[vpcId]

		// Print VPC
		fmt.Printf(
			"%v%v%v ",
			string(colorGreen),
			aws.StringValue(vpc.Id),
			string(colorReset),
		)

		if vpc.IsDefault {
			fmt.Printf(
				"%v(default)%v ",
				string(colorYellow),
				string(colorReset),
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
				string(colorYellow),
				gateway,
				string(colorReset),
			)
		}

		fmt.Printf("\n")

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
				"%s%v%v%v %v %v%v%v\n",
				indent(4),
				string(colorCyan),
				aws.StringValue(peer.Id),
				string(colorReset),
				direction,
				string(colorGreen),
				vpcOperand,
				string(colorReset),
			)
			peersExist = true
		}
		if peersExist {
			fmt.Println()
		}

		// Print Subnets
		subnetKeys := []string{}

		for k, _ := range vpc.Subnets {
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
				string(colorBlue),
				aws.StringValue(subnet.Id),
				string(colorReset),
				aws.StringValue(subnet.AvailabilityZone),
				aws.StringValue(subnet.CidrBlock),
				string(colorYellow),
				aws.StringValue(subnet.RouteTable.Default),
				string(colorReset),
				public,
			)

			//Print Endpoints
			for _, interfaceEndpoint := range subnet.InterfaceEndpoints {
				fmt.Printf(
					"%s%v%v%v interface--> %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(interfaceEndpoint.Id),
					string(colorReset),
					aws.StringValue(interfaceEndpoint.ServiceName),
				)
			}

			for _, gatewayEndpoint := range subnet.GatewayEndpoints {
				fmt.Printf(
					"%s%v%v%v gateway--> %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(gatewayEndpoint.Id),
					string(colorReset),
					aws.StringValue(gatewayEndpoint.ServiceName),
				)
			}

			// Print Interfaces
			for _, iface := range subnet.ENIs {
				fmt.Printf(
					"%s%v%v%v %v %v %v %v %v : %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(iface.Id),
					string(colorReset),
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

			for k, _ := range subnet.EC2s {
				instanceKeys = append(instanceKeys, k)
			}

			sort.Strings(instanceKeys)

			for _, instanceId := range instanceKeys {
				instance := subnet.EC2s[instanceId]

				// Print Instance Info
				fmt.Printf(
					"%s%v%s%v -- %v -- %v -- %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(instance.Id),
					string(colorReset),
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
			for _, natGateway := range subnet.NatGateways {
				fmt.Printf(
					"%s%v%v%v  %v  %v  %v  %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(natGateway.Id),
					string(colorReset),
					aws.StringValue(natGateway.Type),
					aws.StringValue(natGateway.State),
					aws.StringValue(natGateway.PublicIP),
					aws.StringValue(natGateway.PrivateIP),
				)
			}

			//Print Transit Gateway Attachments
			for _, tgw := range subnet.TGWs {
				fmt.Printf(
					"%s%v%v%v ---> %v%v%v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(tgw.AttachmentId),
					string(colorReset),
					string(colorYellow),
					aws.StringValue(tgw.TransitGatewayId),
					string(colorReset),
				)
			}

			fmt.Printf("\n")
		}
	}
}
