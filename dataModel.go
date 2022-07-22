// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

type RegionData struct {
	VPCs map[string]VPC
}

type VPCData struct {
	Id            string
	IsDefault     bool
	CidrBlock     string
	IPv6CidrBlock string
	Name          string
	RawVPC        *ec2.Vpc
}

type VPCSorted struct {
	VPCData
	Gateways []string
	Subnets  []SubnetSorted
	Peers    []VPCPeer
}

type VPC struct {
	VPCData
	Gateways []string
	Subnets  map[string]Subnet
	Peers    map[string]VPCPeer
}

type SubnetData struct {
	Id                 string
	CidrBlock          string
	AvailabilityZone   string
	AvailabilityZoneId string
	Public             bool
	Name               string
	RawSubnet          *ec2.Subnet
	RouteTable         *RouteTable
}

type SubnetSorted struct {
	SubnetData
	Instances          []InstanceSorted
	NatGateways        []NatGateway
	TGWs               []TGWAttachment
	ENIs               []NetworkInterface
	InterfaceEndpoints []InterfaceEndpoint
	GatewayEndpoints   []GatewayEndpoint
}

type Subnet struct {
	SubnetData
	Instances          map[string]Instance
	NatGateways        map[string]NatGateway
	TGWs               map[string]TGWAttachment
	ENIs               map[string]NetworkInterface
	InterfaceEndpoints map[string]InterfaceEndpoint
	GatewayEndpoints   map[string]GatewayEndpoint
}

type InstanceData struct {
	Id             string
	Type           string
	SubnetId       string
	VpcId          string
	State          string
	PublicIP       string
	PrivateIP      string
	Name           string
	InstanceStatus string
	SystemStatus   string
	RawEc2         *ec2.Instance
}

type InstanceSorted struct {
	InstanceData
	Volumes    []Volume
	Interfaces []NetworkInterface
}

type Instance struct {
	InstanceData
	Volumes    map[string]Volume
	Interfaces map[string]NetworkInterface
}

type NetworkInterface struct {
	Id                  string
	PrivateIp           string
	MAC                 string
	DNS                 string
	Type                string
	Description         string
	PublicIp            string
	Name                string
	RawNetworkInterface *ec2.NetworkInterface
}

type Volume struct {
	Id         *string
	DeviceName *string
	Size       *int64
	VolumeType *string
	Name       string
	RawVolume  *ec2.Volume
}
type NatGateway struct {
	Id            *string
	PrivateIP     *string
	PublicIP      *string
	State         *string
	Type          *string
	Name          string
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
	Name             string
	RawAttachment    *ec2.TransitGatewayVpcAttachment
}

type VPCPeer struct {
	Id        *string
	Requester *string
	Accepter  *string
	Name      string
	RawPeer   *ec2.VpcPeeringConnection
}

type InterfaceEndpoint struct {
	Id          *string
	ServiceName *string
	Name        string
	RawEndpoint *ec2.VpcEndpoint
}

type GatewayEndpoint struct {
	Id          *string
	ServiceName *string
	Name        string
	RawEndpoint *ec2.VpcEndpoint
}

type RecievedData struct {
	wg                 sync.WaitGroup
	mu                 sync.Mutex
	Identity           *sts.GetCallerIdentityOutput
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
