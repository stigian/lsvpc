package main

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type RegionData struct {
	VPCs map[string]VPC
}

type VPC struct {
	Id        *string
	IsDefault bool
	CidrBlock *string
	RawVPC    *ec2.Vpc
	Gateways  []string
	Subnets   map[string]Subnet
}

type Subnet struct {
	Id                 *string
	CidrBlock          *string
	AvailabilityZone   *string
	AvailabilityZoneId *string
	Public             bool
	RawSubnet          *ec2.Subnet
	RouteTable         *RouteTable
	EC2s               map[string]EC2
	NatGateways        map[string]NatGateway
}

type EC2 struct {
	Id         *string
	Type       *string
	SubnetId   *string
	VpcId      *string
	State      *string
	PublicIP   *string
	PrivateIP  *string
	Volumes    map[string]Volume
	Interfaces []NetworkInterface
	RawEc2     *ec2.Instance
}

type NetworkInterface struct {
	Id                  *string
	PrivateIp           *string
	MAC                 *string
	DNS                 *string
	RawNetworkInterface *ec2.InstanceNetworkInterface
}

type Volume struct {
	Id         *string
	DeviceName *string
	Size       *int64
	VolumeType *string
	RawVolume  *ec2.Volume
}
type NatGateway struct {
	Id            *string
	PrivateIP     *string
	PublicIP      *string
	State         *string
	Type          *string
	RawNatGateway *ec2.NatGateway
}

type RouteTable struct {
	Id       *string
	Default  *string
	RawRoute *ec2.RouteTable
}

type InternetGateway struct {
	Id                 *string
	RawInternetGateway *ec2.InternetGateway
}

func getVpcs(svc *ec2.EC2) (map[string]VPC, error) {
	vpclist := []*ec2.Vpc{}
	err := svc.DescribeVpcsPages(
		&ec2.DescribeVpcsInput{},
		func(page *ec2.DescribeVpcsOutput, lastPage bool) bool {
			vpclist = append(vpclist, page.Vpcs...)
			return !lastPage
		},
	)

	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to get vpcs: %v", err.Error())
	}
	vpcs := map[string]VPC{}
	for _, v := range vpclist {
		vpc := VPC{
			Id:        v.VpcId,
			IsDefault: *v.IsDefault,
			CidrBlock: v.CidrBlock,
			RawVPC:    v,
			Subnets:   make(map[string]Subnet),
		}
		vpcs[*v.VpcId] = vpc
	}

	return vpcs, nil

}

func getSubnets(svc *ec2.EC2) ([]*ec2.Subnet, error) {
	subnets := []*ec2.Subnet{}
	err := svc.DescribeSubnetsPages(
		&ec2.DescribeSubnetsInput{},
		func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			subnets = append(subnets, page.Subnets...)
			return !lastPage
		},
	)

	if err != nil {
		return []*ec2.Subnet{}, fmt.Errorf("%v", err.Error())
	}

	return subnets, nil
}

func getInstances(svc *ec2.EC2) ([]*ec2.Reservation, error) {
	instances := []*ec2.Reservation{}
	err := svc.DescribeInstancesPages(
		&ec2.DescribeInstancesInput{},
		func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
			instances = append(instances, page.Reservations...)
			return !lastPage
		},
	)
	if err != nil {
		return []*ec2.Reservation{}, fmt.Errorf("failed: %v", err.Error())
	}

	return instances, nil
}

func getNatGatways(svc *ec2.EC2) ([]*ec2.NatGateway, error) {
	natGateways := []*ec2.NatGateway{}

	err := svc.DescribeNatGatewaysPages(
		&ec2.DescribeNatGatewaysInput{},
		func(page *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			natGateways = append(natGateways, page.NatGateways...)
			return !lastPage
		},
	)
	if err != nil {
		return []*ec2.NatGateway{}, err
	}

	return natGateways, nil
}

func getRouteTables(svc *ec2.EC2) ([]*ec2.RouteTable, error) {
	routeTables := []*ec2.RouteTable{}

	err := svc.DescribeRouteTablesPages(
		&ec2.DescribeRouteTablesInput{},
		func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, page.RouteTables...)
			return !lastPage
		},
	)

	if err != nil {
		return []*ec2.RouteTable{}, err
	}

	return routeTables, nil
}

func getInternetGateways(svc *ec2.EC2) ([]*ec2.InternetGateway, error) {
	internetGateways := []*ec2.InternetGateway{}

	err := svc.DescribeInternetGatewaysPages(
		&ec2.DescribeInternetGatewaysInput{},
		func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, page.InternetGateways...)
			return !lastPage
		},
	)
	if err != nil {
		return []*ec2.InternetGateway{}, err
	}

	return internetGateways, nil
}

func mapSubnets(vpcs map[string]VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		isPublic := aws.BoolValue(v.MapCustomerOwnedIpOnLaunch) || aws.BoolValue(v.MapPublicIpOnLaunch)

		vpcs[*v.VpcId].Subnets[*v.SubnetId] = Subnet{
			Id:                 v.SubnetId,
			CidrBlock:          v.CidrBlock,
			AvailabilityZone:   v.AvailabilityZone,
			AvailabilityZoneId: v.AvailabilityZoneId,
			RawSubnet:          v,
			Public:             isPublic,
			EC2s:               make(map[string]EC2),
			NatGateways:        make(map[string]NatGateway),
		}

	}
}

func mapInstances(vpcs map[string]VPC, reservations []*ec2.Reservation) {
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			networkInterfaces := []NetworkInterface{}
			for _, networkInterface := range instance.NetworkInterfaces {
				networkInterfaces = append(networkInterfaces, NetworkInterface{
					Id:                  networkInterface.NetworkInterfaceId,
					PrivateIp:           networkInterface.PrivateIpAddress,
					MAC:                 networkInterface.MacAddress,
					DNS:                 networkInterface.PrivateDnsName,
					RawNetworkInterface: networkInterface,
				})
			}

			volumes := make(map[string]Volume)
			for _, volume := range instance.BlockDeviceMappings {
				volumes[*volume.Ebs.VolumeId] = Volume{
					Id:         volume.Ebs.VolumeId,
					DeviceName: volume.DeviceName,
				}
			}

			if *instance.State.Name != "terminated" {
				vpcs[*instance.VpcId].Subnets[*instance.SubnetId].EC2s[*instance.InstanceId] = EC2{
					Id:         instance.InstanceId,
					Type:       instance.InstanceType,
					SubnetId:   instance.SubnetId,
					VpcId:      instance.VpcId,
					State:      instance.State.Name,
					PublicIP:   instance.PublicIpAddress,
					PrivateIP:  instance.PrivateIpAddress,
					Volumes:    volumes,
					Interfaces: networkInterfaces,
					RawEc2:     instance,
				}
			}
		}
	}
}

func mapNatGateways(vpcs map[string]VPC, natGateways []*ec2.NatGateway) {
	for _, gateway := range natGateways {
		vpcs[*gateway.VpcId].Subnets[*gateway.SubnetId].NatGateways[*gateway.NatGatewayId] = NatGateway{
			Id:            gateway.NatGatewayId,
			PrivateIP:     gateway.NatGatewayAddresses[0].PrivateIp,
			PublicIP:      gateway.NatGatewayAddresses[0].PublicIp,
			State:         gateway.State,
			Type:          gateway.ConnectivityType,
			RawNatGateway: gateway,
		}
	}
}

func getDefaultRoute(rtb *ec2.RouteTable) string {
	for _, route := range rtb.Routes {
		if aws.StringValue(route.DestinationCidrBlock) == "0.0.0.0/0" ||
			aws.StringValue(route.DestinationIpv6CidrBlock) == "::/0" {

			if dest := aws.StringValue(route.CarrierGatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.EgressOnlyInternetGatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.GatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.InstanceId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.LocalGatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.NatGatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.NetworkInterfaceId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.TransitGatewayId); dest != "" {
				return dest
			}
			if dest := aws.StringValue(route.VpcPeeringConnectionId); dest != "" {
				return dest
			}
		}
	}
	return "" //no default route found, which doesn't necessarily mean an error
}

func mapRouteTables(vpcs map[string]VPC, routeTables []*ec2.RouteTable) {
	// AWS doesn't actually have explicit queryable associations of route
	// tables to subnets. if no other route tables say they are associated
	// with a subnet, then that subnet is assumed to be on the default route table.
	// You can't determine this by looking at the subnets themselves, you
	// have to instead look at all route tables and pick out the ones
	// that say they are associated with particular subnets, and the
	// default route table doesn't even say which subnets they are
	// associated with.

	// first pass, associate the default route with everything
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			if association.Main != nil && *association.Main {
				for subnet_id := range vpcs[*routeTable.VpcId].Subnets {
					subnet := vpcs[*routeTable.VpcId].Subnets[subnet_id]
					defaultRoute := getDefaultRoute(routeTable)
					subnet.RouteTable = &RouteTable{
						Id:       routeTable.RouteTableId,
						Default:  &defaultRoute,
						RawRoute: routeTable,
					}
					vpcs[*routeTable.VpcId].Subnets[subnet_id] = subnet
				}
			}
		}
	}

	// second pass, look at each route table's associations and assign them
	// to their explicitly mentioned subnet
	for _, routeTable := range routeTables {
		for _, association := range routeTable.Associations {
			//default route doesn't have subnet ids and will cause a nil dereference
			if aws.StringValue(association.AssociationState.State) != "associated" ||
				aws.BoolValue(association.Main) {
				continue
			}
			subnet := vpcs[*routeTable.VpcId].Subnets[*association.SubnetId]
			defaultRoute := getDefaultRoute(routeTable)
			subnet.RouteTable = &RouteTable{
				Id:       routeTable.RouteTableId,
				Default:  &defaultRoute,
				RawRoute: routeTable,
			}
			vpcs[*routeTable.VpcId].Subnets[*association.SubnetId] = subnet
		}
	}
}

func mapInternetGateways(vpcs map[string]VPC, internetGateways []*ec2.InternetGateway) {
	for _, igw := range internetGateways {
		fmt.Printf("%#v\n", igw)
		for _, attachment := range igw.Attachments {
			if vpcId := aws.StringValue(attachment.VpcId); vpcId != "" {
				vpc := vpcs[vpcId]
				vpc.Gateways = append(vpc.Gateways, aws.StringValue(igw.InternetGatewayId))
				vpcs[vpcId] = vpc
			}
		}
	}
}

func getVolume(svc *ec2.EC2, volumeId string) (*ec2.Volume, error) {
	if volumeId == "" {
		return &ec2.Volume{}, fmt.Errorf("getVolume handed an empty string")
	}
	out, err := svc.DescribeVolumes(&ec2.DescribeVolumesInput{
		VolumeIds: []*string{
			aws.String(volumeId),
		},
	})
	if err != nil {
		return &ec2.Volume{}, err
	}

	if len(out.Volumes) != 1 {
		return &ec2.Volume{}, fmt.Errorf("incorrect number of volumes returned")
	}

	return out.Volumes[0], nil

}

func instantiateVolumes(svc *ec2.EC2, vpcs map[string]VPC) error {
	for vk, v := range vpcs {
		for sk, s := range v.Subnets {
			for ik, i := range s.EC2s {
				for volk, vol := range i.Volumes {
					volume, err := getVolume(svc, aws.StringValue(vol.Id))
					if err != nil {
						return err
					}
					vpcs[vk].Subnets[sk].EC2s[ik].Volumes[volk] = Volume{
						Id:         vol.Id,
						DeviceName: vol.DeviceName,
						Size:       volume.Size,
						VolumeType: volume.VolumeType,
						RawVolume:  volume,
					}
				}
			}
		}
	}
	return nil
}

func indent(num int) string {
	sb := strings.Builder{}

	for i := 0; i <= num; i++ {
		sb.WriteString(" ")
	}

	return sb.String()
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

func printVPCs(vpcs map[string]VPC) {
	for _, vpc := range vpcs {

		// Print VPC
		if vpc.IsDefault {
			fmt.Printf(
				"%v(default)%v ",
				string(colorYellow),
				string(colorReset),
			)
		}
		fmt.Printf(
			"%v%v%v %v: ",
			string(colorGreen),
			aws.StringValue(vpc.Id),
			string(colorReset),
			aws.StringValue(vpc.CidrBlock),
		)
		// Print gateways set to VPC
		for _, gateway := range vpc.Gateways {
			fmt.Printf(
				"%v%v%v ",
				string(colorYellow),
				gateway,
				string(colorReset),
			)
		}

		fmt.Printf("\n")

		// Print Subnets
		for _, subnet := range vpc.Subnets {

			// Print Subnet Info
			public := "Private"
			if subnet.Public {
				public = "Public"
			}
			fmt.Printf(
				"%s%v%v%v  %v  %v %v-->%v%v %v\n",
				indent(4),
				string(colorBlue),
				aws.StringValue(subnet.Id),
				string(colorReset),
				aws.StringValue(subnet.AvailabilityZone),
				aws.StringValue(subnet.CidrBlock),
				string(colorYellow),
				aws.StringValue(subnet.RouteTable.Default),
				string(colorReset),
				public,
			)

			// Print EC2 Instance
			for _, instance := range subnet.EC2s {

				// Print Instance Info
				fmt.Printf(
					"%s%v%s%v -- %v -- %v -- %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(instance.Id),
					string(colorReset),
					aws.StringValue(instance.State),
					aws.StringValue(instance.PublicIP),
					aws.StringValue(instance.PrivateIP),
				)

				// Print Instance Volumes
				for _, iface := range instance.Interfaces {
					fmt.Printf(
						"%s%v  %v  %v  %v\n",
						indent(12),
						aws.StringValue(iface.Id),
						aws.StringValue(iface.MAC),
						aws.StringValue(iface.PrivateIp),
						aws.StringValue(iface.DNS),
					)
				}

				// Print Instance Volumes
				for _, volume := range instance.Volumes {
					fmt.Printf(
						"%s%v  %v  %v  %v GiB\n",
						indent(12),
						aws.StringValue(volume.Id),
						aws.StringValue(volume.VolumeType),
						aws.StringValue(volume.DeviceName),
						aws.Int64Value(volume.Size),
					)
				}
			}

			//Print Nat Gateways
			for _, natGateway := range subnet.NatGateways {
				fmt.Printf(
					"%s%v%v%v  %v  %v  %v  %v\n",
					indent(8),
					string(colorCyan),
					aws.StringValue(natGateway.Id),
					string(colorReset),
					aws.StringValue(natGateway.Type),
					aws.StringValue(natGateway.State),
					aws.StringValue(natGateway.PublicIP),
					aws.StringValue(natGateway.PrivateIP),
				)
			}

			fmt.Printf("\n")
		}
	}
}

func populateVPC(region string) (map[string]VPC, error) {
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String(region),
			},
		},
	))

	svc := ec2.New(sess)

	vpcs, err := getVpcs(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}

	subnets, err := getSubnets(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapSubnets(vpcs, subnets)
	instances, err := getInstances(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapInstances(vpcs, instances)
	err = instantiateVolumes(svc, vpcs)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	natGateways, err := getNatGatways(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapNatGateways(vpcs, natGateways)
	routeTables, err := getRouteTables(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapRouteTables(vpcs, routeTables)

	internetGateways, err := getInternetGateways(svc)
	if err != nil {
		return map[string]VPC{}, fmt.Errorf("failed to populate VPCs: %v", err.Error())
	}
	mapInternetGateways(vpcs, internetGateways)

	return vpcs, nil
}

func getRegions() []string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)
	regions := []string{}
	res, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		panic("Could not get regions")
	}

	for _, region := range res.Regions {
		regions = append(regions, aws.StringValue(region.RegionName))
	}

	return regions
}

func getRegionData(fullData map[string]RegionData, region string, wg *sync.WaitGroup) {
	defer wg.Done()
	vpcs, err := populateVPC(region)
	if err != nil {
		return
	}
	fullData[region] = RegionData{
		VPCs: vpcs,
	}
}

func allRegions() {
	var wg sync.WaitGroup

	regions := getRegions()

	fullData := make(map[string]RegionData)

	for _, region := range regions {
		wg.Add(1)
		go getRegionData(fullData, region, &wg)
	}

	wg.Wait()

	for region, vpcs := range fullData {
		fmt.Printf("===%v===\n", region)
		printVPCs(vpcs.VPCs)
	}

}
func defaultRegion() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	currentRegion := aws.StringValue(sess.Config.Region)
	vpcs, err := populateVPC(currentRegion)
	if err != nil {
		panic("populateVPC failed")
	}

	printVPCs(vpcs)

}
func main() {
	//allRegions()
	defaultRegion()
}
