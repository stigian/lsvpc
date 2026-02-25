// Copyright 2023 Stigian Consulting - reference license in top level of project
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func getRegions() []string {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err.Error()))
	}

	svc := ec2.NewFromConfig(cfg)
	regions := []string{}

	res, err := svc.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		panic(fmt.Sprintf("Could not get regions: %v", err.Error()))
	}

	for _, region := range res.Regions {
		regions = append(regions, aws.ToString(region.RegionName))
	}

	return regions
}
