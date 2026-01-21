// Copyright 2023 Stigian Consulting - reference license in top level of project
package awsfetch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func (c *AWSChan) GetIdentity(ctx context.Context) {
	res, err := c.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})

	c.Identity <- GetIdentityOutput{
		Identity: res,
		Err:      err,
	}
}

func (c *AWSChan) GetVpcs(ctx context.Context) {
	vpcs := []types.Vpc{}
	paginator := ec2.NewDescribeVpcsPaginator(c.svc, &ec2.DescribeVpcsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		vpcs = append(vpcs, page.Vpcs...)
	}

	c.Vpcs <- GetVpcsOutput{
		Err:  err,
		Vpcs: vpcs,
	}
}

func (c *AWSChan) GetSubnets(ctx context.Context) {
	subnets := []types.Subnet{}
	paginator := ec2.NewDescribeSubnetsPaginator(c.svc, &ec2.DescribeSubnetsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		subnets = append(subnets, page.Subnets...)
	}

	c.Subnets <- GetSubnetsOutput{
		Subnets: subnets,
		Err:     err,
	}
}

func (c *AWSChan) GetInstances(ctx context.Context) {
	instances := []types.Reservation{}
	paginator := ec2.NewDescribeInstancesPaginator(c.svc, &ec2.DescribeInstancesInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		instances = append(instances, page.Reservations...)
	}

	c.Instances <- GetInstancesOutput{
		Instances: instances,
		Err:       err,
	}
}

func (c *AWSChan) GetInstanceStatuses(ctx context.Context) {
	statuses := []types.InstanceStatus{}
	paginator := ec2.NewDescribeInstanceStatusPaginator(c.svc, &ec2.DescribeInstanceStatusInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		statuses = append(statuses, page.InstanceStatuses...)
	}

	c.InstanceStatuses <- GetInstanceStatusOutput{
		InstanceStatuses: statuses,
		Err:              err,
	}
}

func (c *AWSChan) GetNatGatways(ctx context.Context) {
	natGateways := []types.NatGateway{}
	paginator := ec2.NewDescribeNatGatewaysPaginator(c.svc, &ec2.DescribeNatGatewaysInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		natGateways = append(natGateways, page.NatGateways...)
	}

	c.NatGateways <- GetNatGatewaysOutput{
		NatGateways: natGateways,
		Err:         err,
	}
}

func (c *AWSChan) GetRouteTables(ctx context.Context) {
	routeTables := []types.RouteTable{}
	paginator := ec2.NewDescribeRouteTablesPaginator(c.svc, &ec2.DescribeRouteTablesInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		routeTables = append(routeTables, page.RouteTables...)
	}

	c.RouteTables <- GetRouteTablesOutput{
		RouteTables: routeTables,
		Err:         err,
	}
}

func (c *AWSChan) GetInternetGateways(ctx context.Context) {
	internetGateways := []types.InternetGateway{}
	paginator := ec2.NewDescribeInternetGatewaysPaginator(c.svc, &ec2.DescribeInternetGatewaysInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		internetGateways = append(internetGateways, page.InternetGateways...)
	}

	c.InternetGateways <- GetInternetGatewaysOutput{
		InternetGateways: internetGateways,
		Err:              err,
	}
}

func (c *AWSChan) GetEgressOnlyInternetGateways(ctx context.Context) {
	EOIGWs := []types.EgressOnlyInternetGateway{}
	paginator := ec2.NewDescribeEgressOnlyInternetGatewaysPaginator(c.svc, &ec2.DescribeEgressOnlyInternetGatewaysInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		EOIGWs = append(EOIGWs, page.EgressOnlyInternetGateways...)
	}

	c.EOInternetGateways <- GetEgressOnlyInternetGatewaysOutput{
		EOInternetGateways: EOIGWs,
		Err:                err,
	}
}

func (c *AWSChan) GetVPNGateways(ctx context.Context) {
	res, err := c.svc.DescribeVpnGateways(ctx, &ec2.DescribeVpnGatewaysInput{})
	c.VPNGateways <- GetVPNGatewaysOutput{
		VPNGateways: res.VpnGateways,
		Err:         err,
	}
}

func (c *AWSChan) GetTransitGatewayVpcAttachments(ctx context.Context) {
	TGWatt := []types.TransitGatewayVpcAttachment{}
	paginator := ec2.NewDescribeTransitGatewayVpcAttachmentsPaginator(c.svc, &ec2.DescribeTransitGatewayVpcAttachmentsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		TGWatt = append(TGWatt, page.TransitGatewayVpcAttachments...)
	}

	c.TransiGateways <- GetTransitGatewaysOutput{
		TransitGateways: TGWatt,
		Err:             err,
	}
}

func (c *AWSChan) GetVpcPeeringConnections(ctx context.Context) {
	peers := []types.VpcPeeringConnection{}
	paginator := ec2.NewDescribeVpcPeeringConnectionsPaginator(c.svc, &ec2.DescribeVpcPeeringConnectionsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		peers = append(peers, page.VpcPeeringConnections...)
	}

	c.PeeringConnections <- GetPeeringConnectionsOutput{
		PeeringConnections: peers,
		Err:                err,
	}
}

func (c *AWSChan) GetNetworkInterfaces(ctx context.Context) {
	ifaces := []types.NetworkInterface{}
	paginator := ec2.NewDescribeNetworkInterfacesPaginator(c.svc, &ec2.DescribeNetworkInterfacesInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		ifaces = append(ifaces, page.NetworkInterfaces...)
	}

	c.NetworkInterfaces <- GetNetworkInterfacesOutput{
		NetworkInterfaces: ifaces,
		Err:               err,
	}
}

func (c *AWSChan) GetSecurityGroups(ctx context.Context) {
	sgs := []types.SecurityGroup{}
	paginator := ec2.NewDescribeSecurityGroupsPaginator(c.svc, &ec2.DescribeSecurityGroupsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		sgs = append(sgs, page.SecurityGroups...)
	}

	c.SecurityGroups <- GetSecurityGroupsOutput{
		SecurityGroups: sgs,
		Err:            err,
	}
}

func (c *AWSChan) GetVpcEndpoints(ctx context.Context) {
	endpoints := []types.VpcEndpoint{}
	paginator := ec2.NewDescribeVpcEndpointsPaginator(c.svc, &ec2.DescribeVpcEndpointsInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		endpoints = append(endpoints, page.VpcEndpoints...)
	}

	c.VPCEndpoints <- GetVPCEndpointsOutput{
		VPCEndpoints: endpoints,
		Err:          err,
	}
}

func (c *AWSChan) GetVolumes(ctx context.Context) {
	volumes := []types.Volume{}
	paginator := ec2.NewDescribeVolumesPaginator(c.svc, &ec2.DescribeVolumesInput{})
	
	var err error
	for paginator.HasMorePages() {
		page, pageErr := paginator.NextPage(ctx)
		if pageErr != nil {
			err = pageErr
			break
		}
		volumes = append(volumes, page.Volumes...)
	}

	c.Volumes <- GetVolumesOutput{
		Volumes: volumes,
		Err:     err,
	}
}
