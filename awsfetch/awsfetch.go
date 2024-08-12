// Copyright 2024 Stigian Consulting - reference license in top level of project

// package awsfetch is a self-contained simple module to obtain all of the data
// needed by lsvpc in a parallel fashion. This is achieved through the use of
// unique channels that are defined for each unique sdk function queried.
package awsfetch

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

// AWSChan is a module within AWSFetch that contains the channels and services
// necessary to query and return data to the main AWSFetch struct.
// This is implemented using channels primarily as a test of the use of
// channels, but also to make this slightly easier to extend than a
// series of functions bounded by a wait group.
type AWSChan struct {
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

// AWSFetch is the primary struct used for obtaining the the data retrieved
// from the sdk functions. Each of the Output members are simply the sdk return
// types paired with an error value.
type AWSFetch struct {
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
	c                  AWSChan // located here to avoid lint alignment complaints
	SecurityGroups     GetSecurityGroupsOutput
	VPCEndpoints       GetVPCEndpointsOutput
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
	Err     error
	Subnets []*ec2.Subnet
}

type GetInstancesOutput struct {
	Err       error
	Instances []*ec2.Reservation
}

type GetInstanceStatusOutput struct {
	Err              error
	InstanceStatuses []*ec2.InstanceStatus
}

type GetVolumesOutput struct {
	Err     error
	Volumes []*ec2.Volume
}

type GetNatGatewaysOutput struct {
	Err         error
	NatGateways []*ec2.NatGateway
}

type GetRouteTablesOutput struct {
	Err         error
	RouteTables []*ec2.RouteTable
}

type GetInternetGatewaysOutput struct {
	Err              error
	InternetGateways []*ec2.InternetGateway
}

type GetEgressOnlyInternetGatewaysOutput struct {
	Err                error
	EOInternetGateways []*ec2.EgressOnlyInternetGateway
}

type GetVPNGatewaysOutput struct {
	Err         error
	VPNGateways []*ec2.VpnGateway
}

type GetTransitGatewaysOutput struct {
	Err             error
	TransitGateways []*ec2.TransitGatewayVpcAttachment
}

type GetPeeringConnectionsOutput struct {
	Err                error
	PeeringConnections []*ec2.VpcPeeringConnection
}

type GetNetworkInterfacesOutput struct {
	Err               error
	NetworkInterfaces []*ec2.NetworkInterface
}

type GetSecurityGroupsOutput struct {
	Err            error
	SecurityGroups []*ec2.SecurityGroup
}

type GetVPCEndpointsOutput struct {
	Err          error
	VPCEndpoints []*ec2.VpcEndpoint
}

// New initializes AWS Fetch and its internal AWSChan structs.
// channels need to be explicitly allocated with make().
func New(sess *session.Session) AWSFetch {
	f := AWSFetch{}
	f.c = AWSChan{}
	f.c.sts = sts.New(sess)
	f.c.svc = ec2.New(sess)
	f.c.Identity = make(chan GetIdentityOutput)
	f.c.Vpcs = make(chan GetVpcsOutput)
	f.c.Subnets = make(chan GetSubnetsOutput)
	f.c.Instances = make(chan GetInstancesOutput)
	f.c.InstanceStatuses = make(chan GetInstanceStatusOutput)
	f.c.Volumes = make(chan GetVolumesOutput)
	f.c.NatGateways = make(chan GetNatGatewaysOutput)
	f.c.RouteTables = make(chan GetRouteTablesOutput)
	f.c.InternetGateways = make(chan GetInternetGatewaysOutput)
	f.c.EOInternetGateways = make(chan GetEgressOnlyInternetGatewaysOutput)
	f.c.VPNGateways = make(chan GetVPNGatewaysOutput)
	f.c.TransiGateways = make(chan GetTransitGatewaysOutput)
	f.c.PeeringConnections = make(chan GetPeeringConnectionsOutput)
	f.c.NetworkInterfaces = make(chan GetNetworkInterfacesOutput)
	f.c.SecurityGroups = make(chan GetSecurityGroupsOutput)
	f.c.VPCEndpoints = make(chan GetVPCEndpointsOutput)

	return f
}

// GetAll spawns goroutines to concurrently request the aws sdk
// for all of the information relevant to a vpc. each function
// returns its results into their specific channel, and those
// channels are then read into the AWSFetch struct members.
func (f *AWSFetch) GetAll() (*AWSFetch, error) {
	go f.c.GetIdentity()
	go f.c.GetVpcs()
	go f.c.GetSubnets()
	go f.c.GetInstances()
	go f.c.GetInstanceStatuses()
	go f.c.GetNatGatways()
	go f.c.GetRouteTables()
	go f.c.GetInternetGateways()
	go f.c.GetEgressOnlyInternetGateways()
	go f.c.GetVPNGateways()
	go f.c.GetTransitGatewayVpcAttachments()
	go f.c.GetVpcPeeringConnections()
	go f.c.GetNetworkInterfaces()
	go f.c.GetSecurityGroups()
	go f.c.GetVpcEndpoints()
	go f.c.GetVolumes()

	f.Identity = <-f.c.Identity
	f.Vpcs = <-f.c.Vpcs
	f.Subnets = <-f.c.Subnets
	f.Instances = <-f.c.Instances
	f.InstanceStatuses = <-f.c.InstanceStatuses
	f.Volumes = <-f.c.Volumes
	f.NatGateways = <-f.c.NatGateways
	f.RouteTables = <-f.c.RouteTables
	f.InternetGateways = <-f.c.InternetGateways
	f.EOInternetGateways = <-f.c.EOInternetGateways
	f.VPNGateways = <-f.c.VPNGateways
	f.TransiGateways = <-f.c.TransiGateways
	f.PeeringConnections = <-f.c.PeeringConnections
	f.NetworkInterfaces = <-f.c.NetworkInterfaces
	f.SecurityGroups = <-f.c.SecurityGroups
	f.VPCEndpoints = <-f.c.VPCEndpoints

	err := f.Error()

	return f, err
}

func (f *AWSFetch) Error() error {
	if f.Identity.Err != nil {
		return f.Identity.Err
	}

	if f.Vpcs.Err != nil {
		return f.Vpcs.Err
	}

	if f.Subnets.Err != nil {
		return f.Subnets.Err
	}

	if f.Instances.Err != nil {
		return f.Instances.Err
	}

	if f.InstanceStatuses.Err != nil {
		return f.InstanceStatuses.Err
	}

	if f.Volumes.Err != nil {
		return f.Volumes.Err
	}

	if f.NatGateways.Err != nil {
		return f.NatGateways.Err
	}

	if f.RouteTables.Err != nil {
		return f.RouteTables.Err
	}

	if f.InternetGateways.Err != nil {
		return f.InternetGateways.Err
	}

	if f.EOInternetGateways.Err != nil {
		return f.EOInternetGateways.Err
	}

	if f.VPNGateways.Err != nil {
		return f.VPNGateways.Err
	}

	if f.TransiGateways.Err != nil {
		return f.TransiGateways.Err
	}

	if f.PeeringConnections.Err != nil {
		return f.PeeringConnections.Err
	}

	if f.NetworkInterfaces.Err != nil {
		return f.NetworkInterfaces.Err
	}

	if f.SecurityGroups.Err != nil {
		return f.SecurityGroups.Err
	}

	if f.VPCEndpoints.Err != nil {
		return f.VPCEndpoints.Err
	}

	return nil
}
