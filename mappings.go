// Copyright 2023 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getNameTag(tags []types.Tag) string {
	var name string

	for _, tag := range tags {
		if aws.ToString(tag.Key) == "Name" {
			name = aws.ToString(tag.Value)
		}
	}

	return name
}

func mapVpcs(vpcs map[string]*VPC, vpcData []types.Vpc) {
	for _, v := range vpcData {
		var v6cidr string

		if v.Ipv6CidrBlockAssociationSet != nil {
			for _, assoc := range v.Ipv6CidrBlockAssociationSet {
				if string(assoc.Ipv6CidrBlockState.State) == "associated" {
					v6cidr = aws.ToString(assoc.Ipv6CidrBlock)
				}
			}
		}

		vpcs[aws.ToString(v.VpcId)] = &VPC{
			VPCData: VPCData{
				ID:            aws.ToString(v.VpcId),
				IsDefault:     aws.ToBool(v.IsDefault),
				CidrBlock:     aws.ToString(v.CidrBlock),
				IPv6CidrBlock: v6cidr,
				Name:          getNameTag(v.Tags),
			},
			RawVPC:  v,
			Subnets: make(map[string]*Subnet),
			Peers:   make(map[string]*VPCPeer),
		}
	}
}

func mapSubnets(vpcs map[string]*VPC, subnets []types.Subnet) {
	for _, v := range subnets {
		isPublic := aws.ToBool(v.MapCustomerOwnedIpOnLaunch) || aws.ToBool(v.MapPublicIpOnLaunch)

		vpcs[aws.ToString(v.VpcId)].Subnets[aws.ToString(v.SubnetId)] = &Subnet{
			SubnetData: SubnetData{
				ID:                 aws.ToString(v.SubnetId),
				CidrBlock:          aws.ToString(v.CidrBlock),
				AvailabilityZone:   aws.ToString(v.AvailabilityZone),
				AvailabilityZoneID: aws.ToString(v.AvailabilityZoneId),
				Name:               getNameTag(v.Tags),
				Public:             isPublic,
			},
			RawSubnet:          v,
			Instances:          make(map[string]*Instance),
			NatGateways:        make(map[string]*NatGateway),
			TGWs:               make(map[string]*TGWAttachment),
			ENIs:               make(map[string]*NetworkInterface),
			InterfaceEndpoints: make(map[string]*InterfaceEndpoint),
			GatewayEndpoints:   make(map[string]*GatewayEndpoint),
		}
	}
}

func mapInstances(vpcs map[string]*VPC, reservations []types.Reservation) {
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if instance.State.Name != types.InstanceStateNameTerminated {
				vpcID := aws.ToString(instance.VpcId)
				subnetID := aws.ToString(instance.SubnetId)
				instanceID := aws.ToString(instance.InstanceId)

				if vpcID != "" && subnetID != "" && instanceID != "" {
					vpcs[vpcID].Subnets[subnetID].Instances[instanceID] = &Instance{
						InstanceData: InstanceData{
							ID:           aws.ToString(instance.InstanceId),
							Type:         string(instance.InstanceType),
							SubnetID:     aws.ToString(instance.SubnetId),
							VpcID:        aws.ToString(instance.VpcId),
							State:        string(instance.State.Name),
							PublicIP:     aws.ToString(instance.PublicIpAddress),
							PrivateIP:    aws.ToString(instance.PrivateIpAddress),
							Name:         getNameTag(instance.Tags),
							PlatformName: aws.ToString(instance.PlatformDetails),
							PlatformType: string(instance.Platform),
						},
						RawEc2:     instance,
						Volumes:    make(map[string]*Volume),
						Interfaces: make(map[string]*NetworkInterface),
					}
				}
			}
		}
	}
}

func mapInstanceStatuses(vpcs map[string]*VPC, statuses []types.InstanceStatus) {
	for _, status := range statuses {
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for instanceID, instance := range subnet.Instances {
					if aws.ToString(status.InstanceId) == instanceID {
						updatedInstance := instance
						updatedInstance.InstanceStatus = string(status.InstanceStatus.Status)
						updatedInstance.SystemStatus = string(status.SystemStatus.Status)
						vpcs[vpcID].
							Subnets[subnetID].
							Instances[instanceID] = updatedInstance
					}
				}
			}
		}
	}
}

func mapVolumes(vpcs map[string]*VPC, volumes []types.Volume) {
	for _, volume := range volumes {
		for _, attachment := range volume.Attachments {
			if volInstanceID := aws.ToString(attachment.InstanceId); volInstanceID != "" {
				for vpcID, vpc := range vpcs {
					for subnetID, subnet := range vpc.Subnets {
						for instanceID := range subnet.Instances {
							if volInstanceID == instanceID {
								vpcs[vpcID].
									Subnets[subnetID].
									Instances[instanceID].
									Volumes[aws.ToString(volume.VolumeId)] = &Volume{
									ID:         aws.ToString(volume.VolumeId),
									DeviceName: aws.ToString(attachment.Device),
									Size:       aws.ToInt32(volume.Size),
									VolumeType: string(volume.VolumeType),
									Encrypted:  aws.ToBool(volume.Encrypted),
									KMSKeyID:   aws.ToString(volume.KmsKeyId),
									RawVolume:  volume,
									Name:       getNameTag(volume.Tags),
								}
							}
						}
					}
				}
			}
		}
	}
}

func mapNatGateways(vpcs map[string]*VPC, natGateways []types.NatGateway) {
	for _, gateway := range natGateways {
		if string(gateway.State) == "deleted" {
			continue
		}

		vpcs[aws.ToString(gateway.VpcId)].Subnets[aws.ToString(gateway.SubnetId)].NatGateways[aws.ToString(gateway.NatGatewayId)] = &NatGateway{
			NatGatewayData: NatGatewayData{
				ID:            aws.ToString(gateway.NatGatewayId),
				PrivateIP:     aws.ToString(gateway.NatGatewayAddresses[0].PrivateIp),
				PublicIP:      aws.ToString(gateway.NatGatewayAddresses[0].PublicIp),
				State:         string(gateway.State),
				Type:          string(gateway.ConnectivityType),
				Name:          getNameTag(gateway.Tags),
				RawNatGateway: gateway,
			},
			Interfaces: make(map[string]*NetworkInterface),
		}
	}
}

func getDefaultRoute(rtb types.RouteTable) string {
	for _, route := range rtb.Routes {
		if !(aws.ToString(route.DestinationCidrBlock) == "0.0.0.0/0" ||
			aws.ToString(route.DestinationIpv6CidrBlock) == "::/0") {
			continue
		}

		if dest := aws.ToString(route.CarrierGatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.EgressOnlyInternetGatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.GatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.InstanceId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.LocalGatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.NatGatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.NetworkInterfaceId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.TransitGatewayId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.VpcPeeringConnectionId); dest != "" {
			return dest
		}

		if dest := aws.ToString(route.CoreNetworkArn); dest != "" {
			return dest
		}
	}

	return "" // No default route found, which doesn't necessarily mean an error
}

func mapRouteTables(vpcs map[string]*VPC, routeTables []types.RouteTable) {
	// AWS doesn't actually have explicit queryable associations of route
	// tables to subnets. if no other route tables say they are associated
	// with a subnet, then that subnet is assumed to be on the default route table.
	// You can't determine this by looking at the subnets themselves, you
	// have to instead look at all route tables and pick out the ones
	// that say they are associated with particular subnets, and the
	// default route table doesn't even say which subnets they are
	// associated with.
	//
	// first pass, associate the default route with everything
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			if association.Main != nil && aws.ToBool(association.Main) {
				for subnetID := range vpcs[aws.ToString(routeTable.VpcId)].Subnets {
					subnet := vpcs[aws.ToString(routeTable.VpcId)].Subnets[subnetID]
					defaultRoute := getDefaultRoute(routeTable)
					subnet.RouteTable = &RouteTable{
						ID:       aws.ToString(routeTable.RouteTableId),
						Default:  defaultRoute,
						RawRoute: routeTable,
					}
					vpcs[aws.ToString(routeTable.VpcId)].Subnets[subnetID] = subnet
				}
			}
		}
	}

	// second pass, look at each route table's associations and assign them
	// to their explicitly mentioned subnet
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			// default route doesn't have subnet ids and will cause a nil dereference
			if string(association.AssociationState.State) != "associated" ||
				aws.ToBool(association.Main) {
				continue
			}

			subnet := vpcs[aws.ToString(routeTable.VpcId)].Subnets[aws.ToString(association.SubnetId)]
			defaultRoute := getDefaultRoute(routeTable)
			subnet.RouteTable = &RouteTable{
				ID:       aws.ToString(routeTable.RouteTableId),
				Default:  defaultRoute,
				RawRoute: routeTable,
			}
			vpcs[aws.ToString(routeTable.VpcId)].Subnets[aws.ToString(association.SubnetId)] = subnet
		}
	}
}

func mapInternetGateways(vpcs map[string]*VPC, internetGateways []types.InternetGateway) {
	for _, igw := range internetGateways {
		for _, attachment := range igw.Attachments {
			if vpcID := aws.ToString(attachment.VpcId); vpcID != "" {
				vpc := vpcs[vpcID]
				vpc.Gateways = append(vpc.Gateways, aws.ToString(igw.InternetGatewayId))
				vpcs[vpcID] = vpc
			}
		}
	}
}

func mapEgressOnlyInternetGateways(vpcs map[string]*VPC, eOIGWs []types.EgressOnlyInternetGateway) {
	for _, eoigw := range eOIGWs {
		for _, attach := range eoigw.Attachments {
			if string(attach.State) == "attached" {
				vpc := vpcs[aws.ToString(attach.VpcId)]
				vpc.Gateways = append(vpc.Gateways, aws.ToString(eoigw.EgressOnlyInternetGatewayId))
				vpcs[aws.ToString(attach.VpcId)] = vpc
			}
		}
	}
}

func mapVPNGateways(vpcs map[string]*VPC, vpnGateways []types.VpnGateway) {
	for _, vpgw := range vpnGateways {
		for _, attach := range vpgw.VpcAttachments {
			if string(attach.State) == "attached" {
				vpc := vpcs[aws.ToString(attach.VpcId)]
				vpc.Gateways = append(vpc.Gateways, aws.ToString(vpgw.VpnGatewayId))
				vpcs[aws.ToString(attach.VpcId)] = vpc
			}
		}
	}
}

func mapTransitGatewayVpcAttachments(vpcs map[string]*VPC, transitGatewayVpcAttachments []types.TransitGatewayVpcAttachment, identity *sts.GetCallerIdentityOutput) {
	for _, tgwatt := range transitGatewayVpcAttachments {
		// Transit Gateway vpc attachments are reported for external accounts too, need to omit those to fit in this data model
		if aws.ToString(tgwatt.VpcOwnerId) == aws.ToString(identity.Account) {
			if vpcID := aws.ToString(tgwatt.VpcId); vpcID != "" {
				for _, subnet := range tgwatt.SubnetIds {
					if subnetID := subnet; subnetID != "" {
						vpcs[vpcID].Subnets[subnetID].TGWs[aws.ToString(tgwatt.TransitGatewayAttachmentId)] = &TGWAttachment{
							AttachmentID:     aws.ToString(tgwatt.TransitGatewayAttachmentId),
							TransitGatewayID: aws.ToString(tgwatt.TransitGatewayId),
							Name:             getNameTag(tgwatt.Tags),
							RawAttachment:    tgwatt,
						}
					}
				}
			}
		}
	}
}

func mapVpcPeeringConnections(vpcs map[string]*VPC, vpcPeeringConnections []types.VpcPeeringConnection) {
	for _, peer := range vpcPeeringConnections {
		if string(peer.Status.Code) != "active" {
			continue
		}

		if requester := aws.ToString(peer.RequesterVpcInfo.VpcId); requester != "" {
			if _, ok := vpcs[requester]; ok {
				vpcs[requester].Peers[aws.ToString(peer.VpcPeeringConnectionId)] = &VPCPeer{
					ID:        aws.ToString(peer.VpcPeeringConnectionId),
					Requester: aws.ToString(peer.RequesterVpcInfo.VpcId),
					Accepter:  aws.ToString(peer.AccepterVpcInfo.VpcId),
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}

		if accepter := aws.ToString(peer.AccepterVpcInfo.VpcId); accepter != "" {
			if _, ok := vpcs[accepter]; ok {
				vpcs[accepter].Peers[aws.ToString(peer.VpcPeeringConnectionId)] = &VPCPeer{
					ID:        aws.ToString(peer.VpcPeeringConnectionId),
					Requester: aws.ToString(peer.RequesterVpcInfo.VpcId),
					Accepter:  aws.ToString(peer.AccepterVpcInfo.VpcId),
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}
	}
}

func mapNetworkInterfaces(vpcs map[string]*VPC, networkInterfaces []types.NetworkInterface) {
	for _, iface := range networkInterfaces {
		var publicIP string
		if iface.Association != nil {
			publicIP = aws.ToString(iface.Association.PublicIp)
		}

		ifaceIn := NetworkInterface{
			NetworkInterfaceData: NetworkInterfaceData{
				ID:                  aws.ToString(iface.NetworkInterfaceId),
				PrivateIP:           aws.ToString(iface.PrivateIpAddress),
				MAC:                 aws.ToString(iface.MacAddress),
				PublicIP:            publicIP,
				Type:                string(iface.InterfaceType),
				Description:         aws.ToString(iface.Description),
				Name:                getNameTag(iface.TagSet),
				SubnetID:            aws.ToString(iface.SubnetId),
				RawNetworkInterface: iface,
			},
			Groups: make(map[string]*SecurityGroup),
		}

		if ifaceIn.Type == "nat_gateway" {
			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for natGatewayID, natGateway := range subnet.NatGateways {
						for _, natGatewayAddress := range natGateway.RawNatGateway.NatGatewayAddresses {
							if aws.ToString(natGatewayAddress.NetworkInterfaceId) == ifaceIn.ID {
								vpcs[vpcID].Subnets[subnetID].NatGateways[natGatewayID].Interfaces[ifaceIn.ID] = &ifaceIn
							}
						}
					}
				}
			}

			continue
		}

		if string(iface.InterfaceType) == "vpc_endpoint" {
			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for endpointID, endpoint := range subnet.InterfaceEndpoints {
						for _, endpointENIId := range endpoint.RawEndpoint.NetworkInterfaceIds {
							if ifaceIn.ID == string(endpointENIId) {
								// Network interface id found in endpoint
								vpcs[vpcID].
									Subnets[subnetID].
									InterfaceEndpoints[endpointID].
									Interfaces[aws.ToString(iface.NetworkInterfaceId)] = &ifaceIn
							}
						}
					}
				}
			}

			continue // Dont duplicate this eni anywhere else
		}

		if iface.Attachment != nil && aws.ToString(iface.Attachment.InstanceId) != "" {
			ifaceInstanceID := aws.ToString(iface.Attachment.InstanceId)

			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for instanceID := range subnet.Instances {
						if ifaceInstanceID == instanceID {
							vpcs[vpcID].
								Subnets[subnetID].
								Instances[instanceID].
								Interfaces[aws.ToString(iface.NetworkInterfaceId)] = &ifaceIn
						}
					}
				}
			}

			continue // The interface is already displayed as a part of the instance, no need to duplicate
		}

		vpcs[aws.ToString(iface.VpcId)].
			Subnets[aws.ToString(iface.SubnetId)].
			ENIs[aws.ToString(iface.NetworkInterfaceId)] = &ifaceIn
	}
}

func extractRules(rules []types.IpPermission) []*SecurityGroupRule {
	rulesOut := []*SecurityGroupRule{}

	for _, rule := range rules {
		IPR := []*IPRange{}
		for _, iprange := range rule.IpRanges {
			IPR = append(IPR, &IPRange{
				CidrIP:      aws.ToString(iprange.CidrIp),
				Description: aws.ToString(iprange.Description),
			})
		}

		IPR6 := []*IPv6Range{}
		for _, ipv6range := range rule.Ipv6Ranges {
			IPR6 = append(IPR6, &IPv6Range{
				CidrIPV6:    aws.ToString(ipv6range.CidrIpv6),
				Description: aws.ToString(ipv6range.Description),
			})
		}

		groups := []*Group{}
		for _, group := range rule.UserIdGroupPairs {
			groups = append(groups, &Group{
				AccountId:   aws.ToString(group.UserId),
				GroupId:     aws.ToString(group.GroupId),
				Description: aws.ToString(group.Description),
			})
		}

		rulesOut = append(rulesOut, &SecurityGroupRule{
			FromPort:   aws.ToInt32(rule.FromPort),
			ToPort:     aws.ToInt32(rule.ToPort),
			IPProtocol: aws.ToString(rule.IpProtocol),
			IPRanges:   IPR,
			IPv6Ranges: IPR6,
			Groups:     groups,
		})
	}

	return rulesOut
}

func mapSecurityGroups(vpcs map[string]*VPC, securityGroups []types.SecurityGroup) {
	for _, securityGroup := range securityGroups {
		InboundRules := extractRules(securityGroup.IpPermissions)
		OutboundRules := extractRules(securityGroup.IpPermissionsEgress)

		securityGroupIn := &SecurityGroup{
			Description:         aws.ToString(securityGroup.Description),
			GroupID:             aws.ToString(securityGroup.GroupId),
			GroupName:           aws.ToString(securityGroup.GroupName),
			IPPermissions:       InboundRules,
			IPPermissionsEgress: OutboundRules,
			TagName:             getNameTag(securityGroup.Tags),
			RawSecurityGroup:    securityGroup,
		}

		// check instance interfaces
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for instanceID, instance := range subnet.Instances {
					for interfaceID, iface := range instance.Interfaces {
						for _, group := range iface.RawNetworkInterface.Groups {
							if aws.ToString(group.GroupId) == securityGroupIn.GroupID {
								vpcs[vpcID].
									Subnets[subnetID].
									Instances[instanceID].
									Interfaces[interfaceID].
									Groups[securityGroupIn.GroupID] = securityGroupIn
							}
						}
					}
				}
			}
		}

		// check ENIs
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for ifaceID, iface := range subnet.ENIs {
					for _, group := range iface.RawNetworkInterface.Groups {
						if aws.ToString(group.GroupId) == securityGroupIn.GroupID {
							vpcs[vpcID].
								Subnets[subnetID].ENIs[ifaceID].Groups[securityGroupIn.GroupID] = securityGroupIn
						}
					}
				}
			}
		}

		// Check Interface Endpoints
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for endpointID, endpoint := range subnet.InterfaceEndpoints {
					for ifaceID, iface := range endpoint.Interfaces {
						for _, group := range iface.RawNetworkInterface.Groups {
							if aws.ToString(group.GroupId) == securityGroupIn.GroupID {
								vpcs[vpcID].
									Subnets[subnetID].
									InterfaceEndpoints[endpointID].
									Interfaces[ifaceID].
									Groups[securityGroupIn.GroupID] = securityGroupIn
							}
						}
					}
				}
			}
		}

		// Check NAT Gateways
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for natGatewayID, natGateway := range subnet.NatGateways {
					for ifaceID, iface := range natGateway.Interfaces {
						for _, group := range iface.RawNetworkInterface.Groups {
							if aws.ToString(group.GroupId) == securityGroupIn.GroupID {
								vpcs[vpcID].Subnets[subnetID].NatGateways[natGatewayID].Interfaces[ifaceID].Groups[securityGroupIn.GroupID] = securityGroupIn
							}
						}
					}
				}
			}
		}
	}
}

func mapVpcEndpoints(vpcs map[string]*VPC, vpcEndpoints []types.VpcEndpoint) {
	vpcIDs := dumpVpcIDs(vpcs)
	subnetIDs := dumpSubnetIDs(vpcs)

	for _, endpoint := range vpcEndpoints {
		// Validate vpc and subnet values
		if _, exists := vpcIDs[aws.ToString(endpoint.VpcId)]; !exists {
			fmt.Printf("Warning: undiscovered VPC %v when processing endpoint %v\n",
				aws.ToString(endpoint.VpcId),
				aws.ToString(endpoint.VpcEndpointId),
			)

			continue
		}

		if string(endpoint.VpcEndpointType) == "Interface" {
			for _, subnet := range endpoint.SubnetIds {
				if _, exists := subnetIDs[subnet]; !exists {
					fmt.Printf("Warning: undiscovered subnet %v when processing endpoint %v\n",
						subnet,
						aws.ToString(endpoint.VpcEndpointId),
					)

					continue
				}

				vpcs[aws.ToString(endpoint.VpcId)].Subnets[subnet].InterfaceEndpoints[aws.ToString(endpoint.VpcEndpointId)] = &InterfaceEndpoint{
					InterfaceEndpointData: InterfaceEndpointData{
						ID:          aws.ToString(endpoint.VpcEndpointId),
						ServiceName: aws.ToString(endpoint.ServiceName),
						Name:        getNameTag(endpoint.Tags),
						RawEndpoint: endpoint,
					},
					Interfaces: make(map[string]*NetworkInterface),
				}
			}
		}

		if string(endpoint.VpcEndpointType) == "Gateway" {
			for _, rtb := range endpoint.RouteTableIds {
				for _, subnet := range vpcs[aws.ToString(endpoint.VpcId)].Subnets {
					if subnet.RouteTable.ID == rtb {
						vpcs[aws.ToString(endpoint.VpcId)].Subnets[subnet.ID].GatewayEndpoints[aws.ToString(endpoint.VpcEndpointId)] = &GatewayEndpoint{
							ID:          aws.ToString(endpoint.VpcEndpointId),
							ServiceName: aws.ToString(endpoint.ServiceName),
							Name:        getNameTag(endpoint.Tags),
							RawEndpoint: endpoint,
						}
					}
				}
			}
		}
	}
}

func dumpVpcIDs(vpcs map[string]*VPC) map[string]bool {
	keys := make(map[string]bool)

	for vpcID := range vpcs {
		keys[vpcID] = true
	}

	return keys
}

func dumpSubnetIDs(vpcs map[string]*VPC) map[string]bool {
	keys := make(map[string]bool)

	for _, vpc := range vpcs {
		for subnetID := range vpc.Subnets {
			keys[subnetID] = true
		}
	}

	return keys
}
