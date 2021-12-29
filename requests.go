// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

func getIdentity(svc *sts.STS, data *RecievedData) {
	defer data.wg.Done()
	out, err := svc.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}
	data.Identity = out
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

func getInstanceStatuses(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	statuses := []*ec2.InstanceStatus{}
	err := svc.DescribeInstanceStatusPages(
		&ec2.DescribeInstanceStatusInput{},
		func(page *ec2.DescribeInstanceStatusOutput, lastPage bool) bool {
			statuses = append(statuses, page.InstanceStatuses...)
			return !lastPage
		},
	)

	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}

	data.InstanceStatuses = statuses
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

func getVolumes(svc *ec2.EC2, data *RecievedData) {
	defer data.wg.Done()
	volumes := []*ec2.Volume{}

	err := svc.DescribeVolumesPages(
		&ec2.DescribeVolumesInput{},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			volumes = append(volumes, page.Volumes...)
			return !lastPage
		},
	)
	if err != nil {
		data.mu.Lock()
		data.Error = err
		data.mu.Unlock()
	}
	data.Volumes = volumes
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
