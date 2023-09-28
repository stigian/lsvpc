// Copyright 2023 Stigian Consulting - reference license in top level of project
package awsfetch

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AWSFetch struct {
	Identity           chan GetIdentityOutput
	Vpcs               chan GetVpcsOutput
	Subnets            chan GetSubnetsOutput
	Instances          chan GetInstancesOutput
	InstanceStatuses   chan GetInstanceStatusOutput
	Volumes            chan GetVolumesOutput
	NatGateways        chan GetNatGatewaysOutput
	RouteTables        chan GetRouteTablesOutput
	InternetGateways   chan GetInternetGatewaysOutput
	EOInternetGateways chan GetEgressOnlyInternetGatewaysOutput
	VPNGateways        chan GetVPNGatewaysOutput
	TransiGateways     chan GetTransitGatewaysOutput
	PeeringConnections chan GetPeeringConnectionsOutput
	NetworkInterfaces  chan GetNetworkInterfacesOutput
	SecurityGroups     chan GetSecurityGroupsOutput
	VPCEndpoints       chan GetVPCEndpointsOutput
	svc                *ec2.EC2
	sts                *sts.STS
}

type AWSRecieve struct {
	Identity           GetIdentityOutput
	Vpcs               GetVpcsOutput
	Subnets            GetSubnetsOutput
	Instances          GetInstancesOutput
	InstanceStatuses   GetInstanceStatusOutput
	Volumes            GetVolumesOutput
	NatGateways        GetNatGatewaysOutput
	RouteTables        GetRouteTablesOutput
	InternetGateways   GetInternetGatewaysOutput
	EOInternetGateways GetEgressOnlyInternetGatewaysOutput
	VPNGateways        GetVPNGatewaysOutput
	TransiGateways     GetTransitGatewaysOutput
	PeeringConnections GetPeeringConnectionsOutput
	NetworkInterfaces  GetNetworkInterfacesOutput
	SecurityGroups     GetSecurityGroupsOutput
	VPCEndpoints       GetVPCEndpointsOutput
}

func New(sess *session.Session) AWSFetch {
	a := AWSFetch{}
	a.sts = sts.New(sess)
	a.svc = ec2.New(sess)
	a.Identity = make(chan GetIdentityOutput)
	a.Vpcs = make(chan GetVpcsOutput)
	a.Subnets = make(chan GetSubnetsOutput)
	a.Instances = make(chan GetInstancesOutput)
	a.InstanceStatuses = make(chan GetInstanceStatusOutput)
	a.Volumes = make(chan GetVolumesOutput)
	a.NatGateways = make(chan GetNatGatewaysOutput)
	a.RouteTables = make(chan GetRouteTablesOutput)
	a.InternetGateways = make(chan GetInternetGatewaysOutput)
	a.EOInternetGateways = make(chan GetEgressOnlyInternetGatewaysOutput)
	a.VPNGateways = make(chan GetVPNGatewaysOutput)
	a.TransiGateways = make(chan GetTransitGatewaysOutput)
	a.PeeringConnections = make(chan GetPeeringConnectionsOutput)
	a.NetworkInterfaces = make(chan GetNetworkInterfacesOutput)
	a.SecurityGroups = make(chan GetSecurityGroupsOutput)
	a.VPCEndpoints = make(chan GetVPCEndpointsOutput)
	return a
}

func (f *AWSFetch) GetAll() (AWSRecieve, error) {
	r := AWSRecieve{}
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

	r.Identity = <-f.Identity
	r.Vpcs = <-f.Vpcs
	r.Subnets = <-f.Subnets
	r.Instances = <-f.Instances
	r.InstanceStatuses = <-f.InstanceStatuses
	r.Volumes = <-f.Volumes
	r.NatGateways = <-f.NatGateways
	r.RouteTables = <-f.RouteTables
	r.InternetGateways = <-f.InternetGateways
	r.EOInternetGateways = <-f.EOInternetGateways
	r.VPNGateways = <-f.VPNGateways
	r.TransiGateways = <-f.TransiGateways
	r.PeeringConnections = <-f.PeeringConnections
	r.NetworkInterfaces = <-f.NetworkInterfaces
	r.SecurityGroups = <-f.SecurityGroups
	r.VPCEndpoints = <-f.VPCEndpoints
	return r, r.Error()
}

func (a *AWSRecieve) Error() error {
	if a.Identity.Err != nil {
		return a.Identity.Err
	}
	if a.Vpcs.Err != nil {
		return a.Vpcs.Err
	}
	if a.Subnets.Err != nil {
		return a.Subnets.Err
	}
	if a.Instances.Err != nil {
		return a.Instances.Err
	}
	if a.InstanceStatuses.Err != nil {
		return a.InstanceStatuses.Err
	}
	if a.Volumes.Err != nil {
		return a.Volumes.Err
	}
	if a.NatGateways.Err != nil {
		return a.NatGateways.Err
	}
	if a.RouteTables.Err != nil {
		return a.RouteTables.Err
	}
	if a.InternetGateways.Err != nil {
		return a.InternetGateways.Err
	}
	if a.EOInternetGateways.Err != nil {
		return a.EOInternetGateways.Err
	}
	if a.VPNGateways.Err != nil {
		return a.VPNGateways.Err
	}
	if a.TransiGateways.Err != nil {
		return a.TransiGateways.Err
	}
	if a.PeeringConnections.Err != nil {
		return a.PeeringConnections.Err
	}
	if a.NetworkInterfaces.Err != nil {
		return a.NetworkInterfaces.Err
	}
	if a.SecurityGroups.Err != nil {
		return a.SecurityGroups.Err
	}
	if a.VPCEndpoints.Err != nil {
		return a.VPCEndpoints.Err
	}
	return nil
}

type GetIdentityOutput struct {
	Identity *sts.GetCallerIdentityOutput
	Err      error
}

type GetVpcsOutput struct {
	Err  error
	Vpcs []*ec2.Vpc
}

type GetSubnetsOutput struct {
	Subnets []*ec2.Subnet
	Err     error
}

type GetInstancesOutput struct {
	Instances []*ec2.Reservation
	Err       error
}

type GetInstanceStatusOutput struct {
	InstanceStatuses []*ec2.InstanceStatus
	Err              error
}

type GetVolumesOutput struct {
	Volumes []*ec2.Volume
	Err     error
}

type GetNatGatewaysOutput struct {
	NatGateways []*ec2.NatGateway
	Err         error
}

type GetRouteTablesOutput struct {
	RouteTables []*ec2.RouteTable
	Err         error
}

type GetInternetGatewaysOutput struct {
	InternetGateways []*ec2.InternetGateway
	Err              error
}

type GetEgressOnlyInternetGatewaysOutput struct {
	EOInternetGateways []*ec2.EgressOnlyInternetGateway
	Err                error
}

type GetVPNGatewaysOutput struct {
	VPNGateways []*ec2.VpnGateway
	Err         error
}

type GetTransitGatewaysOutput struct {
	TransitGateways []*ec2.TransitGatewayVpcAttachment
	Err             error
}

type GetPeeringConnectionsOutput struct {
	PeeringConnections []*ec2.VpcPeeringConnection
	Err                error
}

type GetNetworkInterfacesOutput struct {
	NetworkInterfaces []*ec2.NetworkInterface
	Err               error
}

type GetSecurityGroupsOutput struct {
	SecurityGroups []*ec2.SecurityGroup
	Err            error
}

type GetVPCEndpointsOutput struct {
	VPCEndpoints []*ec2.VpcEndpoint
	Err          error
}
