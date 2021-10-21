package main

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type VPC struct {
	Id        *string
	IsDefault bool
	CidrBlock *string
	RawVPC    *ec2.Vpc
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
	Id                  string
	PrivateIp           string
	MAC                 string
	DNS                 string
	RawNetworkInterface *ec2.InstanceNetworkInterface
}

type Volume struct {
	Id         string
	DeviceName string
	Size       int64
	VolumeType string
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

func mapSubnets(vpcs map[string]VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		isPublic := false
		if v.MapCustomerOwnedIpOnLaunch != nil {
			isPublic = *v.MapCustomerOwnedIpOnLaunch || *v.MapPublicIpOnLaunch
		} else {
			isPublic = *v.MapPublicIpOnLaunch
		}

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
					Id:                  *networkInterface.NetworkInterfaceId,
					PrivateIp:           *networkInterface.PrivateIpAddress,
					MAC:                 *networkInterface.MacAddress,
					DNS:                 *networkInterface.PrivateDnsName,
					RawNetworkInterface: networkInterface,
				})
			}

			volumes := make(map[string]Volume)
			for _, volume := range instance.BlockDeviceMappings {
				volumes[*volume.Ebs.VolumeId] = Volume{
					Id:         *volume.Ebs.VolumeId,
					DeviceName: *volume.DeviceName,
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

func getVolume(svc *ec2.EC2, volumeId string) (*ec2.Volume, error) {
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
					volume, err := getVolume(svc, vol.Id)
					if err != nil {
						return err
					}
					vpcs[vk].Subnets[sk].EC2s[ik].Volumes[volk] = Volume{
						Id:         vol.Id,
						DeviceName: vol.DeviceName,
						Size:       *volume.Size,
						VolumeType: *volume.VolumeType,
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
			"%v%v%v --- %v\n",
			string(colorGreen),
			aws.StringValue(vpc.Id),
			string(colorReset),
			aws.StringValue(vpc.CidrBlock),
		)

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

				// Print Instance Interfaces
				for _, volume := range instance.Interfaces {
					fmt.Printf("%s%v  %v  %v  %v\n", indent(12), volume.Id, volume.MAC, volume.PrivateIp, volume.DNS)
				}

				// Print Instance Volumes
				for _, volume := range instance.Volumes {
					fmt.Printf("%s%v  %v  %v  %v GiB\n", indent(12), volume.Id, volume.VolumeType, volume.DeviceName, volume.Size)
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

func main() {
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
		},
	))

	svc := ec2.New(sess)

	vpcs, err := getVpcs(svc)
	if err != nil {
		fmt.Printf("failed to get vpcs: %v", err.Error())
	}

	subnets, _ := getSubnets(svc)
	mapSubnets(vpcs, subnets)
	instances, _ := getInstances(svc)
	mapInstances(vpcs, instances)
	err = instantiateVolumes(svc, vpcs)
	if err != nil {
		fmt.Printf("failed to instantiate volumes")
	}
	natGateways, _ := getNatGatways(svc)
	mapNatGateways(vpcs, natGateways)
	routeTables, _ := getRouteTables(svc)
	mapRouteTables(vpcs, routeTables)

	printVPCs(vpcs)
}
