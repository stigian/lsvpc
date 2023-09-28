// Copyright 2021 Stigian Consulting - reference license in top level of project

package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/stigian/lsvpc/awsfetch"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

	vpcs := make(map[string]*VPC)
	fetch := awsfetch.New(sess)

	recieved, err := fetch.GetAll()
	if err != nil {
		return map[string]*VPC{}, err
	}

	mapVpcs(vpcs, recieved.Vpcs.Vpcs)
	mapSubnets(vpcs, recieved.Subnets.Subnets)
	mapInstances(vpcs, recieved.Instances.Instances)
	mapInstanceStatuses(vpcs, recieved.InstanceStatuses.InstanceStatuses)
	mapVolumes(vpcs, recieved.Volumes.Volumes)
	mapNatGateways(vpcs, recieved.NatGateways.NatGateways)
	mapRouteTables(vpcs, recieved.RouteTables.RouteTables)
	mapInternetGateways(vpcs, recieved.InternetGateways.InternetGateways)
	mapEgressOnlyInternetGateways(vpcs, recieved.EOInternetGateways.EOInternetGateways)
	mapVPNGateways(vpcs, recieved.VPNGateways.VPNGateways)
	mapTransitGatewayVpcAttachments(vpcs, recieved.TransiGateways.TransitGateways, recieved.Identity.Identity)
	mapVpcPeeringConnections(vpcs, recieved.PeeringConnections.PeeringConnections)
	mapVpcEndpoints(vpcs, recieved.VPCEndpoints.VPCEndpoints)
	mapNetworkInterfaces(vpcs, recieved.NetworkInterfaces.NetworkInterfaces)
	mapSecurityGroups(vpcs, recieved.SecurityGroups.SecurityGroups)

	return vpcs, nil
}

func getRegionData(region string, out chan RegionData) {
	defer close(out)

	vpcs, err := populateVPC(region)
	if err != nil {
		out <- RegionData{}
	} else {
		out <- RegionData{
			VPCs: vpcs,
		}
	}
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
	regions := getRegions()
	fullData := make(map[string]RegionData)
	channels := make(map[string]chan RegionData)

	for _, region := range regions {
		channels[region] = make(chan RegionData)
		go getRegionData(region, channels[region])
	}
	for _, region := range regions {
		fullData[region] = <-channels[region]
	}

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
