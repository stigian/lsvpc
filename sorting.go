// Copyright 2023 Stigian Consulting - reference license in top level of project
package main

import (
	"sort"
)

func asyncSortRegion(vpcs map[string]*VPC, out chan []*VPCSorted) {
	out <- sortVPCs(vpcs)
}

func sortRegionData(regionData map[string]RegionData) []*RegionDataSorted {
	regionKeys := []string{}

	for k := range regionData {
		regionKeys = append(regionKeys, k)
	}

	sort.Strings(regionKeys)

	regionDataInterstitial := make(map[string]chan []*VPCSorted)

	// Dispatch parallel sorting
	for _, region := range regionKeys {
		regionDataInterstitial[region] = make(chan []*VPCSorted)

		go asyncSortRegion(regionData[region].VPCs, regionDataInterstitial[region])
	}

	// Receive sorting results from goroutines
	regionDataIn := make(map[string]*RegionDataSorted)
	for _, region := range regionKeys {
		regionDataIn[region] = &RegionDataSorted{
			VPCs:   <-regionDataInterstitial[region],
			Region: region,
		}
	}

	// Append to regionDataSorted
	regionDataSorted := []*RegionDataSorted{}
	for _, region := range regionKeys {
		regionDataSorted = append(regionDataSorted, regionDataIn[region])
	}

	return regionDataSorted
}

func sortVPCs(vpcs map[string]*VPC) []*VPCSorted {
	vpcKeys := []string{}
	for k := range vpcs {
		vpcKeys = append(vpcKeys, k)
	}

	sort.Strings(vpcKeys)

	vpcsSorted := []*VPCSorted{}
	for _, vpcID := range vpcKeys {
		vpcsSorted = append(vpcsSorted, sortVPC(vpcs[vpcID]))
	}

	return vpcsSorted
}

func sortVPC(vpc *VPC) *VPCSorted {
	// Sort Gateways
	gatewaysSorted := vpc.Gateways
	sort.Strings(gatewaysSorted)

	// Sort subnets
	subnetKeys := []string{}
	for k := range vpc.Subnets {
		subnetKeys = append(subnetKeys, k)
	}

	sort.Strings(subnetKeys)

	subnetsSorted := []*SubnetSorted{}
	for _, subnetID := range subnetKeys {
		subnetsSorted = append(subnetsSorted, sortSubnet(vpc.Subnets[subnetID]))
	}

	// Sort peers
	peerKeys := []string{}
	for k := range vpc.Peers {
		peerKeys = append(peerKeys, k)
	}

	sort.Strings(peerKeys)

	peersSorted := []*VPCPeer{}
	for _, peerID := range peerKeys {
		peersSorted = append(peersSorted, vpc.Peers[peerID])
	}

	return &VPCSorted{
		VPCData:  vpc.VPCData,
		Gateways: gatewaysSorted,
		Subnets:  subnetsSorted,
		Peers:    peersSorted,
	}
}

func sortSubnet(subnet *Subnet) *SubnetSorted {
	// Sort Instances
	instanceKeys := []string{}
	for k := range subnet.Instances {
		instanceKeys = append(instanceKeys, k)
	}

	sort.Strings(instanceKeys)

	instancesSorted := []*InstanceSorted{}
	for _, instanceID := range instanceKeys {
		instancesSorted = append(instancesSorted, sortInstance(subnet.Instances[instanceID]))
	}

	// Sort NatGateways
	natGatewayKeys := []string{}
	for k := range subnet.NatGateways {
		natGatewayKeys = append(natGatewayKeys, k)
	}

	sort.Strings(natGatewayKeys)

	natGatewaysSorted := []*NatGateway{}
	for _, natGatewayID := range natGatewayKeys {
		natGatewaysSorted = append(natGatewaysSorted, subnet.NatGateways[natGatewayID])
	}

	// Sort TGWAttachments
	transitGatewayKeys := []string{}
	for k := range subnet.TGWs {
		transitGatewayKeys = append(transitGatewayKeys, k)
	}

	sort.Strings(transitGatewayKeys)

	transitGatewaysSorted := []*TGWAttachment{}
	for _, transitGatewayID := range transitGatewayKeys {
		transitGatewaysSorted = append(transitGatewaysSorted, subnet.TGWs[transitGatewayID])
	}

	// Sort ENIs
	networkInterfaceKeys := []string{}
	for k := range subnet.ENIs {
		networkInterfaceKeys = append(networkInterfaceKeys, k)
	}

	sort.Strings(networkInterfaceKeys)

	networkInterfacesSorted := []*NetworkInterface{}
	for _, networkInterfaceID := range networkInterfaceKeys {
		networkInterfacesSorted = append(networkInterfacesSorted, subnet.ENIs[networkInterfaceID])
	}

	// Sort InterfaceEndpoints
	interfaceEndpointKeys := []string{}
	for k := range subnet.InterfaceEndpoints {
		interfaceEndpointKeys = append(interfaceEndpointKeys, k)
	}

	sort.Strings(interfaceEndpointKeys)

	interfaceEndpointsSorted := []*InterfaceEndpointSorted{}
	for _, interfaceEndpointID := range interfaceEndpointKeys {
		interfaceEndpointsSorted = append(interfaceEndpointsSorted, sortInterfaceEndpoint(subnet.InterfaceEndpoints[interfaceEndpointID]))
	}

	// Sort GatewayEndpoints
	gatewayEndpointKeys := []string{}
	for k := range subnet.GatewayEndpoints {
		gatewayEndpointKeys = append(gatewayEndpointKeys, k)
	}

	sort.Strings(gatewayEndpointKeys)

	gatewayEndpointsSorted := []*GatewayEndpoint{}
	for _, gatewayEndpointID := range gatewayEndpointKeys {
		gatewayEndpointsSorted = append(gatewayEndpointsSorted, subnet.GatewayEndpoints[gatewayEndpointID])
	}

	return &SubnetSorted{
		SubnetData:         subnet.SubnetData,
		Instances:          instancesSorted,
		NatGateways:        natGatewaysSorted,
		TGWs:               transitGatewaysSorted,
		ENIs:               networkInterfacesSorted,
		InterfaceEndpoints: interfaceEndpointsSorted,
		GatewayEndpoints:   gatewayEndpointsSorted,
	}
}

func sortInstance(instance *Instance) *InstanceSorted {
	// Sort volumes
	volumeKeys := []string{}
	for k := range instance.Volumes {
		volumeKeys = append(volumeKeys, k)
	}

	sort.Strings(volumeKeys)

	volumesSorted := []*Volume{}
	for _, volumeID := range volumeKeys {
		volumesSorted = append(volumesSorted, instance.Volumes[volumeID])
	}

	// Sort network interfaces
	interfaceKeys := []string{}
	for k := range instance.Interfaces {
		interfaceKeys = append(interfaceKeys, k)
	}

	sort.Strings(interfaceKeys)

	interfacesSorted := []*NetworkInterfaceSorted{}
	for _, interfaceID := range interfaceKeys {
		interfacesSorted = append(interfacesSorted, sortNetworkInterface(instance.Interfaces[interfaceID]))
	}

	return &InstanceSorted{
		InstanceData: instance.InstanceData,
		Volumes:      volumesSorted,
		Interfaces:   interfacesSorted,
	}
}

func sortInterfaceEndpoint(endpoint *InterfaceEndpoint) *InterfaceEndpointSorted {
	ifaceKeys := []string{}
	for k := range endpoint.Interfaces {
		ifaceKeys = append(ifaceKeys, k)
	}

	sort.Strings(ifaceKeys)

	interfacesSorted := []*NetworkInterfaceSorted{}
	for _, interfaceID := range ifaceKeys {
		interfacesSorted = append(interfacesSorted, sortNetworkInterface(endpoint.Interfaces[interfaceID]))
	}

	return &InterfaceEndpointSorted{
		InterfaceEndpointData: endpoint.InterfaceEndpointData,
		Interfaces:            interfacesSorted,
	}
}

func sortNetworkInterface(iface *NetworkInterface) *NetworkInterfaceSorted {
	groupKeys := []string{}
	for k := range iface.Groups {
		groupKeys = append(groupKeys, k)
	}

	sort.Strings(groupKeys)

	groupsSorted := []*SecurityGroup{}
	for _, groupID := range groupKeys {
		groupsSorted = append(groupsSorted, iface.Groups[groupID])
	}

	return &NetworkInterfaceSorted{
		NetworkInterfaceData: iface.NetworkInterfaceData,
		Groups:               groupsSorted,
	}
}

func sortNatGateway(ng *NatGateway) *NatGatewaySorted {
	ifaceKeys := []string{}
	for k := range ng.Interfaces {
		ifaceKeys = append(ifaceKeys, k)
	}

	sort.Strings(ifaceKeys)

	ifaceSorted := []*NetworkInterfaceSorted{}
	for _, ifaceID := range ifaceKeys {
		ifaceSorted = append(ifaceSorted, sortNetworkInterface(ng.Interfaces[ifaceID]))
	}

	return &NatGatewaySorted{
		NatGatewayData: ng.NatGatewayData,
		Interfaces:     ifaceSorted,
	}
}
