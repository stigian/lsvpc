// Copyright 2021 Stigian Consulting - reference license in top level of project

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/sts"
)

type lsvpcConfig struct {
	regionOverride string
	noColor        bool
	noSpace        bool
	allRegions     bool
	Color          bool
	jsonOutput     bool
	Verbose        bool
	HideIP         bool
	Truncate       bool
}

var Config lsvpcConfig

func populateVPC(region string) (map[string]*VPC, error) {
	sess := session.Must(session.NewSessionWithOptions(
		session.Options{
			SharedConfigState: session.SharedConfigEnable,
			Config: aws.Config{
				Region: aws.String(region),
			},
		},
	))

	svc := ec2.New(sess)
	stsSvc := sts.New(sess)
	data := RecievedData{}
	vpcs := make(map[string]*VPC)

	data.wg.Add(5) //nolint:gomnd // Wait groups increase when requests increase

	fetch := AWSFetch{}
	fetch.Make()

	go getIdentity(stsSvc, fetch.Identity)
	go getVpcs(svc, fetch.Vpcs)
	go getSubnets(svc, fetch.Subnets)
	go getInstances(svc, fetch.Instances)
	go getInstanceStatuses(svc, fetch.InstanceStatuses)
	go getVolumes(svc, fetch.Volumes)
	go getNatGatways(svc, fetch.NatGateways)
	go getRouteTables(svc, fetch.RouteTables)
	go getInternetGateways(svc, fetch.InternetGateways)
	go getEgressOnlyInternetGateways(svc, fetch.EOInternetGateways)
	go getVPNGateways(svc, fetch.VPNGateways)
	go getTransitGatewayVpcAttachments(svc, &data)
	go getVpcPeeringConnections(svc, &data)
	go getNetworkInterfaces(svc, &data)
	go getSecurityGroups(svc, &data)
	go getVpcEndpoints(svc, &data)

	data.wg.Wait()

	// This isn't exhaustive error reporting, but all we really care about is if anything failed at all
	if data.Error != nil {
		return map[string]*VPC{}, fmt.Errorf("failed to populate VPCs: %v", data.Error.Error())
	}
	identityOut := <-fetch.Identity
	if identityOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch identity information: %v", data.Error.Error())
	}
	vpcsOut := <-fetch.Vpcs
	if vpcsOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch vpc information: %v", data.Error.Error())
	}
	subnetsOut := <-fetch.Subnets
	if subnetsOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch subnet information: %v", data.Error.Error())
	}
	instancesOut := <-fetch.Instances
	if instancesOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch instance information: %v", data.Error.Error())
	}
	instanceStatusOut := <-fetch.InstanceStatuses
	if instanceStatusOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch instance status information: %v", data.Error.Error())
	}
	volumesOut := <-fetch.Volumes
	if volumesOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch volume information: %v", data.Error.Error())
	}
	natGatewaysOut := <-fetch.NatGateways
	if natGatewaysOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch NatGateway information: %v", data.Error.Error())
	}
	routeTablesOut := <-fetch.RouteTables
	if routeTablesOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch route table information: %v", data.Error.Error())
	}
	internetGatewaysOut := <-fetch.InternetGateways
	if internetGatewaysOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch internet gateway information: %v", data.Error.Error())
	}
	eointernetGatewaysOut := <-fetch.EOInternetGateways
	if eointernetGatewaysOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch egress-only internet gateway information: %v", data.Error.Error())
	}
	vpnGatewaysOut := <-fetch.VPNGateways
	if vpnGatewaysOut.Err != nil {
		return map[string]*VPC{}, fmt.Errorf("Failed to fetch vpn gateway information: %v", data.Error.Error())
	}

	mapVpcs(vpcs, vpcsOut.Vpcs)
	mapSubnets(vpcs, subnetsOut.Subnets)
	mapInstances(vpcs, instancesOut.Instances)
	mapInstanceStatuses(vpcs, instanceStatusOut.InstanceStatuses)
	mapVolumes(vpcs, volumesOut.Volumes)
	mapNatGateways(vpcs, natGatewaysOut.NatGateways)
	mapRouteTables(vpcs, routeTablesOut.RouteTables)
	mapInternetGateways(vpcs, internetGatewaysOut.InternetGateways)
	mapEgressOnlyInternetGateways(vpcs, eointernetGatewaysOut.EOInternetGateways)
	mapVPNGateways(vpcs, vpnGatewaysOut.VPNGateways)
	mapTransitGatewayVpcAttachments(vpcs, data.TransitGateways, identityOut.Identity)
	mapVpcPeeringConnections(vpcs, data.PeeringConnections)
	mapVpcEndpoints(vpcs, data.VPCEndpoints)
	mapNetworkInterfaces(vpcs, data.NetworkInterfaces)
	mapSecurityGroups(vpcs, data.SecurityGroups)

	return vpcs, nil
}

func getRegionData(fullData map[string]RegionData, region string, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()

	vpcs, err := populateVPC(region)
	if err != nil {
		return
	}

	mu.Lock()

	fullData[region] = RegionData{
		VPCs: vpcs,
	}

	mu.Unlock()
}

func doSpecificRegion() {
	region := Config.regionOverride

	vpcs, err := populateVPC(region)
	if err != nil {
		return
	}

	if Config.jsonOutput {
		printVPCsJSON(sortVPCs(vpcs))
	} else {
		printVPCs(sortVPCs(vpcs))
	}
}

func doAllRegions() {
	var wg sync.WaitGroup

	regions := getRegions()

	fullData := make(map[string]RegionData)
	mu := sync.Mutex{}

	for _, region := range regions {
		wg.Add(1)

		go getRegionData(fullData, region, &wg, &mu)
	}

	wg.Wait()

	regionDataSorted := sortRegionData(fullData)

	if Config.jsonOutput {
		printRegionsJSON(regionDataSorted)
	} else {
		for _, region := range regionDataSorted {
			fmt.Printf("===%v===\n", region.Region)
			printVPCs(region.VPCs)
		}
	}
}

func doDefaultRegion() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	currentRegion := aws.StringValue(sess.Config.Region)

	vpcs, err := populateVPC(currentRegion)
	if err != nil {
		panic(fmt.Sprintf("populateVPC failed: %v", err.Error()))
	}

	if Config.jsonOutput {
		printVPCsJSON(sortVPCs(vpcs))
	} else {
		printVPCs(sortVPCs(vpcs))
	}
}

func init() {
	flag.BoolVar(&Config.noColor, "nocolor", false, "Suppresses color printing of listing")
	flag.BoolVar(&Config.Color, "color", false, "Force color output, even through pipe")
	flag.BoolVar(&Config.noSpace, "nospace", false, "Suppresses line-spacing of items")
	flag.BoolVar(&Config.allRegions, "all", false, "Fetches and prints data on all regions")
	flag.BoolVar(&Config.allRegions, "a", false, "Fetches and prints data on all regions (abbrev.)")
	flag.StringVar(&Config.regionOverride, "region", "", "Specify region (default: profile default region)")
	flag.StringVar(&Config.regionOverride, "r", "", "Specify region (default: profile default region) (abbrev.)")
	flag.BoolVar(&Config.jsonOutput, "j", false, "Output json instead of the typical textual output")
	flag.BoolVar(&Config.HideIP, "n", false, "do not display IP addresses and CIDRs (does not affect json output)")
	flag.BoolVar(&Config.Verbose, "v", false, "output verbose information about assets in vpc")
	flag.BoolVar(&Config.Truncate, "t", false, "truncate nametags")
}

func stdoutIsPipe() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		panic("Failed to stat stdout")
	}

	mode := info.Mode()

	return mode&fs.ModeNamedPipe != 0
}

func credentialsLoaded() bool {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		return false
	}

	if !creds.HasKeys() {
		return false
	}

	if aws.StringValue(sess.Config.Region) == "" {
		// No default region in profile/credentials, check that AWS_DEFAULT_REGION exists
		if os.Getenv("AWS_DEFAULT_REGION") == "" {
			os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
		}
	}

	return true
}

func validateRegion(region string) bool {
	regions := getRegions()
	isValid := false

	for _, reg := range regions {
		if region == reg {
			isValid = true
		}
	}

	return isValid
}

func main() {
	flag.Parse()

	if stdoutIsPipe() {
		if !Config.Color {
			Config.noColor = true
		}
	}

	if !credentialsLoaded() {
		fmt.Println("Failed to load aws credentials.")
		fmt.Println("Please set your AWS_PROFILE environment variable to a valid profile.")
		os.Exit(1)
	}

	// exit if region override is not valid
	if Config.regionOverride != "" && !validateRegion(Config.regionOverride) {
		fmt.Printf("Region: '%v' is not valid\n", Config.regionOverride)
		os.Exit(1)
	}

	switch {
	case Config.allRegions:
		doAllRegions()
	case Config.regionOverride != "":
		doSpecificRegion()
	default:
		doDefaultRegion()
	}
}
