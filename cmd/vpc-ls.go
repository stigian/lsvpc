// Copyright 2021 Travis James - reference license in top level of project

package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type RegionData struct {
	VPCs map[string]VPC
}

type VPC struct {
	Id            *string
	IsDefault     bool
	CidrBlock     *string
	IPv6CidrBlock *string
	RawVPC        *ec2.Vpc
	Gateways      []string
	Subnets       map[string]Subnet
	Peers         map[string]VPCPeer
}

type Subnet struct {
	Id                 *string
	CidrBlock          *string
	AvailabilityZone   *string
	AvailabilityZoneId *string
	Public             bool
	RawSubnet          *ec2.Subnet
	RouteTable         *RouteTable
	EC2s               map[string]EC2
	NatGateways        map[string]NatGateway
	TGWs               map[string]TGWAttachment
	ENIs               map[string]NetworkInterface
	InterfaceEndpoints map[string]InterfaceEndpoint
	GatewayEndpoints   map[string]GatewayEndpoint
}

type EC2 struct {
	Id         *string
	Type       *string
	SubnetId   *string
	VpcId      *string
	State      *string
	PublicIP   *string
	PrivateIP  *string
	Volumes    map[string]Volume
	Interfaces []InstanceNetworkInterface
	RawEc2     *ec2.Instance
}

type InstanceNetworkInterface struct {
	Id                  *string
	PrivateIp           *string
	MAC                 *string
	DNS                 *string
	Type                *string
	Description         *string
	PublicIp            *string
	RawNetworkInterface *ec2.InstanceNetworkInterface
}

type NetworkInterface struct {
	Id                  *string
	PrivateIp           *string
	MAC                 *string
	DNS                 *string
	Type                *string
	Description         *string
	PublicIp            *string
	RawNetworkInterface *ec2.NetworkInterface
}

type Volume struct {
	Id         *string
	DeviceName *string
	Size       *int64
	VolumeType *string
	RawVolume  *ec2.Volume
}
type NatGateway struct {
	Id            *string
	PrivateIP     *string
	PublicIP      *string
	State         *string
	Type          *string
	RawNatGateway *ec2.NatGateway
}

type RouteTable struct {
	Id       *string
	Default  *string
	RawRoute *ec2.RouteTable
}

type TGWAttachment struct {
	AttachmentId     *string
	TransitGatewayId *string
	RawAttachment    *ec2.TransitGatewayVpcAttachment
}

type VPCPeer struct {
	Id        *string
	Requester *string
	Accepter  *string
	RawPeer   *ec2.VpcPeeringConnection
}

type VPCEndpoint struct {
	Id      *string
	Type    *string
	Service *string
}

type InterfaceEndpoint struct {
	Id          *string
	ServiceName *string
	RawEndpoint *ec2.VpcEndpoint
}

type GatewayEndpoint struct {
	Id          *string
	ServiceName *string
	RawEndpoint *ec2.VpcEndpoint
}

type RecievedData struct {
	wg                 sync.WaitGroup
	mu                 sync.Mutex
	Vpcs               []*ec2.Vpc
	Subnets            []*ec2.Subnet
	Instances          []*ec2.Reservation
	NatGateways        []*ec2.NatGateway
	RouteTables        []*ec2.RouteTable
	InternetGateways   []*ec2.InternetGateway
	EOInternetGateways []*ec2.EgressOnlyInternetGateway
	VPNGateways        []*ec2.VpnGateway
	TransitGateways    []*ec2.TransitGatewayVpcAttachment
	PeeringConnections []*ec2.VpcPeeringConnection
	NetworkInterfaces  []*ec2.NetworkInterface
	VPCEndpoints       []*ec2.VpcEndpoint
	Error              error
}

func getVpcs(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	vpcs := []*ec2.Vpc{}
	err := svc.DescribeVpcsPages(
		&ec2.DescribeVpcsInput{},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}
	data.Vpcs = vpcs
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
			Id:            v.VpcId,
			IsDefault:     aws.BoolValue(v.IsDefault),
			CidrBlock:     v.CidrBlock,
			IPv6CidrBlock: v6cidr,
			RawVPC:        v,
			Subnets:       make(map[string]Subnet),
			Peers:         make(map[string]VPCPeer),
		}
	}
}

func getSubnets(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	subnets := []*ec2.Subnet{}
	err := svc.DescribeSubnetsPages(
		&ec2.DescribeSubnetsInput{},
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			subnets = append(subnets, page.Subnets...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.Subnets = subnets
}

func getInstances(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	instances := []*ec2.Reservation{}
	err := svc.DescribeInstancesPages(
		&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instances = append(instances, page.Reservations...)
			return !lastPage
		},
	)
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.Instances = instances
}

func getNatGatways(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	natGateways := []*ec2.NatGateway{}

	err := svc.DescribeNatGatewaysPages(
		&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			natGateways = append(natGateways, page.NatGateways...)
			return !lastPage
		},
	)
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.NatGateways = natGateways
}

func getRouteTables(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	routeTables := []*ec2.RouteTable{}

	err := svc.DescribeRouteTablesPages(
		&ec2.DescribeRouteTablesInput{},
		func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, page.RouteTables...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.RouteTables = routeTables
}

func getInternetGateways(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	internetGateways := []*ec2.InternetGateway{}

	err := svc.DescribeInternetGatewaysPages(
		&ec2.DescribeInternetGatewaysInput{},
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, page.InternetGateways...)
			return !lastPage
		},
	)
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.InternetGateways = internetGateways
}

func getEgressOnlyInternetGateways(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	EOIGWs := []*ec2.EgressOnlyInternetGateway{}

	err := svc.DescribeEgressOnlyInternetGatewaysPages(
		&ec2.DescribeEgressOnlyInternetGatewaysInput{},
		func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
			EOIGWs = append(EOIGWs, page.EgressOnlyInternetGateways...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.EOInternetGateways = EOIGWs
}

func getVPNGateways(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	out, err := svc.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.VPNGateways = out.VpnGateways
}

func getTransitGatewayVpcAttachments(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	TGWatt := []*ec2.TransitGatewayVpcAttachment{}

	err := svc.DescribeTransitGatewayVpcAttachmentsPages(
		&ec2.DescribeTransitGatewayVpcAttachmentsInput{},
		func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
			TGWatt = append(TGWatt, page.TransitGatewayVpcAttachments...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}
	data.TransitGateways = TGWatt
}

func getVpcPeeringConnections(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	peers := []*ec2.VpcPeeringConnection{}

	err := svc.DescribeVpcPeeringConnectionsPages(
		&ec2.DescribeVpcPeeringConnectionsInput{},
		func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
			peers = append(peers, page.VpcPeeringConnections...)
			return !lastPage
		},
	)
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.PeeringConnections = peers
}

func getNetworkInterfaces(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	ifaces := []*ec2.NetworkInterface{}

	err := svc.DescribeNetworkInterfacesPages(
		&ec2.DescribeNetworkInterfacesInput{},
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			ifaces = append(ifaces, page.NetworkInterfaces...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.NetworkInterfaces = ifaces
}

func getVpcEndpoints(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	endpoints := []*ec2.VpcEndpoint{}

	err := svc.DescribeVpcEndpointsPages(
		&ec2.DescribeVpcEndpointsInput{},
		func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
			endpoints = append(endpoints, page.VpcEndpoints...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.VPCEndpoints = endpoints
}

func mapSubnets(vpcs map[string]VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		isPublic := aws.BoolValue(v.MapCustomerOwnedIpOnLaunch) || aws.BoolValue(v.MapPublicIpOnLaunch)

		vpcs[*v.VpcId].Subnets[*v.SubnetId] = Subnet{
			Id:                 v.SubnetId,
			CidrBlock:          v.CidrBlock,
			AvailabilityZone:   v.AvailabilityZone,
			AvailabilityZoneId: v.AvailabilityZoneId,
			RawSubnet:          v,
			Public:             isPublic,
			EC2s:               make(map[string]EC2),
			NatGateways:        make(map[string]NatGateway),
			TGWs:               make(map[string]TGWAttachment),
			ENIs:               make(map[string]NetworkInterface),
			InterfaceEndpoints: make(map[string]InterfaceEndpoint),
		}

	}
}

func mapInstances(vpcs map[string]VPC, reservations []*ec2.Reservation) {
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			networkInterfaces := []InstanceNetworkInterface{}
			for _, networkInterface := range instance.NetworkInterfaces {
				networkInterfaces = append(networkInterfaces, InstanceNetworkInterface{
					Id:                  networkInterface.NetworkInterfaceId,
					PrivateIp:           networkInterface.PrivateIpAddress,
					MAC:                 networkInterface.MacAddress,
					DNS:                 networkInterface.PrivateDnsName,
					RawNetworkInterface: networkInterface,
				})
			}

			volumes := make(map[string]Volume)
			for _, volume := range instance.BlockDeviceMappings {
				volumes[*volume.Ebs.VolumeId] = Volume{
					Id:         volume.Ebs.VolumeId,
					DeviceName: volume.DeviceName,
				}
			}

			if *instance.State.Name != "terminated" {
				vpcs[*instance.VpcId].Subnets[*instance.SubnetId].EC2s[*instance.InstanceId] = EC2{
					Id:         instance.InstanceId,
					Type:       instance.InstanceType,
					SubnetId:   instance.SubnetId,
					VpcId:      instance.VpcId,
					State:      instance.State.Name,
					PublicIP:   instance.PublicIpAddress,
					PrivateIP:  instance.PrivateIpAddress,
					Volumes:    volumes,
					Interfaces: networkInterfaces,
					RawEc2:     instance,
				}
			}
		}
	}
}

func mapNatGateways(vpcs map[string]VPC, natGateways []*ec2.NatGateway) {
	for _, gateway := range natGateways {
		vpcs[*gateway.VpcId].Subnets[*gateway.SubnetId].NatGateways[*gateway.NatGatewayId] = NatGateway{
			Id:            gateway.NatGatewayId,
			PrivateIP:     gateway.NatGatewayAddresses[0].PrivateIp,
			PublicIP:      gateway.NatGatewayAddresses[0].PublicIp,
			State:         gateway.State,
			Type:          gateway.ConnectivityType,
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

func mapTransitGatewayVpcAttachments(vpcs map[string]VPC, TransitGatewayVpcAttachments []*ec2.TransitGatewayVpcAttachment) {
	for _, tgwatt := range TransitGatewayVpcAttachments {
		if vpcId := aws.StringValue(tgwatt.VpcId); vpcId != "" {
			for _, subnet := range tgwatt.SubnetIds {
				if subnetId := aws.StringValue(subnet); subnetId != "" {
					vpcs[vpcId].Subnets[subnetId].TGWs[*tgwatt.TransitGatewayAttachmentId] = TGWAttachment{
						AttachmentId:     tgwatt.TransitGatewayAttachmentId,
						TransitGatewayId: tgwatt.TransitGatewayId,
						RawAttachment:    tgwatt,
					}
				}
			}
		}
	}
}

func mapVpcPeeringConnections(vpcs map[string]VPC, VpcPeeringConnections []*ec2.VpcPeeringConnection) {
	for _, peer := range VpcPeeringConnections {
		if requester := aws.StringValue(peer.RequesterVpcInfo.VpcId); requester != "" {
			if _, ok := vpcs[requester]; ok {
				vpcs[requester].Peers[aws.StringValue(peer.VpcPeeringConnectionId)] = VPCPeer{
					Id:        peer.VpcPeeringConnectionId,
					Requester: peer.RequesterVpcInfo.VpcId,
					Accepter:  peer.AccepterVpcInfo.VpcId,
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
					RawPeer:   peer,
				}
			}
		}
	}
}

func mapNetworkInterfaces(vpcs map[string]VPC, networkInterfaces []*ec2.NetworkInterface) {
	for _, iface := range networkInterfaces {
		if aws.StringValue(iface.Attachment.InstanceId) != "" {
			continue //we handle instance interfaces elsewhere
		}

		if aws.StringValue(iface.InterfaceType) == "nat_gateway" {
			continue //nat gateways are already adequately reported
		}

		var publicIp *string
		if iface.Association != nil {
			publicIp = iface.Association.PublicIp
		}

		description := iface.Description
		vpcs[*iface.VpcId].Subnets[*iface.SubnetId].ENIs[*iface.NetworkInterfaceId] = NetworkInterface{
			Id:                  iface.NetworkInterfaceId,
			PrivateIp:           iface.PrivateIpAddress,
			MAC:                 iface.MacAddress,
			PublicIp:            publicIp,
			Type:                iface.InterfaceType,
			Description:         description,
			RawNetworkInterface: iface,
		}
	}
}

func mapVpcEndpoints(vpcs map[string]VPC, vpcEndpoints []*ec2.VpcEndpoint) {
	for _, endpoint := range vpcEndpoints {
		if aws.StringValue(endpoint.VpcEndpointType) == "Interface" {
			for _, subnet := range endpoint.SubnetIds {
				vpcs[*endpoint.VpcId].Subnets[*subnet].InterfaceEndpoints[*endpoint.VpcEndpointId] = InterfaceEndpoint{
					Id:          endpoint.VpcEndpointId,
					ServiceName: endpoint.ServiceName,
					RawEndpoint: endpoint,
				}
			}
		}

		if aws.StringValue(endpoint.VpcEndpointType) == "Gateway" {
			for _, rtb := range endpoint.RouteTableIds {
				for _, subnet := range vpcs[*endpoint.VpcId].Subnets {
					if aws.StringValue(subnet.RouteTable.Id) == aws.StringValue(rtb) {
						vpcs[*endpoint.VpcId].Subnets[*subnet.Id].GatewayEndpoints[*endpoint.VpcEndpointId] = GatewayEndpoint{
							Id:          endpoint.VpcEndpointId,
							ServiceName: endpoint.ServiceName,
						}
					}
				}
			}
		}
	}
}

func getVolume(svc *ec2.EC2, volumeId string) (*ec2.Volume, error) {
	if volumeId == "" {
		return &ec2.Volume{}, fmt.Errorf("getVolume handed an empty string")
	}
	out, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{
			aws.String(volumeId),
		},
	})
	if err != nil {
		return &ec2.Volume{}, err
	}

	if len(out.Volumes) != 1 {
		return &ec2.Volume{}, fmt.Errorf("incorrect number of volumes returned")
	}

	return out.Volumes[0], nil

}

func instantiateVolumes(svc *ec2.EC2, vpcs map[string]VPC) error {
	for vk, v := range vpcs {
		for sk, s := range v.Subnets {
			for ik, i := range s.EC2s {
				for volk, vol := range i.Volumes {
					volume, err := getVolume(svc, aws.StringValue(vol.Id))
					if err != nil {
						return err
					}
					vpcs[vk].Subnets[sk].EC2s[ik].Volumes[volk] = Volume{
						Id:         vol.Id,
						DeviceName: vol.DeviceName,
						Size:       volume.Size,
						VolumeType: volume.VolumeType,
						RawVolume:  volume,
					}
				}
			}
		}
	}
	return nil
}

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
	for _, vpc := range vpcs {

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
		for _, subnet := range vpc.Subnets {

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
			for _, instance := range subnet.EC2s {

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

func populateVPC(region string) (map[string]VPC, error) {
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String(region),
			},
		},
	))

	svc := ec2.New(sess)
	var data RecievedData
	vpcs := make(map[string]VPC)

	data.wg.Add(12)
	go getVpcs(svc, &data)
	go getSubnets(svc, &data)
	go getInstances(svc, &data)
	go getNatGatways(svc, &data)
	go getRouteTables(svc, &data)
	go getInternetGateways(svc, &data)
	go getEgressOnlyInternetGateways(svc, &data)
	go getVPNGateways(svc, &data)
	go getTransitGatewayVpcAttachments(svc, &data)
	go getVpcPeeringConnections(svc, &data)
	go getNetworkInterfaces(svc, &data)
	go getVpcEndpoints(svc, &data)

	data.wg.Wait()

	if data.Error != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", data.Error.Error())
	}

	mapVpcs(vpcs, data.Vpcs)
	mapSubnets(vpcs, data.Subnets)
	mapInstances(vpcs, data.Instances)
	err := instantiateVolumes(svc, vpcs)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapNatGateways(vpcs, data.NatGateways)
	mapRouteTables(vpcs, data.RouteTables)
	mapInternetGateways(vpcs, data.InternetGateways)
	mapEgressOnlyInternetGateways(vpcs, data.EOInternetGateways)
	mapVPNGateways(vpcs, data.VPNGateways)
	mapTransitGatewayVpcAttachments(vpcs, data.TransitGateways)
	mapVpcPeeringConnections(vpcs, data.PeeringConnections)
	mapNetworkInterfaces(vpcs, data.NetworkInterfaces)
	mapVpcEndpoints(vpcs, data.VPCEndpoints)

	return vpcs, nil
}

func getRegions() []string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)
	regions := []string{}
	res, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		panic("Could not get regions")
	}

	for _, region := range res.Regions {
		regions = append(regions, aws.StringValue(region.RegionName))
	}

	return regions
}

func getRegionData(fullData map[string]RegionData, region string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	vpcs, err := populateVPC(region)
	if err != nil {
		return
	}
	mu.Lock()
	fullData[region] = RegionData{
		VPCs: vpcs,
	}
	mu.Unlock()
}

func allRegions() {
	var wg sync.WaitGroup

	regions := getRegions()

	fullData := make(map[string]RegionData)
	mu := sync.Mutex{}

	for _, region := range regions {
		wg.Add(1)
		go getRegionData(fullData, region, &wg, &mu)
	}

	wg.Wait()

	for region, vpcs := range fullData {
		fmt.Printf("===%v===\n", region)
		printVPCs(vpcs.VPCs)
	}

}
func defaultRegion() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	currentRegion := aws.StringValue(sess.Config.Region)
	vpcs, err := populateVPC(currentRegion)
	if err != nil {
		panic("populateVPC failed")
	}

	printVPCs(vpcs)

}
func main() {
	//allRegions()
	defaultRegion()
}
