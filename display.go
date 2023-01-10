// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	nameTruncate   = 20 // Max size a Name tag can be before its truncated with a "..." at the end
	expungedIP     = "xxx.xxx.xxx.xxx"
	expungedDomain = "xxxx.xxxx.xxxx"
	expungedMAC    = "xx:xx:xx:xx:xx:xx"
	expungedCIDR   = "xxx.xxx.xxx.xxx/xx"
	expungedV6CIDR = "xxxx::xxxx/xx"
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

func formatName(name string) string {
	if name == "" {
		return ""
	}
	// Names can be up to 255 utf8 runes, we should truncate it
	runes := []rune(name)
	if Config.Truncate {
		if len(runes) > nameTruncate {
			runes = runes[:(nameTruncate - 1 - 3)]
			runes = append(runes, []rune("...")...)
		}
	}

	return fmt.Sprintf(" [%s]", string(runes))
}

func printRegionsJSON(regions []RegionDataSorted) {
	export, _ := json.Marshal(regions)
	fmt.Printf("%v", string(export))
}

func printVPCsJSON(vpcs []VPCSorted) {
	export, _ := json.Marshal(vpcs)
	fmt.Printf("%v", string(export))
}

func printVPCs(vpcs []VPCSorted) {
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

	// sort the keys
	for vpcIdx := range vpcs {
		vpc := vpcs[vpcIdx]
		// Print VPC
		fmt.Printf(
			"%v%v%v%v ",
			color.Green,
			vpc.ID,
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

		if Config.HideIP {
			vpc.CidrBlock = expungedCIDR
			if vpc.IPv6CidrBlock != "" {
				vpc.IPv6CidrBlock = expungedV6CIDR
			}
		}

		fmt.Printf(
			"%v %v -- ",
			vpc.CidrBlock,
			vpc.IPv6CidrBlock,
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

		fmt.Printf("\n")

		// Print Peers
		peersExist := false

		for peerIdx := range vpc.Peers {
			peer := vpc.Peers[peerIdx]
			direction := "peer-->"
			vpcOperand := peer.Accepter
			peersExist = true

			if peer.Accepter == vpc.ID {
				direction = "<--peer"
				vpcOperand = peer.Requester
			}

			fmt.Printf(
				"%s%v%v%v%v %v %v%v%v\n",
				indent(4),
				color.Cyan,
				peer.ID,
				formatName(peer.Name),
				color.Reset,
				direction,
				color.Green,
				vpcOperand,
				color.Reset,
			)
		}

		if peersExist {
			lineFeed()
		}

		// Print Subnets
		for subnetIdx := range vpc.Subnets {
			subnet := vpc.Subnets[subnetIdx]

			// Print Subnet Info
			public := "Private"
			if subnet.Public {
				public = "Public"
			}

			if Config.HideIP {
				subnet.CidrBlock = expungedCIDR
			}

			fmt.Printf(
				"%s%v%v%v%v  %v  %v %v-->%v%v %v\n",
				indent(4),
				color.Blue,
				subnet.ID,
				formatName(subnet.Name),
				color.Reset,
				subnet.AvailabilityZone,
				subnet.CidrBlock,
				color.Yellow,
				subnet.RouteTable.Default,
				color.Reset,
				public,
			)

			// Print Endpoints
			if Config.Verbose {
				for interfaceEndpointIdx := range subnet.InterfaceEndpoints {
					interfaceEndpoint := subnet.InterfaceEndpoints[interfaceEndpointIdx]

					fmt.Printf(
						"%s%v%v%v%v interface--> %v\n",
						indent(8),
						color.Cyan,
						interfaceEndpoint.ID,
						formatName(interfaceEndpoint.Name),
						color.Reset,
						interfaceEndpoint.ServiceName,
					)

					for ifaceIdx := range interfaceEndpoint.Interfaces {
						iface := interfaceEndpoint.Interfaces[ifaceIdx]
						// An endpoint can have multiple interfaces in multiple subnets, we only want to display whats relevant to the subnet
						if Config.HideIP {
							iface.MAC = expungedMAC
							iface.PublicIP = expungedIP
							iface.PrivateIP = expungedIP

							if iface.DNS != "" {
								iface.DNS = expungedDomain
							}
						}

						if iface.SubnetID == subnet.ID {
							fmt.Printf(
								"%s%v%v %v %v %v %v %v \n",
								indent(12),
								iface.ID,
								formatName(iface.Name),
								iface.Type,
								iface.MAC,
								iface.PublicIP,
								iface.PrivateIP,
								iface.DNS,
							)
						}
					}
				}
			}

			for gatewayEndpointIdx := range subnet.GatewayEndpoints {
				gatewayEndpoint := subnet.GatewayEndpoints[gatewayEndpointIdx]
				fmt.Printf(
					"%s%v%v%v%v gateway--> %v\n",
					indent(8),
					color.Cyan,
					gatewayEndpoint.ID,
					formatName(gatewayEndpoint.Name),
					color.Reset,
					gatewayEndpoint.ServiceName,
				)
			}

			// Print Interfaces
			for ifaceIdx := range subnet.ENIs {
				iface := subnet.ENIs[ifaceIdx]

				if Config.HideIP {
					iface.MAC = expungedMAC
					iface.PrivateIP = expungedIP
					iface.PublicIP = expungedIP

					if iface.DNS != "" {
						iface.DNS = expungedDomain
					}
				}

				fmt.Printf(
					"%s%v%v%v%v %v %v %v %v %v : %v\n",
					indent(8),
					color.Cyan,
					iface.ID,
					formatName(iface.Name),
					color.Reset,
					iface.Type,
					iface.MAC,
					iface.PublicIP,
					iface.PrivateIP,
					iface.DNS,
					iface.Description,
				)
			}
			// Print EC2 Instances
			for instanceIdx := range subnet.Instances {
				instance := subnet.Instances[instanceIdx]
				// Its too clunky to directly report SystemStatus and InstanceStatus, lets do it like the console does
				status := 0

				if instance.SystemStatus == "ok" {
					status++
				}

				if instance.InstanceStatus == "ok" {
					status++
				}

				// Print Instance Info
				if Config.HideIP {
					instance.PublicIP = expungedIP
					instance.PrivateIP = expungedIP
				}

				fmt.Printf(
					"%s%v%s%v%v %v -- %v (%v/2) -- %v -- %v\n",
					indent(8),
					color.Cyan,
					instance.ID,
					formatName(instance.Name),
					color.Reset,
					instance.Type,
					instance.State,
					status,
					instance.PublicIP,
					instance.PrivateIP,
				)

				// Print Instance Interfaces
				if Config.Verbose {
					for ifaceIdx := range instance.Interfaces {
						iface := instance.Interfaces[ifaceIdx]

						if Config.HideIP {
							iface.MAC = expungedMAC
							iface.PrivateIP = expungedIP
							iface.PublicIP = expungedIP

							if iface.DNS != "" {
								iface.DNS = expungedDomain
							}
						}

						fmt.Printf(
							"%s%v%v  %v  %v  %v\n",
							indent(12),
							iface.ID,
							formatName(iface.Name),
							iface.MAC,
							iface.PrivateIP,
							iface.DNS,
						)
					}

					// Print Instance Volumes
					for volumeIdx := range instance.Volumes {
						volume := instance.Volumes[volumeIdx]
						fmt.Printf(
							"%s%v%v  %v  %v  %v GiB\n",
							indent(12),
							volume.ID,
							formatName(volume.Name),
							volume.VolumeType,
							volume.DeviceName,
							volume.Size,
						)
					}
				}
			}

			// Print Nat Gateways
			for natGatewayIdx := range subnet.NatGateways {
				natGateway := subnet.NatGateways[natGatewayIdx]

				if Config.HideIP {
					natGateway.PrivateIP = expungedIP
					natGateway.PublicIP = expungedIP
				}

				fmt.Printf(
					"%s%v%v%v%v  %v  %v  %v  %v\n",
					indent(8),
					color.Cyan,
					natGateway.ID,
					formatName(natGateway.Name),
					color.Reset,
					natGateway.Type,
					natGateway.State,
					natGateway.PublicIP,
					natGateway.PrivateIP,
				)
			}

			// Print Transit Gateway Attachments
			for tgwIdx := range subnet.TGWs {
				tgw := subnet.TGWs[tgwIdx]
				fmt.Printf(
					"%s%v%v%v%v ---> %v%v%v\n",
					indent(8),
					color.Cyan,
					tgw.AttachmentID,
					formatName(tgw.Name),
					color.Reset,
					color.Yellow,
					tgw.TransitGatewayID,
					color.Reset,
				)
			}

			lineFeed()
		}
	}
}
