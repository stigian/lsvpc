package awsfetch

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func (f *AWSFetch) GetAll() (AWSRecieve, error) {
	go f.GetIdentity()
	go f.GetVpcs()
	go f.GetSubnets()
	go f.GetInstances()
	go f.GetInstanceStatuses()
	go f.GetNatGatways()
	go f.GetRouteTables()
	go f.GetInternetGateways()
	go f.GetEgressOnlyInternetGateways()
	go f.GetVPNGateways()
	go f.GetTransitGatewayVpcAttachments()
	go f.GetVpcPeeringConnections()
	go f.GetNetworkInterfaces()
	go f.GetSecurityGroups()
	go f.GetVpcEndpoints()
	go f.GetVolumes()
	return f.Recieve()
}

func (f *AWSFetch) GetIdentity() {
	res, err := f.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	f.Identity <- GetIdentityOutput{
		Identity: res,
		Err:      err,
	}
}

func (f *AWSFetch) GetVpcs() {
	vpcs := []*ec2.Vpc{}
	err := f.svc.DescribeVpcsPages(
		&ec2.DescribeVpcsInput{},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return !lastPage
		},
	)

	f.Vpcs <- GetVpcsOutput{
		Err:  err,
		Vpcs: vpcs,
	}
}

func (f *AWSFetch) GetSubnets() {
	subnets := []*ec2.Subnet{}
	err := f.svc.DescribeSubnetsPages(
		&ec2.DescribeSubnetsInput{},
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			subnets = append(subnets, page.Subnets...)
			return !lastPage
		},
	)

	f.Subnets <- GetSubnetsOutput{
		Subnets: subnets,
		Err:     err,
	}
}

func (f *AWSFetch) GetInstances() {
	instances := []*ec2.Reservation{}
	err := f.svc.DescribeInstancesPages(
		&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instances = append(instances, page.Reservations...)
			return !lastPage
		},
	)
	f.Instances <- GetInstancesOutput{
		Instances: instances,
		Err:       err,
	}
}

func (f *AWSFetch) GetInstanceStatuses() {
	statuses := []*ec2.InstanceStatus{}
	err := f.svc.DescribeInstanceStatusPages(
		&ec2.DescribeInstanceStatusInput{},
		func(page *ec2.DescribeInstanceStatusOutput, lastPage bool) bool {
			statuses = append(statuses, page.InstanceStatuses...)
			return !lastPage
		},
	)

	f.InstanceStatuses <- GetInstanceStatusOutput{
		InstanceStatuses: statuses,
		Err:              err,
	}
}

func (f *AWSFetch) GetNatGatways() {
	natGateways := []*ec2.NatGateway{}

	err := f.svc.DescribeNatGatewaysPages(
		&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			natGateways = append(natGateways, page.NatGateways...)
			return !lastPage
		},
	)
	f.NatGateways <- GetNatGatewaysOutput{
		NatGateways: natGateways,
		Err:         err,
	}
}

func (f *AWSFetch) GetRouteTables() {
	routeTables := []*ec2.RouteTable{}

	err := f.svc.DescribeRouteTablesPages(
		&ec2.DescribeRouteTablesInput{},
		func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, page.RouteTables...)
			return !lastPage
		},
	)
	f.RouteTables <- GetRouteTablesOutput{
		RouteTables: routeTables,
		Err:         err,
	}
}

func (f *AWSFetch) GetInternetGateways() {
	internetGateways := []*ec2.InternetGateway{}
	err := f.svc.DescribeInternetGatewaysPages(
		&ec2.DescribeInternetGatewaysInput{},
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, page.InternetGateways...)
			return !lastPage
		},
	)

	f.InternetGateways <- GetInternetGatewaysOutput{
		InternetGateways: internetGateways,
		Err:              err,
	}
}

func (f *AWSFetch) GetEgressOnlyInternetGateways() {
	EOIGWs := []*ec2.EgressOnlyInternetGateway{}

	err := f.svc.DescribeEgressOnlyInternetGatewaysPages(
		&ec2.DescribeEgressOnlyInternetGatewaysInput{},
		func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
			EOIGWs = append(EOIGWs, page.EgressOnlyInternetGateways...)
			return !lastPage
		},
	)
	f.EOInternetGateways <- GetEgressOnlyInternetGatewaysOutput{
		EOInternetGateways: EOIGWs,
		Err:                err,
	}
}

func (f *AWSFetch) GetVPNGateways() {
	res, err := f.svc.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})
	f.VPNGateways <- GetVPNGatewaysOutput{
		VPNGateways: res.VpnGateways,
		Err:         err,
	}
}

func (f *AWSFetch) GetTransitGatewayVpcAttachments() {
	TGWatt := []*ec2.TransitGatewayVpcAttachment{}

	err := f.svc.DescribeTransitGatewayVpcAttachmentsPages(
		&ec2.DescribeTransitGatewayVpcAttachmentsInput{},
		func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
			TGWatt = append(TGWatt, page.TransitGatewayVpcAttachments...)
			return !lastPage
		},
	)

	f.TransiGateways <- GetTransitGatewaysOutput{
		TransitGateways: TGWatt,
		Err:             err,
	}
}

func (f *AWSFetch) GetVpcPeeringConnections() {
	peers := []*ec2.VpcPeeringConnection{}

	err := f.svc.DescribeVpcPeeringConnectionsPages(
		&ec2.DescribeVpcPeeringConnectionsInput{},
		func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
			peers = append(peers, page.VpcPeeringConnections...)
			return !lastPage
		},
	)
	f.PeeringConnections <- GetPeeringConnectionsOutput{
		PeeringConnections: peers,
		Err:                err,
	}
}

func (f *AWSFetch) GetNetworkInterfaces() {
	ifaces := []*ec2.NetworkInterface{}

	err := f.svc.DescribeNetworkInterfacesPages(
		&ec2.DescribeNetworkInterfacesInput{},
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			ifaces = append(ifaces, page.NetworkInterfaces...)
			return !lastPage
		},
	)

	f.NetworkInterfaces <- GetNetworkInterfacesOutput{
		NetworkInterfaces: ifaces,
		Err:               err,
	}
}

func (f *AWSFetch) GetSecurityGroups() {
	sgs := []*ec2.SecurityGroup{}

	err := f.svc.DescribeSecurityGroupsPages(
		&ec2.DescribeSecurityGroupsInput{},
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			sgs = append(sgs, page.SecurityGroups...)
			return !lastPage
		},
	)
	f.SecurityGroups <- GetSecurityGroupsOutput{
		SecurityGroups: sgs,
		Err:            err,
	}
}

func (f *AWSFetch) GetVpcEndpoints() {
	endpoints := []*ec2.VpcEndpoint{}

	err := f.svc.DescribeVpcEndpointsPages(
		&ec2.DescribeVpcEndpointsInput{},
		func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
			endpoints = append(endpoints, page.VpcEndpoints...)
			return !lastPage
		},
	)

	f.VPCEndpoints <- GetVPCEndpointsOutput{
		VPCEndpoints: endpoints,
		Err:          err,
	}
}

func (f *AWSFetch) GetVolumes() {
	volumes := []*ec2.Volume{}

	err := f.svc.DescribeVolumesPages(
		&ec2.DescribeVolumesInput{},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			volumes = append(volumes, page.Volumes...)
			return !lastPage
		},
	)
	f.Volumes <- GetVolumesOutput{
		Volumes: volumes,
		Err:     err,
	}
}
