package main

import (
	"sort"
)

func sortRegionData(regionData map[string]RegionData) []RegionDataSorted {
	regionKeys := []string{}
	for k := range regionData {
		regionKeys = append(regionKeys, k)
	}
	sort.Strings(regionKeys)

	regionDataSorted := []RegionDataSorted{}
	for _, region := range regionKeys {
		regionDataSorted = append(regionDataSorted, RegionDataSorted{
			Region: region,
			VPCs:   sortVPCs(regionData[region].VPCs),
		})
	}

	return regionDataSorted
}

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

	subnetsSorted := []SubnetSorted{}
	for _, subnetId := range subnetKeys {
		subnetsSorted = append(subnetsSorted, sortSubnet(vpc.Subnets[subnetId]))
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

func sortSubnet(subnet Subnet) SubnetSorted {
	//sort Instances
	instanceKeys := []string{}
	for k := range subnet.Instances {
		instanceKeys = append(instanceKeys, k)
	}
	sort.Strings(instanceKeys)
	instancesSorted := []InstanceSorted{}
	for _, instanceId := range instanceKeys {
		instancesSorted = append(instancesSorted, sortInstance(subnet.Instances[instanceId]))
	}

	//sort NatGateways
	natGatewayKeys := []string{}
	for k := range subnet.NatGateways {
		natGatewayKeys = append(natGatewayKeys, k)
	}
	sort.Strings(natGatewayKeys)
	natGatewaysSorted := []NatGateway{}
	for _, natGatewayId := range natGatewayKeys {
		natGatewaysSorted = append(natGatewaysSorted, subnet.NatGateways[natGatewayId])
	}

	//sort TGWAttachments
	transitGatewayKeys := []string{}
	for k := range subnet.TGWs {
		transitGatewayKeys = append(transitGatewayKeys, k)
	}
	sort.Strings(transitGatewayKeys)
	transitGatewaysSorted := []TGWAttachment{}
	for _, transitGatewayId := range transitGatewayKeys {
		transitGatewaysSorted = append(transitGatewaysSorted, subnet.TGWs[transitGatewayId])
	}

	//sort ENIs
	networkInterfaceKeys := []string{}
	for k := range subnet.ENIs {
		networkInterfaceKeys = append(networkInterfaceKeys, k)
	}
	sort.Strings(networkInterfaceKeys)
	networkInterfacesSorted := []NetworkInterface{}
	for _, networkInterfaceId := range networkInterfaceKeys {
		networkInterfacesSorted = append(networkInterfacesSorted, subnet.ENIs[networkInterfaceId])
	}

	//sort InterfaceEndpoints
	interfaceEndpointKeys := []string{}
	for k := range subnet.InterfaceEndpoints {
		interfaceEndpointKeys = append(interfaceEndpointKeys, k)
	}
	sort.Strings(interfaceEndpointKeys)
	interfaceEndpointsSorted := []InterfaceEndpointSorted{}
	for _, interfaceEndpointID := range interfaceEndpointKeys {
		interfaceEndpointsSorted = append(interfaceEndpointsSorted, sortInterfaceEndpoint(subnet.InterfaceEndpoints[interfaceEndpointID]))
	}

	//sort GatewayEndpoints
	gatewayEndpointKeys := []string{}
	for k := range subnet.GatewayEndpoints {
		gatewayEndpointKeys = append(gatewayEndpointKeys, k)
	}
	sort.Strings(gatewayEndpointKeys)
	gatewayEndpointsSorted := []GatewayEndpoint{}
	for _, gatewayEndpointId := range gatewayEndpointKeys {
		gatewayEndpointsSorted = append(gatewayEndpointsSorted, subnet.GatewayEndpoints[gatewayEndpointId])
	}

	return SubnetSorted{
		SubnetData:         subnet.SubnetData,
		Instances:          instancesSorted,
		NatGateways:        natGatewaysSorted,
		TGWs:               transitGatewaysSorted,
		ENIs:               networkInterfacesSorted,
		InterfaceEndpoints: interfaceEndpointsSorted,
		GatewayEndpoints:   gatewayEndpointsSorted,
	}
}

func sortInstance(instance Instance) InstanceSorted {

	// sort volumes
	volumeKeys := []string{}
	for k := range instance.Volumes {
		volumeKeys = append(volumeKeys, k)
	}
	sort.Strings(volumeKeys)
	volumesSorted := []Volume{}
	for _, volumeId := range volumeKeys {
		volumesSorted = append(volumesSorted, instance.Volumes[volumeId])
	}

	// sort network interfaces
	interfaceKeys := []string{}
	for k := range instance.Interfaces {
		interfaceKeys = append(interfaceKeys, k)
	}
	sort.Strings(interfaceKeys)
	interfacesSorted := []NetworkInterface{}
	for _, interfaceId := range interfaceKeys {
		interfacesSorted = append(interfacesSorted, instance.Interfaces[interfaceId])
	}

	return InstanceSorted{
		InstanceData: instance.InstanceData,
		Volumes:      volumesSorted,
		Interfaces:   interfacesSorted,
	}

}

func sortInterfaceEndpoint(endpoint InterfaceEndpoint) InterfaceEndpointSorted {
	ifaceKeys := []string{}
	for k := range endpoint.Interfaces {
		ifaceKeys = append(ifaceKeys, k)
	}
	sort.Strings(ifaceKeys)

	interfacesSorted := []NetworkInterface{}
	for _, interfaceId := range ifaceKeys {
		interfacesSorted = append(interfacesSorted, endpoint.Interfaces[interfaceId])
	}
	return InterfaceEndpointSorted{
		InterfaceEndpointData: endpoint.InterfaceEndpointData,
		Interfaces:            interfacesSorted,
	}
}
