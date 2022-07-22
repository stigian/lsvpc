// Copyright 2021 Stigian Consulting - reference license in top level of project
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

func mapVpcs(vpcs map[string]VPC, vpcData []*ec2.Vpc) {

	for _, v := range vpcData {

		var v6cidr *string
		if v.Ipv6CidrBlockAssociationSet != nil {
			for _, assoc := range v.Ipv6CidrBlockAssociationSet {
				if aws.StringValue(assoc.Ipv6CidrBlockState.State) == "associated" {
					v6cidr = assoc.Ipv6CidrBlock
				}
			}
		}

		vpcs[aws.StringValue(v.VpcId)] = VPC{
			VPCData: VPCData{
				Id:            aws.StringValue(v.VpcId),
				IsDefault:     aws.BoolValue(v.IsDefault),
				CidrBlock:     aws.StringValue(v.CidrBlock),
				IPv6CidrBlock: aws.StringValue(v6cidr),
				Name:          getNameTag(v.Tags),
				RawVPC:        v,
			},
			Subnets: make(map[string]Subnet),
			Peers:   make(map[string]VPCPeer),
		}
	}
}

func mapSubnets(vpcs map[string]VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		isPublic := aws.BoolValue(v.MapCustomerOwnedIpOnLaunch) || aws.BoolValue(v.MapPublicIpOnLaunch)

		vpcs[*v.VpcId].Subnets[*v.SubnetId] = Subnet{
			SubnetData: SubnetData{
				Id:                 aws.StringValue(v.SubnetId),
				CidrBlock:          aws.StringValue(v.CidrBlock),
				AvailabilityZone:   aws.StringValue(v.AvailabilityZone),
				AvailabilityZoneId: aws.StringValue(v.AvailabilityZoneId),
				Name:               getNameTag(v.Tags),
				RawSubnet:          v,
				Public:             isPublic,
			},
			Instances:          make(map[string]Instance),
			NatGateways:        make(map[string]NatGateway),
			TGWs:               make(map[string]TGWAttachment),
			ENIs:               make(map[string]NetworkInterface),
			InterfaceEndpoints: make(map[string]InterfaceEndpoint),
			GatewayEndpoints:   make(map[string]GatewayEndpoint),
		}

	}
}

func mapInstances(vpcs map[string]VPC, reservations []*ec2.Reservation) {
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {

			if *instance.State.Name != "terminated" {
				vpcId := aws.StringValue(instance.VpcId)
				subnetId := aws.StringValue(instance.SubnetId)
				instanceId := aws.StringValue(instance.InstanceId)

				if vpcId != "" && subnetId != "" && instanceId != "" {

					vpcs[vpcId].Subnets[subnetId].Instances[instanceId] = Instance{
						InstanceData: InstanceData{
							Id:        aws.StringValue(instance.InstanceId),
							Type:      aws.StringValue(instance.InstanceType),
							SubnetId:  aws.StringValue(instance.SubnetId),
							VpcId:     aws.StringValue(instance.VpcId),
							State:     aws.StringValue(instance.State.Name),
							PublicIP:  aws.StringValue(instance.PublicIpAddress),
							PrivateIP: aws.StringValue(instance.PrivateIpAddress),
							RawEc2:    instance,
							Name:      getNameTag(instance.Tags),
						},
						Volumes:    make(map[string]Volume),
						Interfaces: make(map[string]NetworkInterface),
					}
				}
			}
		}
	}
}

func mapInstanceStatuses(vpcs map[string]VPC, statuses []*ec2.InstanceStatus) {
	for _, status := range statuses {
		for vpcId, vpc := range vpcs {
			for subnetId, subnet := range vpc.Subnets {
				for instanceId, instance := range subnet.Instances {
					if aws.StringValue(status.InstanceId) == instanceId {
						updatedInstance := instance
						updatedInstance.InstanceStatus = aws.StringValue(status.InstanceStatus.Status)
						updatedInstance.SystemStatus = aws.StringValue(status.SystemStatus.Status)
						vpcs[vpcId].
							Subnets[subnetId].
							Instances[instanceId] = updatedInstance
					}
				}
			}
		}
	}
}

func mapVolumes(vpcs map[string]VPC, volumes []*ec2.Volume) {
	for _, volume := range volumes {
		for _, attachment := range volume.Attachments {
			if volInstanceId := aws.StringValue(attachment.InstanceId); volInstanceId != "" {
				for vpcId, vpc := range vpcs {
					for subnetId, subnet := range vpc.Subnets {
						for instanceId := range subnet.Instances {
							if volInstanceId == instanceId {
								vpcs[vpcId].
									Subnets[subnetId].
									Instances[instanceId].
									Volumes[*volume.VolumeId] = Volume{
									Id:         volume.VolumeId,
									DeviceName: attachment.Device,
									Size:       volume.Size,
									VolumeType: volume.VolumeType,
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

func mapNatGateways(vpcs map[string]VPC, natGateways []*ec2.NatGateway) {
	for _, gateway := range natGateways {
		if aws.StringValue(gateway.State) == "deleted" {
			continue
		}
		vpcs[*gateway.VpcId].Subnets[*gateway.SubnetId].NatGateways[*gateway.NatGatewayId] = NatGateway{
			Id:            gateway.NatGatewayId,
			PrivateIP:     gateway.NatGatewayAddresses[0].PrivateIp,
			PublicIP:      gateway.NatGatewayAddresses[0].PublicIp,
			State:         gateway.State,
			Type:          gateway.ConnectivityType,
			Name:          getNameTag(gateway.Tags),
			RawNatGateway: gateway,
		}
	}
}

func getDefaultRoute(rtb *ec2.RouteTable) string {
	for _, route := range rtb.Routes {
		if aws.StringValue(route.DestinationCidrBlock) == "0.0.0.0/0" ||
			aws.StringValue(route.DestinationIpv6CidrBlock) == "::/0" {

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
	}
	return "" //no default route found, which doesn't necessarily mean an error
}

func mapRouteTables(vpcs map[string]VPC, routeTables []*ec2.RouteTable) {
	// AWS doesn't actually have explicit queryable associations of route
	// tables to subnets. if no other route tables say they are associated
	// with a subnet, then that subnet is assumed to be on the default route table.
	// You can't determine this by looking at the subnets themselves, you
	// have to instead look at all route tables and pick out the ones
	// that say they are associated with particular subnets, and the
	// default route table doesn't even say which subnets they are
	// associated with.

	// first pass, associate the default route with everything
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			if association.Main != nil && *association.Main {
				for subnet_id := range vpcs[*routeTable.VpcId].Subnets {
					subnet := vpcs[*routeTable.VpcId].Subnets[subnet_id]
					defaultRoute := getDefaultRoute(routeTable)
					subnet.RouteTable = &RouteTable{
						Id:       routeTable.RouteTableId,
						Default:  &defaultRoute,
						RawRoute: routeTable,
					}
					vpcs[*routeTable.VpcId].Subnets[subnet_id] = subnet
				}
			}
		}
	}

	// second pass, look at each route table's associations and assign them
	// to their explicitly mentioned subnet
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			//default route doesn't have subnet ids and will cause a nil dereference
			if aws.StringValue(association.AssociationState.State) != "associated" ||
				aws.BoolValue(association.Main) {
				continue
			}
			subnet := vpcs[*routeTable.VpcId].Subnets[*association.SubnetId]
			defaultRoute := getDefaultRoute(routeTable)
			subnet.RouteTable = &RouteTable{
				Id:       routeTable.RouteTableId,
				Default:  &defaultRoute,
				RawRoute: routeTable,
			}
			vpcs[*routeTable.VpcId].Subnets[*association.SubnetId] = subnet
		}
	}
}

func mapInternetGateways(vpcs map[string]VPC, internetGateways []*ec2.InternetGateway) {
	for _, igw := range internetGateways {
		for _, attachment := range igw.Attachments {
			if vpcId := aws.StringValue(attachment.VpcId); vpcId != "" {
				vpc := vpcs[vpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(igw.InternetGatewayId))
				vpcs[vpcId] = vpc
			}
		}
	}
}

func mapEgressOnlyInternetGateways(vpcs map[string]VPC, EOIGWs []*ec2.EgressOnlyInternetGateway) {
	for _, eoigw := range EOIGWs {
		for _, attach := range eoigw.Attachments {
			if aws.StringValue(attach.State) == "attached" {
				vpc := vpcs[*attach.VpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(eoigw.EgressOnlyInternetGatewayId))
				vpcs[*attach.VpcId] = vpc
			}
		}
	}
}

func mapVPNGateways(vpcs map[string]VPC, VPNGateways []*ec2.VpnGateway) {
	for _, vpgw := range VPNGateways {
		for _, attach := range vpgw.VpcAttachments {
			if aws.StringValue(attach.State) == "attached" {
				vpc := vpcs[*attach.VpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(vpgw.VpnGatewayId))
				vpcs[*attach.VpcId] = vpc
			}
		}
	}
}

func mapTransitGatewayVpcAttachments(vpcs map[string]VPC, TransitGatewayVpcAttachments []*ec2.TransitGatewayVpcAttachment, identity *sts.GetCallerIdentityOutput) {
	for _, tgwatt := range TransitGatewayVpcAttachments {
		//Transit Gateway vpc attachments are reported for external accounts too, need to omit those to fit in this data model
		if aws.StringValue(tgwatt.VpcOwnerId) == aws.StringValue(identity.Account) {
			if vpcId := aws.StringValue(tgwatt.VpcId); vpcId != "" {
				for _, subnet := range tgwatt.SubnetIds {
					if subnetId := aws.StringValue(subnet); subnetId != "" {
						vpcs[vpcId].Subnets[subnetId].TGWs[aws.StringValue(tgwatt.TransitGatewayAttachmentId)] = TGWAttachment{
							AttachmentId:     tgwatt.TransitGatewayAttachmentId,
							TransitGatewayId: tgwatt.TransitGatewayId,
							Name:             getNameTag(tgwatt.Tags),
							RawAttachment:    tgwatt,
						}
					}
				}
			}
		}
	}
}

func mapVpcPeeringConnections(vpcs map[string]VPC, VpcPeeringConnections []*ec2.VpcPeeringConnection) {
	for _, peer := range VpcPeeringConnections {
		if aws.StringValue(peer.Status.Code) != "active" {
			continue
		}
		if requester := aws.StringValue(peer.RequesterVpcInfo.VpcId); requester != "" {
			if _, ok := vpcs[requester]; ok {
				vpcs[requester].Peers[aws.StringValue(peer.VpcPeeringConnectionId)] = VPCPeer{
					Id:        peer.VpcPeeringConnectionId,
					Requester: peer.RequesterVpcInfo.VpcId,
					Accepter:  peer.AccepterVpcInfo.VpcId,
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}
		if accepter := aws.StringValue(peer.AccepterVpcInfo.VpcId); accepter != "" {
			if _, ok := vpcs[accepter]; ok {
				vpcs[accepter].Peers[aws.StringValue(peer.VpcPeeringConnectionId)] = VPCPeer{
					Id:        peer.VpcPeeringConnectionId,
					Requester: peer.RequesterVpcInfo.VpcId,
					Accepter:  peer.AccepterVpcInfo.VpcId,
					Name:      getNameTag(peer.Tags),
					RawPeer:   peer,
				}
			}
		}
	}
}

func mapNetworkInterfaces(vpcs map[string]VPC, networkInterfaces []*ec2.NetworkInterface) {
	for _, iface := range networkInterfaces {
		if aws.StringValue(iface.InterfaceType) == "nat_gateway" {
			continue //nat gateways are already adequately reported
		}

		var publicIp *string
		if iface.Association != nil {
			publicIp = iface.Association.PublicIp
		}

		ifaceIn := NetworkInterface{
			Id:                  iface.NetworkInterfaceId,
			PrivateIp:           iface.PrivateIpAddress,
			MAC:                 iface.MacAddress,
			PublicIp:            publicIp,
			Type:                iface.InterfaceType,
			Description:         iface.Description,
			Name:                getNameTag(iface.TagSet),
			RawNetworkInterface: iface,
		}

		if iface.Attachment != nil && aws.StringValue(iface.Attachment.InstanceId) != "" {
			ifaceInstanceId := aws.StringValue(iface.Attachment.InstanceId)
			for vpcId, vpc := range vpcs {
				for subnetId, subnet := range vpc.Subnets {
					for instanceId := range subnet.Instances {
						if ifaceInstanceId == instanceId {
							vpcs[vpcId].
								Subnets[subnetId].
								Instances[instanceId].
								Interfaces[*iface.NetworkInterfaceId] = ifaceIn
						}
					}
				}
			}
			continue //The interface is already displayed as a part of the instance, no need to duplicate
		}

		vpcs[*iface.VpcId].
			Subnets[*iface.SubnetId].
			ENIs[*iface.NetworkInterfaceId] = ifaceIn
	}
}

func mapVpcEndpoints(vpcs map[string]VPC, vpcEndpoints []*ec2.VpcEndpoint) {
	vpcIds := dumpVpcIds(vpcs)
	subnetIds := dumpSubnetIds(vpcs)
	for _, endpoint := range vpcEndpoints {
		//validate vpc and subnet values
		if _, exists := vpcIds[aws.StringValue(endpoint.VpcId)]; !exists {
			fmt.Printf("Warning: undiscovered VPC %v when processing endpoint %v\n",
				aws.StringValue(endpoint.VpcId),
				aws.StringValue(endpoint.VpcEndpointId))
			continue
		}

		if aws.StringValue(endpoint.VpcEndpointType) == "Interface" {
			for _, subnet := range endpoint.SubnetIds {
				if _, exists := subnetIds[aws.StringValue(subnet)]; !exists {
					fmt.Printf("Warning: undiscovered subnet %v when processing endpoint %v\n",
						aws.StringValue(subnet),
						aws.StringValue(endpoint.VpcEndpointId))
					continue
				}
				vpcs[*endpoint.VpcId].Subnets[*subnet].InterfaceEndpoints[*endpoint.VpcEndpointId] = InterfaceEndpoint{
					Id:          endpoint.VpcEndpointId,
					ServiceName: endpoint.ServiceName,
					Name:        getNameTag(endpoint.Tags),
					RawEndpoint: endpoint,
				}
			}
		}

		if aws.StringValue(endpoint.VpcEndpointType) == "Gateway" {
			for _, rtb := range endpoint.RouteTableIds {
				for _, subnet := range vpcs[*endpoint.VpcId].Subnets {
					if aws.StringValue(subnet.RouteTable.Id) == aws.StringValue(rtb) {
						vpcs[*endpoint.VpcId].Subnets[subnet.Id].GatewayEndpoints[*endpoint.VpcEndpointId] = GatewayEndpoint{
							Id:          endpoint.VpcEndpointId,
							ServiceName: endpoint.ServiceName,
							Name:        getNameTag(endpoint.Tags),
							RawEndpoint: endpoint,
						}
					}
				}
			}
		}
	}
}

func dumpVpcIds(vpcs map[string]VPC) map[string]bool {
	keys := make(map[string]bool)
	for vpcId := range vpcs {
		keys[vpcId] = true
	}
	return keys
}

func dumpSubnetIds(vpcs map[string]VPC) map[string]bool {
	keys := make(map[string]bool)

	for _, vpc := range vpcs {
		for subnetId := range vpc.Subnets {
			keys[subnetId] = true
		}
	}
	return keys
}
