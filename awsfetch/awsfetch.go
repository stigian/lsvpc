// Copyright 2024 Stigian Consulting - reference license in top level of project

// package awsfetch is a self-contained simple module to obtain all of the data
// needed by lsvpc in a parallel fashion. This is achieved through the use of
// unique channels that are defined for each unique sdk function queried.
package awsfetch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
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
	svc                *ec2.Client
	sts                *sts.Client
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
	Vpcs []types.Vpc
}

type GetSubnetsOutput struct {
	Err     error
	Subnets []types.Subnet
}

type GetInstancesOutput struct {
	Err       error
	Instances []types.Reservation
}

type GetInstanceStatusOutput struct {
	Err              error
	InstanceStatuses []types.InstanceStatus
}

type GetVolumesOutput struct {
	Err     error
	Volumes []types.Volume
}

type GetNatGatewaysOutput struct {
	Err         error
	NatGateways []types.NatGateway
}

type GetRouteTablesOutput struct {
	Err         error
	RouteTables []types.RouteTable
}

type GetInternetGatewaysOutput struct {
	Err              error
	InternetGateways []types.InternetGateway
}

type GetEgressOnlyInternetGatewaysOutput struct {
	Err                error
	EOInternetGateways []types.EgressOnlyInternetGateway
}

type GetVPNGatewaysOutput struct {
	Err         error
	VPNGateways []types.VpnGateway
}

type GetTransitGatewaysOutput struct {
	Err             error
	TransitGateways []types.TransitGatewayVpcAttachment
}

type GetPeeringConnectionsOutput struct {
	Err                error
	PeeringConnections []types.VpcPeeringConnection
}

type GetNetworkInterfacesOutput struct {
	Err               error
	NetworkInterfaces []types.NetworkInterface
}

type GetSecurityGroupsOutput struct {
	Err            error
	SecurityGroups []types.SecurityGroup
}

type GetVPCEndpointsOutput struct {
	Err          error
	VPCEndpoints []types.VpcEndpoint
}

// New initializes AWS Fetch and its internal AWSChan structs.
// channels need to be explicitly allocated with make().
func New(cfg aws.Config) AWSFetch {
	f := AWSFetch{}
	f.c = AWSChan{}
	f.c.sts = sts.NewFromConfig(cfg)
	f.c.svc = ec2.NewFromConfig(cfg)
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
func (f *AWSFetch) GetAll(ctx context.Context) (*AWSFetch, error) {
	go f.c.GetIdentity(ctx)
	go f.c.GetVpcs(ctx)
	go f.c.GetSubnets(ctx)
	go f.c.GetInstances(ctx)
	go f.c.GetInstanceStatuses(ctx)
	go f.c.GetNatGatways(ctx)
	go f.c.GetRouteTables(ctx)
	go f.c.GetInternetGateways(ctx)
	go f.c.GetEgressOnlyInternetGateways(ctx)
	go f.c.GetVPNGateways(ctx)
	go f.c.GetTransitGatewayVpcAttachments(ctx)
	go f.c.GetVpcPeeringConnections(ctx)
	go f.c.GetNetworkInterfaces(ctx)
	go f.c.GetSecurityGroups(ctx)
	go f.c.GetVpcEndpoints(ctx)
	go f.c.GetVolumes(ctx)

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
