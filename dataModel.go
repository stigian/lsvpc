// Copyright 2021 Stigian Consulting - reference license in top level of project
package main

import (
	"sync"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

type RegionData struct {
	VPCs map[string]*VPC
}

type RegionDataSorted struct {
	Region string `json:"region"`
	VPCs   []*VPCSorted
}

type VPCData struct {
	ID            string `json:"id"`
	CidrBlock     string `json:"cidrBlock"`
	IPv6CidrBlock string `json:"iPv6CidrBlock"`
	Name          string `json:"name"`
	IsDefault     bool   `json:"isDefault"`
}

type VPCSorted struct {
	VPCData
	Gateways []string        `json:"gateways,omitempty"`
	Subnets  []*SubnetSorted `json:"subnets,omitempty"`
	Peers    []*VPCPeer      `json:"peers,omitempty"`
}

type VPC struct {
	VPCData
	RawVPC   *ec2.Vpc
	Subnets  map[string]*Subnet
	Peers    map[string]*VPCPeer
	Gateways []string
}

type SubnetData struct {
	RouteTable         *RouteTable
	ID                 string `json:"id"`
	CidrBlock          string `json:"cidrBlock"`
	AvailabilityZone   string `json:"availabilityZone"`
	AvailabilityZoneID string `json:"availabilityZoneId"`
	Name               string `json:"name"`
	Public             bool   `json:"public"`
}

type SubnetSorted struct {
	SubnetData
	Instances          []*InstanceSorted          `json:"instances,omitempty"`
	NatGateways        []*NatGateway              `json:"natGateways,omitempty"`
	TGWs               []*TGWAttachment           `json:"tgws,omitempty"`
	ENIs               []*NetworkInterface        `json:"enis,omitempty"`
	InterfaceEndpoints []*InterfaceEndpointSorted `json:"interfaceEndpoints,omitempty"`
	GatewayEndpoints   []*GatewayEndpoint         `json:"gatewayEndpoints,omitempty"`
}

type Subnet struct {
	RawSubnet          *ec2.Subnet
	Instances          map[string]*Instance
	NatGateways        map[string]*NatGateway
	TGWs               map[string]*TGWAttachment
	ENIs               map[string]*NetworkInterface
	InterfaceEndpoints map[string]*InterfaceEndpoint
	GatewayEndpoints   map[string]*GatewayEndpoint
	SubnetData
}

type InstanceData struct {
	ID             string `json:"id"`
	Type           string `json:"type"`
	SubnetID       string `json:"subnetId"`
	VpcID          string `json:"vpcId"`
	State          string `json:"state"`
	PublicIP       string `json:"publicIP"`
	PrivateIP      string `json:"privateIP"`
	Name           string `json:"name"`
	InstanceStatus string `json:"instanceStatus"`
	SystemStatus   string `json:"systemStatus"`
}

type InstanceSorted struct {
	InstanceData
	Volumes    []*Volume                 `json:"volumes,omitempty"`
	Interfaces []*NetworkInterfaceSorted `json:"interfaces,omitempty"`
}

type Instance struct {
	Volumes    map[string]*Volume
	Interfaces map[string]*NetworkInterface
	RawEc2     *ec2.Instance
	InstanceData
}

type NetworkInterfaceData struct {
	RawNetworkInterface *ec2.NetworkInterface `json:"-"`
	ID                  string                `json:"id"`
	PrivateIP           string                `json:"privateIp"`
	MAC                 string                `json:"mAC"`
	DNS                 string                `json:"dNS"`
	Type                string                `json:"type"`
	Description         string                `json:"description"`
	PublicIP            string                `json:"publicIp"`
	Name                string                `json:"name"`
	SubnetID            string                `json:"subnetId"` // we're just accounting for this for display purposes
}

type NetworkInterface struct {
	Groups map[string]*SecurityGroup `json:"groups"`
	NetworkInterfaceData
}

type NetworkInterfaceSorted struct {
	NetworkInterfaceData
	Groups []*SecurityGroup `json:"groups"`
}

type SecurityGroup struct {
	RawSecurityGroup    *ec2.SecurityGroup   `json:"-"`
	Description         string               `json:"description"`
	GroupID             string               `json:"groupId"`
	GroupName           string               `json:"groupName"`
	TagName             string               `json:"tagName"`
	IPPermissions       []*SecurityGroupRule `json:"ipPermissions"`
	IPPermissionsEgress []*SecurityGroupRule `json:"ipPermissionsEgress"`
}

type SecurityGroupRule struct {
	IPProtocol string       `json:"ipProtocol"`
	IPRanges   []*IPRange   `json:"ipRanges"`
	IPv6Ranges []*IPv6Range `json:"ipv6Ranges"`
	FromPort   int64        `json:"fromPort"`
	ToPort     int64        `json:"toPort"`
}

type IPRange struct {
	CidrIP      string `json:"cidrIp"`
	Description string `json:"description"`
}

type IPv6Range struct {
	CidrIPV6    string `json:"cidrIpv6"`
	Description string `json:"description"`
}
type Volume struct {
	RawVolume  *ec2.Volume `json:"-"`
	ID         string      `json:"id"`
	DeviceName string      `json:"deviceName"`
	VolumeType string      `json:"volumeType"`
	Name       string      `json:"name"`
	Encrypted  bool        `json:"encrypted"`
	KMSKeyId   string      `json:"kmsKeyId"`
	Size       int64       `json:"size"`
}
type NatGatewayData struct {
	RawNatGateway *ec2.NatGateway `json:"-"`
	ID            string          `json:"id"`
	PrivateIP     string          `json:"privateIP"`
	PublicIP      string          `json:"publicIP"`
	State         string          `json:"state"`
	Type          string          `json:"type"`
	Name          string          `json:"name"`
}
type NatGateway struct {
	Interfaces map[string]*NetworkInterface `json:"interfaces"`
	NatGatewayData
}

type NatGatewaySorted struct {
	NatGatewayData
	Interfaces []*NetworkInterfaceSorted `json:"interfaces"`
}

type RouteTable struct {
	RawRoute *ec2.RouteTable `json:"-"`
	ID       string          `json:"id"`
	Default  string          `json:"default"`
}

type TGWAttachment struct {
	RawAttachment    *ec2.TransitGatewayVpcAttachment `json:"-"`
	AttachmentID     string                           `json:"attachmentId"`
	TransitGatewayID string                           `json:"transitGatewayId"`
	Name             string                           `json:"name"`
}

type VPCPeer struct {
	RawPeer   *ec2.VpcPeeringConnection `json:"-"`
	ID        string                    `json:"id"`
	Requester string                    `json:"requester"`
	Accepter  string                    `json:"accepter"`
	Name      string                    `json:"name"`
}

type InterfaceEndpointData struct {
	RawEndpoint *ec2.VpcEndpoint `json:"-"`
	ID          string           `json:"id"`
	ServiceName string           `json:"serviceName"`
	Name        string           `json:"name"`
}

type InterfaceEndpoint struct {
	Interfaces map[string]*NetworkInterface
	InterfaceEndpointData
}

type InterfaceEndpointSorted struct {
	InterfaceEndpointData
	Interfaces []*NetworkInterfaceSorted
}

type GatewayEndpoint struct {
	RawEndpoint *ec2.VpcEndpoint `json:"-"`
	ID          string           `json:"id"`
	ServiceName string           `json:"serviceName"`
	Name        string           `json:"name"`
}

type RecievedData struct {
	Error              error
	TransitGateways    []*ec2.TransitGatewayVpcAttachment
	PeeringConnections []*ec2.VpcPeeringConnection
	NetworkInterfaces  []*ec2.NetworkInterface
	SecurityGroups     []*ec2.SecurityGroup
	VPCEndpoints       []*ec2.VpcEndpoint
	wg                 sync.WaitGroup
	mu                 sync.Mutex
}

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
}

func (a *AWSFetch) Make() {
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
