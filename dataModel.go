// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"sync"

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
	Name          *string
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
	Name               *string
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
	Id             *string
	Type           *string
	SubnetId       *string
	VpcId          *string
	State          *string
	PublicIP       *string
	PrivateIP      *string
	Name           *string
	InstanceStatus *string
	SystemStatus   *string
	Volumes        map[string]Volume
	Interfaces     map[string]NetworkInterface
	RawEc2         *ec2.Instance
}

type NetworkInterface struct {
	Id                  *string
	PrivateIp           *string
	MAC                 *string
	DNS                 *string
	Type                *string
	Description         *string
	PublicIp            *string
	Name                *string
	RawNetworkInterface *ec2.NetworkInterface
}

type Volume struct {
	Id         *string
	DeviceName *string
	Size       *int64
	VolumeType *string
	Name       *string
	RawVolume  *ec2.Volume
}
type NatGateway struct {
	Id            *string
	PrivateIP     *string
	PublicIP      *string
	State         *string
	Type          *string
	Name          *string
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
	Name             *string
	RawAttachment    *ec2.TransitGatewayVpcAttachment
}

type VPCPeer struct {
	Id        *string
	Requester *string
	Accepter  *string
	Name      *string
	RawPeer   *ec2.VpcPeeringConnection
}

type InterfaceEndpoint struct {
	Id          *string
	ServiceName *string
	Name        *string
	RawEndpoint *ec2.VpcEndpoint
}

type GatewayEndpoint struct {
	Id          *string
	ServiceName *string
	Name        *string
	RawEndpoint *ec2.VpcEndpoint
}

type RecievedData struct {
	wg                 sync.WaitGroup
	mu                 sync.Mutex
	Vpcs               []*ec2.Vpc
	Subnets            []*ec2.Subnet
	Instances          []*ec2.Reservation
	InstanceStatuses   []*ec2.InstanceStatus
	NatGateways        []*ec2.NatGateway
	RouteTables        []*ec2.RouteTable
	InternetGateways   []*ec2.InternetGateway
	EOInternetGateways []*ec2.EgressOnlyInternetGateway
	VPNGateways        []*ec2.VpnGateway
	TransitGateways    []*ec2.TransitGatewayVpcAttachment
	PeeringConnections []*ec2.VpcPeeringConnection
	NetworkInterfaces  []*ec2.NetworkInterface
	VPCEndpoints       []*ec2.VpcEndpoint
	Volumes            []*ec2.Volume
	Error              error
}
