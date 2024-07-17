// Copyright 2023 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getNameTag(tags []*ec2.Tag) string {
	var name *string

	for _, tag := range tags {
		if aws.StringValue(tag.Key) == "Name" {
			name = tag.Value
		}
	}

	return aws.StringValue(name)
}

func mapVpcs(vpcs map[string]*VPC, vpcData []*ec2.Vpc) {
	for _, v := range vpcData {
		var v6cidr *string

		if v.Ipv6CidrBlockAssociationSet != nil {
			for _, assoc := range v.Ipv6CidrBlockAssociationSet {
				if aws.StringValue(assoc.Ipv6CidrBlockState.State) == "associated" {
					v6cidr = assoc.Ipv6CidrBlock
				}
			}
		}

		vpcs[aws.StringValue(v.VpcId)] = &VPC{
			VPCData: VPCData{
				ID:            aws.StringValue(v.VpcId),
				IsDefault:     aws.BoolValue(v.IsDefault),
				CidrBlock:     aws.StringValue(v.CidrBlock),
				IPv6CidrBlock: aws.StringValue(v6cidr),
				Name:          getNameTag(v.Tags),
			},
			RawVPC:  v,
			Subnets: make(map[string]*Subnet),
			Peers:   make(map[string]*VPCPeer),
		}
	}
}

func mapSubnets(vpcs map[string]*VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		isPublic := aws.BoolValue(v.MapCustomerOwnedIpOnLaunch) || aws.BoolValue(v.MapPublicIpOnLaunch)

		vpcs[*v.VpcId].Subnets[*v.SubnetId] = &Subnet{
			SubnetData: SubnetData{
				ID:                 aws.StringValue(v.SubnetId),
				CidrBlock:          aws.StringValue(v.CidrBlock),
				AvailabilityZone:   aws.StringValue(v.AvailabilityZone),
				AvailabilityZoneID: aws.StringValue(v.AvailabilityZoneId),
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

func mapInstances(vpcs map[string]*VPC, reservations []*ec2.Reservation) {
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if *instance.State.Name != "terminated" {
				vpcID := aws.StringValue(instance.VpcId)
				subnetID := aws.StringValue(instance.SubnetId)
				instanceID := aws.StringValue(instance.InstanceId)

				if vpcID != "" && subnetID != "" && instanceID != "" {
					vpcs[vpcID].Subnets[subnetID].Instances[instanceID] = &Instance{
						InstanceData: InstanceData{
							ID:        aws.StringValue(instance.InstanceId),
							Type:      aws.StringValue(instance.InstanceType),
							SubnetID:  aws.StringValue(instance.SubnetId),
							VpcID:     aws.StringValue(instance.VpcId),
							State:     aws.StringValue(instance.State.Name),
							PublicIP:  aws.StringValue(instance.PublicIpAddress),
							PrivateIP: aws.StringValue(instance.PrivateIpAddress),
							Name:      getNameTag(instance.Tags),
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

func mapInstanceStatuses(vpcs map[string]*VPC, statuses []*ec2.InstanceStatus) {
	for _, status := range statuses {
		for vpcID, vpc := range vpcs {
			for subnetID, subnet := range vpc.Subnets {
				for instanceID, instance := range subnet.Instances {
					if aws.StringValue(status.InstanceId) == instanceID {
						updatedInstance := instance
						updatedInstance.InstanceStatus = aws.StringValue(status.InstanceStatus.Status)
						updatedInstance.SystemStatus = aws.StringValue(status.SystemStatus.Status)
						vpcs[vpcID].
							Subnets[subnetID].
							Instances[instanceID] = updatedInstance
					}
				}
			}
		}
	}
}

func mapVolumes(vpcs map[string]*VPC, volumes []*ec2.Volume) {
	for _, volume := range volumes {
		for _, attachment := range volume.Attachments {
			if volInstanceID := aws.StringValue(attachment.InstanceId); volInstanceID != "" {
				for vpcID, vpc := range vpcs {
					for subnetID, subnet := range vpc.Subnets {
						for instanceID := range subnet.Instances {
							if volInstanceID == instanceID {
								vpcs[vpcID].
									Subnets[subnetID].
									Instances[instanceID].
									Volumes[*volume.VolumeId] = &Volume{
									ID:         aws.StringValue(volume.VolumeId),
									DeviceName: aws.StringValue(attachment.Device),
									Size:       aws.Int64Value(volume.Size),
									VolumeType: aws.StringValue(volume.VolumeType),
									Encrypted:  aws.BoolValue(volume.Encrypted),
									KMSKeyId:   aws.StringValue(volume.KmsKeyId),
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

func mapNatGateways(vpcs map[string]*VPC, natGateways []*ec2.NatGateway) {
	for _, gateway := range natGateways {
		if aws.StringValue(gateway.State) == "deleted" {
			continue
		}

		vpcs[*gateway.VpcId].Subnets[*gateway.SubnetId].NatGateways[*gateway.NatGatewayId] = &NatGateway{
			NatGatewayData: NatGatewayData{
				ID:            aws.StringValue(gateway.NatGatewayId),
				PrivateIP:     aws.StringValue(gateway.NatGatewayAddresses[0].PrivateIp),
				PublicIP:      aws.StringValue(gateway.NatGatewayAddresses[0].PublicIp),
				State:         aws.StringValue(gateway.State),
				Type:          aws.StringValue(gateway.ConnectivityType),
				Name:          getNameTag(gateway.Tags),
				RawNatGateway: gateway,
			},
			Interfaces: make(map[string]*NetworkInterface),
		}
	}
}

func getDefaultRoute(rtb *ec2.RouteTable) string {
	for _, route := range rtb.Routes {
		if !(aws.StringValue(route.DestinationCidrBlock) == "0.0.0.0/0" ||
			aws.StringValue(route.DestinationIpv6CidrBlock) == "::/0") {
			continue
		}

		if dest := aws.StringValue(route.CarrierGatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.EgressOnlyInternetGatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.GatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.InstanceId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.LocalGatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.NatGatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.NetworkInterfaceId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.TransitGatewayId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.VpcPeeringConnectionId); dest != "" {
			return dest
		}

		if dest := aws.StringValue(route.CoreNetworkArn); dest != "" {
			return dest
		}
	}

	return "" // No default route found, which doesn't necessarily mean an error
}

func mapRouteTables(vpcs map[string]*VPC, routeTables []*ec2.RouteTable) {
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
			if association.Main != nil && *association.Main {
				for subnetID := range vpcs[*routeTable.VpcId].Subnets {
					subnet := vpcs[*routeTable.VpcId].Subnets[subnetID]
					defaultRoute := getDefaultRoute(routeTable)
					subnet.RouteTable = &RouteTable{
						ID:       aws.StringValue(routeTable.RouteTableId),
						Default:  aws.StringValue(&defaultRoute),
						RawRoute: routeTable,
					}
					vpcs[*routeTable.VpcId].Subnets[subnetID] = subnet
				}
			}
		}
	}

	// second pass, look at each route table's associations and assign them
	// to their explicitly mentioned subnet
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			// default route doesn't have subnet ids and will cause a nil dereference
			if aws.StringValue(association.AssociationState.State) != "associated" ||
				aws.BoolValue(association.Main) {
				continue
			}

			subnet := vpcs[*routeTable.VpcId].Subnets[*association.SubnetId]
			defaultRoute := getDefaultRoute(routeTable)
			subnet.RouteTable = &RouteTable{
				ID:       aws.StringValue(routeTable.RouteTableId),
				Default:  aws.StringValue(&defaultRoute),
				RawRoute: routeTable,
			}
			vpcs[*routeTable.VpcId].Subnets[*association.SubnetId] = subnet
		}
	}
}

func mapInternetGateways(vpcs map[string]*VPC, internetGateways []*ec2.InternetGateway) {
	for _, igw := range internetGateways {
		for _, attachment := range igw.Attachments {
			if vpcID := aws.StringValue(attachment.VpcId); vpcID != "" {
				vpc := vpcs[vpcID]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(igw.InternetGatewayId))
				vpcs[vpcID] = vpc
			}
		}
	}
}

func mapEgressOnlyInternetGateways(vpcs map[string]*VPC, eOIGWs []*ec2.EgressOnlyInternetGateway) {
	for _, eoigw := range eOIGWs {
		for _, attach := range eoigw.Attachments {
			if aws.StringValue(attach.State) == "attached" {
				vpc := vpcs[*attach.VpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(eoigw.EgressOnlyInternetGatewayId))
				vpcs[*attach.VpcId] = vpc
			}
		}
	}
}

func mapVPNGateways(vpcs map[string]*VPC, vpnGateways []*ec2.VpnGateway) {
	for _, vpgw := range vpnGateways {
		for _, attach := range vpgw.VpcAttachments {
			if aws.StringValue(attach.State) == "attached" {
				vpc := vpcs[*attach.VpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(vpgw.VpnGatewayId))
				vpcs[*attach.VpcId] = vpc
			}
		}
	}
}

func mapTransitGatewayVpcAttachments(vpcs map[string]*VPC, transitGatewayVpcAttachments []*ec2.TransitGatewayVpcAttachment, identity *sts.GetCallerIdentityOutput) {
	for _, tgwatt := range transitGatewayVpcAttachments {
		// Transit Gateway vpc attachments are reported for external accounts too, need to omit those to fit in this data model
		if aws.StringValue(tgwatt.VpcOwnerId) == aws.StringValue(identity.Account) {
			if vpcID := aws.StringValue(tgwatt.VpcId); vpcID != "" {
				for _, subnet := range tgwatt.SubnetIds {
					if subnetID := aws.StringValue(subnet); subnetID != "" {
						vpcs[vpcID].Subnets[subnetID].TGWs[aws.StringValue(tgwatt.TransitGatewayAttachmentId)] = &TGWAttachment{
							AttachmentID:     aws.StringValue(tgwatt.TransitGatewayAttachmentId),
							TransitGatewayID: aws.StringValue(tgwatt.TransitGatewayId),
							Name:             getNameTag(tgwatt.Tags),
							RawAttachment:    tgwatt,
						}
					}
				}
			}
		}
	}
}

func mapVpcPeeringConnections(vpcs map[string]*VPC, vpcPeeringConnections []*ec2.VpcPeeringConnection) {
	for _, peer := range vpcPeeringConnections {
		if aws.StringValue(peer.Status.Code) != "active" {
			continue
		}

		if requester := aws.StringValue(peer.RequesterVpcInfo.VpcId); requester != "" {
			if _, ok := vpcs[requester]; ok {
				vpcs[requester].Peers[aws.StringValue(peer.VpcPeeringConnectionId)] = &VPCPeer{
					ID:        aws.StringValue(peer.VpcPeeringConnectionId),
					Requester: aws.StringValue(peer.RequesterVpcInfo.VpcId),
					Accepter:  aws.StringValue(peer.AccepterVpcInfo.VpcId),
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}

		if accepter := aws.StringValue(peer.AccepterVpcInfo.VpcId); accepter != "" {
			if _, ok := vpcs[accepter]; ok {
				vpcs[accepter].Peers[aws.StringValue(peer.VpcPeeringConnectionId)] = &VPCPeer{
					ID:        aws.StringValue(peer.VpcPeeringConnectionId),
					Requester: aws.StringValue(peer.RequesterVpcInfo.VpcId),
					Accepter:  aws.StringValue(peer.AccepterVpcInfo.VpcId),
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}
	}
}

func mapNetworkInterfaces(vpcs map[string]*VPC, networkInterfaces []*ec2.NetworkInterface) {
	for _, iface := range networkInterfaces {
		var publicIP *string
		if iface.Association != nil {
			publicIP = iface.Association.PublicIp
		}

		ifaceIn := NetworkInterface{
			NetworkInterfaceData: NetworkInterfaceData{
				ID:                  aws.StringValue(iface.NetworkInterfaceId),
				PrivateIP:           aws.StringValue(iface.PrivateIpAddress),
				MAC:                 aws.StringValue(iface.MacAddress),
				PublicIP:            aws.StringValue(publicIP),
				Type:                aws.StringValue(iface.InterfaceType),
				Description:         aws.StringValue(iface.Description),
				Name:                getNameTag(iface.TagSet),
				SubnetID:            aws.StringValue(iface.SubnetId),
				RawNetworkInterface: iface,
			},
			Groups: make(map[string]*SecurityGroup),
		}

		if ifaceIn.Type == "nat_gateway" {
			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for natGatewayID, natGateway := range subnet.NatGateways {
						for _, natGatewayAddress := range natGateway.RawNatGateway.NatGatewayAddresses {
							if aws.StringValue(natGatewayAddress.NetworkInterfaceId) == ifaceIn.ID {
								vpcs[vpcID].Subnets[subnetID].NatGateways[natGatewayID].Interfaces[ifaceIn.ID] = &ifaceIn
							}
						}
					}
				}
			}

			continue
		}

		if aws.StringValue(iface.InterfaceType) == "vpc_endpoint" {
			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for endpointID, endpoint := range subnet.InterfaceEndpoints {
						for _, endpointENIId := range endpoint.RawEndpoint.NetworkInterfaceIds {
							if ifaceIn.ID == aws.StringValue(endpointENIId) {
								// Network interface id found in endpoint
								vpcs[vpcID].
									Subnets[subnetID].
									InterfaceEndpoints[endpointID].
									Interfaces[aws.StringValue(iface.NetworkInterfaceId)] = &ifaceIn
							}
						}
					}
				}
			}

			continue // Dont duplicate this eni anywhere else
		}

		if iface.Attachment != nil && aws.StringValue(iface.Attachment.InstanceId) != "" {
			ifaceInstanceID := aws.StringValue(iface.Attachment.InstanceId)

			for vpcID, vpc := range vpcs {
				for subnetID, subnet := range vpc.Subnets {
					for instanceID := range subnet.Instances {
						if ifaceInstanceID == instanceID {
							vpcs[vpcID].
								Subnets[subnetID].
								Instances[instanceID].
								Interfaces[*iface.NetworkInterfaceId] = &ifaceIn
						}
					}
				}
			}

			continue // The interface is already displayed as a part of the instance, no need to duplicate
		}

		vpcs[*iface.VpcId].
			Subnets[*iface.SubnetId].
			ENIs[*iface.NetworkInterfaceId] = &ifaceIn
	}
}

func extractRules(rules []*ec2.IpPermission) []*SecurityGroupRule {
	rulesOut := []*SecurityGroupRule{}

	for _, rule := range rules {
		IPR := []*IPRange{}
		for _, iprange := range rule.IpRanges {
			IPR = append(IPR, &IPRange{
				CidrIP:      aws.StringValue(iprange.CidrIp),
				Description: aws.StringValue(iprange.Description),
			})
		}

		IPR6 := []*IPv6Range{}
		for _, ipv6range := range rule.Ipv6Ranges {
			IPR6 = append(IPR6, &IPv6Range{
				CidrIPV6:    aws.StringValue(ipv6range.CidrIpv6),
				Description: aws.StringValue(ipv6range.Description),
			})
		}

		rulesOut = append(rulesOut, &SecurityGroupRule{
			FromPort:   aws.Int64Value(rule.FromPort),
			ToPort:     aws.Int64Value(rule.ToPort),
			IPProtocol: aws.StringValue(rule.IpProtocol),
			IPRanges:   IPR,
			IPv6Ranges: IPR6,
		})
	}

	return rulesOut
}

func mapSecurityGroups(vpcs map[string]*VPC, securityGroups []*ec2.SecurityGroup) {
	for _, securityGroup := range securityGroups {
		InboundRules := extractRules(securityGroup.IpPermissions)
		OutboundRules := extractRules(securityGroup.IpPermissionsEgress)

		securityGroupIn := &SecurityGroup{
			Description:         aws.StringValue(securityGroup.Description),
			GroupID:             aws.StringValue(securityGroup.GroupId),
			GroupName:           aws.StringValue(securityGroup.GroupName),
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
							if aws.StringValue(group.GroupId) == securityGroupIn.GroupID {
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
						if aws.StringValue(group.GroupId) == securityGroupIn.GroupID {
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
							if aws.StringValue(group.GroupId) == securityGroupIn.GroupID {
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
							if aws.StringValue(group.GroupId) == securityGroupIn.GroupID {
								vpcs[vpcID].Subnets[subnetID].NatGateways[natGatewayID].Interfaces[ifaceID].Groups[securityGroupIn.GroupID] = securityGroupIn
							}
						}
					}
				}
			}
		}
	}
}

func mapVpcEndpoints(vpcs map[string]*VPC, vpcEndpoints []*ec2.VpcEndpoint) {
	vpcIDs := dumpVpcIDs(vpcs)
	subnetIDs := dumpSubnetIDs(vpcs)

	for _, endpoint := range vpcEndpoints {
		// Validate vpc and subnet values
		if _, exists := vpcIDs[aws.StringValue(endpoint.VpcId)]; !exists {
			fmt.Printf("Warning: undiscovered VPC %v when processing endpoint %v\n",
				aws.StringValue(endpoint.VpcId),
				aws.StringValue(endpoint.VpcEndpointId),
			)

			continue
		}

		if aws.StringValue(endpoint.VpcEndpointType) == "Interface" {
			for _, subnet := range endpoint.SubnetIds {
				if _, exists := subnetIDs[aws.StringValue(subnet)]; !exists {
					fmt.Printf("Warning: undiscovered subnet %v when processing endpoint %v\n",
						aws.StringValue(subnet),
						aws.StringValue(endpoint.VpcEndpointId),
					)

					continue
				}

				vpcs[*endpoint.VpcId].Subnets[*subnet].InterfaceEndpoints[*endpoint.VpcEndpointId] = &InterfaceEndpoint{
					InterfaceEndpointData: InterfaceEndpointData{
						ID:          aws.StringValue(endpoint.VpcEndpointId),
						ServiceName: aws.StringValue(endpoint.ServiceName),
						Name:        getNameTag(endpoint.Tags),
						RawEndpoint: endpoint,
					},
					Interfaces: make(map[string]*NetworkInterface),
				}
			}
		}

		if aws.StringValue(endpoint.VpcEndpointType) == "Gateway" {
			for _, rtb := range endpoint.RouteTableIds {
				for _, subnet := range vpcs[*endpoint.VpcId].Subnets {
					if subnet.RouteTable.ID == aws.StringValue(rtb) {
						vpcs[*endpoint.VpcId].Subnets[subnet.ID].GatewayEndpoints[*endpoint.VpcEndpointId] = &GatewayEndpoint{
							ID:          aws.StringValue(endpoint.VpcEndpointId),
							ServiceName: aws.StringValue(endpoint.ServiceName),
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
