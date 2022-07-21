package main

import (
	"sort"
)

func sortVPCs(vpcs map[string]VPC) []VPCSorted {

	vpcKeys := []string{}
	for k := range vpcs {
		vpcKeys = append(vpcKeys, k)
	}

	sort.Strings(vpcKeys)

	vpcsSorted := []VPCSorted{}
	for _, vpcId := range vpcKeys {
		vpcsSorted = append(vpcsSorted, sortVPC(vpcs[vpcId]))
	}
	return vpcsSorted
}

func sortVPC(vpc VPC) VPCSorted {

	//sort Gateways
	gatewaysSorted := vpc.Gateways
	sort.Strings(gatewaysSorted)

	//sort subnets
	subnetKeys := []string{}
	for k := range vpc.Subnets {
		subnetKeys = append(subnetKeys, k)
	}
	sort.Strings(subnetKeys)

	subnetsSorted := []Subnet{}
	for _, subnetId := range subnetKeys {
		subnetsSorted = append(subnetsSorted, vpc.Subnets[subnetId])
	}

	//sort peers
	peerKeys := []string{}
	for k := range vpc.Peers {
		peerKeys = append(peerKeys, k)
	}
	sort.Strings(peerKeys)
	peersSorted := []VPCPeer{}
	for _, peerId := range peerKeys {
		peersSorted = append(peersSorted, vpc.Peers[peerId])
	}

	return VPCSorted{
		VPCData:  vpc.VPCData,
		Gateways: gatewaysSorted,
		Subnets:  subnetsSorted,
		Peers:    peersSorted,
	}
}
