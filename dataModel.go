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
	VPCs   []VPCSorted
}

type VPCData struct {
	ID            string `json:"id"`
	IsDefault     bool   `json:"isDefault"`
	CidrBlock     string `json:"cidrBlock"`
	IPv6CidrBlock string `json:"iPv6CidrBlock"`
	Name          string `json:"name"`
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
	Gateways []string
	Subnets  map[string]*Subnet
	Peers    map[string]*VPCPeer
}

type SubnetData struct {
	ID                 string `json:"id"`
	CidrBlock          string `json:"cidrBlock"`
	AvailabilityZone   string `json:"availabilityZone"`
	AvailabilityZoneID string `json:"availabilityZoneId"`
	Public             bool   `json:"public"`
	Name               string `json:"name"`
	RouteTable         *RouteTable
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
	SubnetData
	RawSubnet          *ec2.Subnet
	Instances          map[string]*Instance
	NatGateways        map[string]*NatGateway
	TGWs               map[string]*TGWAttachment
	ENIs               map[string]*NetworkInterface
	InterfaceEndpoints map[string]*InterfaceEndpoint
	GatewayEndpoints   map[string]*GatewayEndpoint
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
	Volumes    []Volume           `json:"volumes,omitempty"`
	Interfaces []NetworkInterface `json:"interfaces,omitempty"`
}

type Instance struct {
	InstanceData
	RawEc2     *ec2.Instance
	Volumes    map[string]*Volume
	Interfaces map[string]*NetworkInterface
}

type NetworkInterface struct {
	ID                  string                `json:"id"`
	PrivateIP           string                `json:"privateIp"`
	MAC                 string                `json:"mAC"`
	DNS                 string                `json:"dNS"`
	Type                string                `json:"type"`
	Description         string                `json:"description"`
	PublicIP            string                `json:"publicIp"`
	Name                string                `json:"name"`
	SubnetID            string                `json:"subnetId"` // we're just accounting for this for display purposes
	RawNetworkInterface *ec2.NetworkInterface `json:"-"`
}

type Volume struct {
	ID         string      `json:"id"`
	DeviceName string      `json:"deviceName"`
	Size       int64       `json:"size"`
	VolumeType string      `json:"volumeType"`
	Name       string      `json:"name"`
	RawVolume  *ec2.Volume `json:"-"`
}
type NatGateway struct {
	ID            string          `json:"id"`
	PrivateIP     string          `json:"privateIP"`
	PublicIP      string          `json:"publicIP"`
	State         string          `json:"state"`
	Type          string          `json:"type"`
	Name          string          `json:"name"`
	RawNatGateway *ec2.NatGateway `json:"-"`
}

type RouteTable struct {
	ID       string          `json:"id"`
	Default  string          `json:"default"`
	RawRoute *ec2.RouteTable `json:"-"`
}

type TGWAttachment struct {
	AttachmentID     string                           `json:"attachmentId"`
	TransitGatewayID string                           `json:"transitGatewayId"`
	Name             string                           `json:"name"`
	RawAttachment    *ec2.TransitGatewayVpcAttachment `json:"-"`
}

type VPCPeer struct {
	ID        string                    `json:"id"`
	Requester string                    `json:"requester"`
	Accepter  string                    `json:"accepter"`
	Name      string                    `json:"name"`
	RawPeer   *ec2.VpcPeeringConnection `json:"-"`
}

type InterfaceEndpointData struct {
	ID          string           `json:"id"`
	ServiceName string           `json:"serviceName"`
	Name        string           `json:"name"`
	RawEndpoint *ec2.VpcEndpoint `json:"-"`
}

type InterfaceEndpoint struct {
	InterfaceEndpointData
	Interfaces map[string]*NetworkInterface
}

type InterfaceEndpointSorted struct {
	InterfaceEndpointData
	Interfaces []*NetworkInterface
}

type GatewayEndpoint struct {
	ID          string           `json:"id"`
	ServiceName string           `json:"serviceName"`
	Name        string           `json:"name"`
	RawEndpoint *ec2.VpcEndpoint `json:"-"`
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
