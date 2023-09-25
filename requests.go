// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getIdentity(svc *sts.STS, out chan GetIdentityOutput) {
	defer close(out)
	res, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})

	out <- GetIdentityOutput{
		Identity: res,
		Err:      err,
	}
}

func getVpcs(svc *ec2.EC2, out chan GetVpcsOutput) {
	defer close(out)
	vpcs := []*ec2.Vpc{}
	err := svc.DescribeVpcsPages(
		&ec2.DescribeVpcsInput{},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpcs = append(vpcs, page.Vpcs...)
			return !lastPage
		},
	)

	out <- GetVpcsOutput{
		Err:  err,
		Vpcs: vpcs,
	}
}

func getSubnets(svc *ec2.EC2, out chan GetSubnetsOutput) {
	defer close(out)
	subnets := []*ec2.Subnet{}
	err := svc.DescribeSubnetsPages(
		&ec2.DescribeSubnetsInput{},
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			subnets = append(subnets, page.Subnets...)
			return !lastPage
		},
	)

	out <- GetSubnetsOutput{
		Subnets: subnets,
		Err:     err,
	}
}

func getInstances(svc *ec2.EC2, out chan GetInstancesOutput) {
	defer close(out)
	instances := []*ec2.Reservation{}
	err := svc.DescribeInstancesPages(
		&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instances = append(instances, page.Reservations...)
			return !lastPage
		},
	)
	out <- GetInstancesOutput{
		Instances: instances,
		Err:       err,
	}
}

func getInstanceStatuses(svc *ec2.EC2, out chan GetInstanceStatusOutput) {
	defer close(out)
	statuses := []*ec2.InstanceStatus{}
	err := svc.DescribeInstanceStatusPages(
		&ec2.DescribeInstanceStatusInput{},
		func(page *ec2.DescribeInstanceStatusOutput, lastPage bool) bool {
			statuses = append(statuses, page.InstanceStatuses...)
			return !lastPage
		},
	)

	out <- GetInstanceStatusOutput{
		InstanceStatuses: statuses,
		Err:              err,
	}
}

func getNatGatways(svc *ec2.EC2, out chan GetNatGatewaysOutput) {
	defer close(out)
	natGateways := []*ec2.NatGateway{}

	err := svc.DescribeNatGatewaysPages(
		&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			natGateways = append(natGateways, page.NatGateways...)
			return !lastPage
		},
	)
	out <- GetNatGatewaysOutput{
		NatGateways: natGateways,
		Err:         err,
	}
}

func getRouteTables(svc *ec2.EC2, out chan GetRouteTablesOutput) {
	defer close(out)
	routeTables := []*ec2.RouteTable{}

	err := svc.DescribeRouteTablesPages(
		&ec2.DescribeRouteTablesInput{},
		func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, page.RouteTables...)
			return !lastPage
		},
	)
	out <- GetRouteTablesOutput{
		RouteTables: routeTables,
		Err:         err,
	}
}

func getInternetGateways(svc *ec2.EC2, out chan GetInternetGatewaysOutput) {
	defer close(out)
	internetGateways := []*ec2.InternetGateway{}
	err := svc.DescribeInternetGatewaysPages(
		&ec2.DescribeInternetGatewaysInput{},
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, page.InternetGateways...)
			return !lastPage
		},
	)

	out <- GetInternetGatewaysOutput{
		InternetGateways: internetGateways,
		Err:              err,
	}
}

func getEgressOnlyInternetGateways(svc *ec2.EC2, out chan GetEgressOnlyInternetGatewaysOutput) {
	defer close(out)

	EOIGWs := []*ec2.EgressOnlyInternetGateway{}

	err := svc.DescribeEgressOnlyInternetGatewaysPages(
		&ec2.DescribeEgressOnlyInternetGatewaysInput{},
		func(page *ec2.DescribeEgressOnlyInternetGatewaysOutput, lastPage bool) bool {
			EOIGWs = append(EOIGWs, page.EgressOnlyInternetGateways...)
			return !lastPage
		},
	)
	out <- GetEgressOnlyInternetGatewaysOutput{
		EOInternetGateways: EOIGWs,
		Err:                err,
	}
}

func getVPNGateways(svc *ec2.EC2, out chan GetVPNGatewaysOutput) {
	defer close(out)

	res, err := svc.DescribeVpnGateways(&ec2.DescribeVpnGatewaysInput{})
	out <- GetVPNGatewaysOutput{
		VPNGateways: res.VpnGateways,
		Err:         err,
	}
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

func getSecurityGroups(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()

	sgs := []*ec2.SecurityGroup{}

	err := svc.DescribeSecurityGroupsPages(
		&ec2.DescribeSecurityGroupsInput{},
		func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
			sgs = append(sgs, page.SecurityGroups...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.SecurityGroups = sgs
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

func getVolumes(svc *ec2.EC2, out chan GetVolumesOutput) {
	volumes := []*ec2.Volume{}

	err := svc.DescribeVolumesPages(
		&ec2.DescribeVolumesInput{},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			volumes = append(volumes, page.Volumes...)
			return !lastPage
		},
	)
	out <- GetVolumesOutput{
		Volumes: volumes,
		Err:     err,
	}
	close(out)
}

func getRegions() []string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)
	regions := []string{}

	res, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		panic(fmt.Sprintf("Could not get regions: %v", err.Error()))
	}

	for _, region := range res.Regions {
		regions = append(regions, aws.StringValue(region.RegionName))
	}

	return regions
}
