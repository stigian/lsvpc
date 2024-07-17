// Copyright 2023 Stigian Consulting - reference license in top level of project
package awsfetch

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func (c *AWSChan) GetIdentity() {
	res, err := c.sts.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	c.Identity <- GetIdentityOutput{
		Identity: res,
		Err:      err,
	}
}

func (c *AWSChan) GetVpcs() {
	vpcs := []*ec2.Vpc{}
	err := c.svc.DescribeVpcsPages(
		&ec2.DescribeVpcsInput{},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return !lastPage
		},
	)

	c.Vpcs <- GetVpcsOutput{
		Err:  err,
		Vpcs: vpcs,
	}
}

func (c *AWSChan) GetSubnets() {
	subnets := []*ec2.Subnet{}
	err := c.svc.DescribeSubnetsPages(
		&ec2.DescribeSubnetsInput{},
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			subnets = append(subnets, page.Subnets...)
			return !lastPage
		},
	)

	c.Subnets <- GetSubnetsOutput{
		Subnets: subnets,
		Err:     err,
	}
}

func (c *AWSChan) GetInstances() {
	instances := []*ec2.Reservation{}
	err := c.svc.DescribeInstancesPages(
		&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instances = append(instances, page.Reservations...)
			return !lastPage
		},
	)
	c.Instances <- GetInstancesOutput{
		Instances: instances,
		Err:       err,
	}
}

func (c *AWSChan) GetInstanceStatuses() {
	statuses := []*ec2.InstanceStatus{}
	err := c.svc.DescribeInstanceStatusPages(
		&ec2.DescribeInstanceStatusInput{},
		func(page *ec2.DescribeInstanceStatusOutput, lastPage bool) bool {
			statuses = append(statuses, page.InstanceStatuses...)
			return !lastPage
		},
	)

	c.InstanceStatuses <- GetInstanceStatusOutput{
		InstanceStatuses: statuses,
		Err:              err,
	}
}

func (c *AWSChan) GetNatGatways() {
	natGateways := []*ec2.NatGateway{}

	err := c.svc.DescribeNatGatewaysPages(
		&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			natGateways = append(natGateways, page.NatGateways...)
			return !lastPage
		},
	)
	c.NatGateways <- GetNatGatewaysOutput{
		NatGateways: natGateways,
		Err:         err,
	}
}

func (c *AWSChan) GetRouteTables() {
	routeTables := []*ec2.RouteTable{}

	err := c.svc.DescribeRouteTablesPages(
		&ec2.DescribeRouteTablesInput{},
		func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, page.RouteTables...)
			return !lastPage
		},
	)
	c.RouteTables <- GetRouteTablesOutput{
		RouteTables: routeTables,
		Err:         err,
	}
}

func (c *AWSChan) GetInternetGateways() {
	internetGateways := []*ec2.InternetGateway{}
	err := c.svc.DescribeInternetGatewaysPages(
		&ec2.DescribeInternetGatewaysInput{},
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, page.InternetGateways...)
			return !lastPage
		},
	)

	c.InternetGateways <- GetInternetGatewaysOutput{
		InternetGateways: internetGateways,
		Err:              err,
	}
}

func (c *AWSChan) GetEgressOnlyInternetGateways() {
	EOIGWs := []*ec2.EgressOnlyInternetGateway{}

	err := c.svc.DescribeEgressOnlyInternetGatewaysPages(
		&ec2.DescribeEgressOnlyInternetGatewaysInput{},
		func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
			EOIGWs = append(EOIGWs, page.EgressOnlyInternetGateways...)
			return !lastPage
		},
	)
	c.EOInternetGateways <- GetEgressOnlyInternetGatewaysOutput{
		EOInternetGateways: EOIGWs,
		Err:                err,
	}
}

func (c *AWSChan) GetVPNGateways() {
	res, err := c.svc.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})
	c.VPNGateways <- GetVPNGatewaysOutput{
		VPNGateways: res.VpnGateways,
		Err:         err,
	}
}

func (c *AWSChan) GetTransitGatewayVpcAttachments() {
	TGWatt := []*ec2.TransitGatewayVpcAttachment{}

	err := c.svc.DescribeTransitGatewayVpcAttachmentsPages(
		&ec2.DescribeTransitGatewayVpcAttachmentsInput{},
		func(page *ec2.DescribeTransitGatewayVpcAttachmentsOutput, lastPage bool) bool {
			TGWatt = append(TGWatt, page.TransitGatewayVpcAttachments...)
			return !lastPage
		},
	)

	c.TransiGateways <- GetTransitGatewaysOutput{
		TransitGateways: TGWatt,
		Err:             err,
	}
}

func (c *AWSChan) GetVpcPeeringConnections() {
	peers := []*ec2.VpcPeeringConnection{}

	err := c.svc.DescribeVpcPeeringConnectionsPages(
		&ec2.DescribeVpcPeeringConnectionsInput{},
		func(page *ec2.DescribeVpcPeeringConnectionsOutput, lastPage bool) bool {
			peers = append(peers, page.VpcPeeringConnections...)
			return !lastPage
		},
	)
	c.PeeringConnections <- GetPeeringConnectionsOutput{
		PeeringConnections: peers,
		Err:                err,
	}
}

func (c *AWSChan) GetNetworkInterfaces() {
	ifaces := []*ec2.NetworkInterface{}

	err := c.svc.DescribeNetworkInterfacesPages(
		&ec2.DescribeNetworkInterfacesInput{},
		func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
			ifaces = append(ifaces, page.NetworkInterfaces...)
			return !lastPage
		},
	)

	c.NetworkInterfaces <- GetNetworkInterfacesOutput{
		NetworkInterfaces: ifaces,
		Err:               err,
	}
}

func (c *AWSChan) GetSecurityGroups() {
	sgs := []*ec2.SecurityGroup{}

	err := c.svc.DescribeSecurityGroupsPages(
		&ec2.DescribeSecurityGroupsInput{},
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			sgs = append(sgs, page.SecurityGroups...)
			return !lastPage
		},
	)
	c.SecurityGroups <- GetSecurityGroupsOutput{
		SecurityGroups: sgs,
		Err:            err,
	}
}

func (c *AWSChan) GetVpcEndpoints() {
	endpoints := []*ec2.VpcEndpoint{}

	err := c.svc.DescribeVpcEndpointsPages(
		&ec2.DescribeVpcEndpointsInput{},
		func(page *ec2.DescribeVpcEndpointsOutput, lastPage bool) bool {
			endpoints = append(endpoints, page.VpcEndpoints...)
			return !lastPage
		},
	)

	c.VPCEndpoints <- GetVPCEndpointsOutput{
		VPCEndpoints: endpoints,
		Err:          err,
	}
}

func (c *AWSChan) GetVolumes() {
	volumes := []*ec2.Volume{}

	err := c.svc.DescribeVolumesPages(
		&ec2.DescribeVolumesInput{},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			volumes = append(volumes, page.Volumes...)
			return !lastPage
		},
	)
	c.Volumes <- GetVolumesOutput{
		Volumes: volumes,
		Err:     err,
	}
}
