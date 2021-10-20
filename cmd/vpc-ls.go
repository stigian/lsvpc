package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type VPC struct {
	Id        string
	IsDefault bool
	CidrBlock string
	RawVPC    *ec2.Vpc
	Subnets   map[string]Subnet
}

type Subnet struct {
	Id                 string
	CidrBlock          string
	AvailabilityZone   string
	AvailabilityZoneId string
	RawSubnet          *ec2.Subnet
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
			Id:        *v.VpcId,
			IsDefault: *v.IsDefault,
			CidrBlock: *v.CidrBlock,
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

func mapSubnets(vpcs map[string]VPC, subnets []*ec2.Subnet) {
	for _, v := range subnets {
		vpcs[*v.VpcId].Subnets[*v.SubnetId] = Subnet{
			Id:                 *v.SubnetId,
			CidrBlock:          *v.CidrBlock,
			AvailabilityZone:   *v.AvailabilityZone,
			AvailabilityZoneId: *v.AvailabilityZoneId,
			RawSubnet:          v,
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
	for _, v := range vpcs {
		if v.IsDefault {
			fmt.Printf("(default) ")
		}
		fmt.Printf("%v --- %v\n", v.Id, v.CidrBlock)
		for _, w := range v.Subnets {
			fmt.Printf("    ")
			fmt.Printf("%v %v --- %v\n", w.Id, w.AvailabilityZone, w.CidrBlock)
		}
	}
}
