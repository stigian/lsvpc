// Copyright 2023 Stigian Consulting - reference license in top level of project
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getRegions() []string {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := ec2.New(sess)
	regions := []string{}

	res, err := svc.DescribeRegions(&ec2.DescribeRegionsInput{})
	if err != nil {
		panic(fmt.Sprintf("Could not get regions: %v", err.Error()))
	}

	for _, region := range res.Regions {
		regions = append(regions, aws.StringValue(region.RegionName))
	}

	return regions
}
