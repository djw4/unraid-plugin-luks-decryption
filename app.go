package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	stsTypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
)

var (
	session_prefix     string = "unraid-secrets"
	keyfile_location   string = "/root/keyfile"
	default_aws_region string = "us-east-1"
)

type SSMGetParameterAPI interface {
	GetParameter(ctx context.Context,
		params *ssm.GetParameterInput,
		optFns ...func(*ssm.Options)) (*ssm.GetParameterOutput, error)
}

func findParameter(c context.Context, api SSMGetParameterAPI, input *ssm.GetParameterInput) (*ssm.GetParameterOutput, error) {
	return api.GetParameter(c, input)
}

func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
		panic(e)
	}
}

func main() {
	log.SetOutput(os.Stdout)

	// Allow the region to be set via command line argument, or default to 'us-east-1'
	pRegion := flag.String("region", default_aws_region, "The AWS region to operate in")
	pRoleARN := flag.String("role-arn", "", "The role ARN to use for accessing the secret (mandatory)")
	pParamPath := flag.String("param-path", "", "The SSM parameter path to retrieve (mandatory)")
	pKeyPath := flag.String("key-path", keyfile_location, "The path to write the keyfile to")
	required := []string{"role-arn", "param-path"}
	flag.Parse()

	seen := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { seen[f.Name] = true })
	for _, req := range required {
		if !seen[req] {
			log.Fatalf("Missing required '%v' argument", req)
			os.Exit(1)
		}
	}

	log.Printf("AWS Region: %s", *pRegion)
	log.Printf("AWS Role ARN: %s", *pRoleARN)
	log.Printf("AWS SSM Parameter Path: %s", *pParamPath)
	log.Printf("Output Key Path: %s", *pKeyPath)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(*pRegion),
	)
	check(err)
	sourceAccount := sts.NewFromConfig(cfg)

	// Assume the role and extract the credentials
	rand.Seed(time.Now().UnixNano())
	response, err := sourceAccount.AssumeRole(context.TODO(), &sts.AssumeRoleInput{
		RoleArn:         aws.String(*pRoleARN),
		RoleSessionName: aws.String(session_prefix + strconv.Itoa(10000+rand.Intn(25000))),
	})
	if err != nil {
		log.Fatalf("Unable to assume target role, %v", err)
		os.Exit(1)
	}
	var assumedRoleCreds *stsTypes.Credentials = response.Credentials

	cfg, err = config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(*pRegion),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				*assumedRoleCreds.AccessKeyId,
				*assumedRoleCreds.SecretAccessKey,
				*assumedRoleCreds.SessionToken),
		),
	)
	if err != nil {
		log.Fatalf("Unable to load static credentials for service client config, %v", err)
		os.Exit(1)
	}

	// Get the value of the relevant parameter
	ssmClient := ssm.NewFromConfig(cfg)

	input := &ssm.GetParameterInput{
		Name: aws.String(*pParamPath),
		WithDecryption: func() *bool { b := true; return &b }(),
	}

	results, err := findParameter(context.TODO(), ssmClient, input)
	if err != nil {
		fmt.Println(err.Error())
		return
	} else {
		f, err := os.Create(*pKeyPath)
		check(err)
		defer f.Close()

		w := bufio.NewWriter(f)
		k, err := w.WriteString(*results.Parameter.Value)
		check(err)
		log.Printf("Wrote keyfile to disk (%v bytes)", k)

		w.Flush()
	}
}
