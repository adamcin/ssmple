/*
 * Copyright 2018 Mark Adamcin
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ParsedArgs struct {
	// flag to enable credential resolution from ec2 endpoint
	UseEc2Role bool

	// pass-through profile and region args to aws sdk
	AwsProfile, AwsRegion string

	// get, put, delete, clear
	SsmCmd string

	// config directory for filename relative path resolution
	ConfDir string

	// KMS key ID or key alias for encrypting all params on put
	KeyIdPutAll string

	// true to overwrite existing params on put
	OverwritePut bool

	// true to clear all params in a path prefix before put
	ClearOnPut bool

	// true to avoid storing secure string parameters on get
	NoStoreSecureString bool

	// true to avoid sending secure strings on put
	NoPutSecureString bool

	// the slice of filenames, in order of declaration
	Filenames []string

	// the slice of path prefixes, in order of declaration
	Prefixes []string
}

const NoOptPrefix = "--no-"

func parseArgs() ParsedArgs {
	awsProfile := ""
	awsRegion := ""
	useEc2Role := false
	ssmCmd := ""
	rawConfDir := "."
	_, cwdErr := os.Getwd()
	if cwdErr != nil {
		log.Fatal("Failed to get current working directory")
	}

	filenames := make([]string, 0)
	prefixes := make([]string, 0)

	keyIdPutAll := ""
	overwritePut := false
	clearOnPut := false
	noStoreSecureString := false
	noPutSecureString := false
	isHelp := false

	for i := 1; i < len(os.Args); i++ {
		opt := os.Args[i]
		isNoOpt := strings.HasPrefix(opt, NoOptPrefix)
		if isNoOpt {
			opt = "--" + strings.TrimPrefix(opt, NoOptPrefix)
		}

		switch opt {
		case "-h", "--help":
			isHelp = true
		case "-p", "--profile":
			awsProfile = os.Args[i+1]
			i++
		case "-r", "--region":
			awsRegion = os.Args[i+1]
			i++
		case "--use-ec2-role":
			useEc2Role = !isNoOpt
		case "-C", "--conf-dir":
			rawConfDir = os.Args[i+1]
			i++
		case "-f", "--filename":
			filenames = append(filenames, os.Args[i+1])
			i++
		case "-s", "--starts-with":
			prefixes = append(prefixes, os.Args[i+1])
			i++
		case "-k", "--key-id-put-all":
			keyIdPutAll = os.Args[i+1]
			i++
		case "-o", "--overwrite-put":
			overwritePut = !isNoOpt
		case "--clear-on-put":
			clearOnPut = !isNoOpt
		case "--store-secure-string":
			noStoreSecureString = isNoOpt
		case "--put-secure-string":
			noPutSecureString = isNoOpt
		case "get", "put", "delete", "clear":
			ssmCmd = opt
		default:
			usage(ssmCmd)
			log.Fatal(fmt.Sprintf("Unrecognized option %s", opt))
		}
	}

	if isHelp {
		usage(ssmCmd)
		os.Exit(0)
	}

	if len(ssmCmd) == 0 {
		usage(ssmCmd)
		os.Exit(1)
	}

	confDir, confErr := filepath.Abs(rawConfDir)
	if confErr != nil {
		log.Fatal("Failed to resolve confDir "+rawConfDir, confErr)
	}

	if len(prefixes) == 0 {
		log.Fatal("At least one -s/--starts-with path is required, like /ecs/dev/myapp")
	}

	if len(filenames) == 0 {
		log.Fatal("At least one -f/--filename argument is required, like instance.properties")
	}

	return ParsedArgs{
		UseEc2Role:          useEc2Role,
		AwsProfile:          awsProfile,
		AwsRegion:           awsRegion,
		SsmCmd:              ssmCmd,
		ConfDir:             confDir,
		Filenames:           filenames,
		Prefixes:            prefixes,
		KeyIdPutAll:         keyIdPutAll,
		OverwritePut:        overwritePut,
		ClearOnPut:          clearOnPut,
		NoStoreSecureString: noStoreSecureString,
		NoPutSecureString:   noPutSecureString}
}

func getAwsConfigResolvers(authEc2 bool) []external.AWSConfigResolver {
	resolvers := []external.AWSConfigResolver{
		external.ResolveDefaultAWSConfig,
		external.ResolveCustomCABundle,
		external.ResolveRegion,
	}

	if authEc2 {
		resolvers = append(resolvers, external.ResolveFallbackEC2Credentials)
	}

	resolvers = append(resolvers, external.ResolveCredentialsValue)
	resolvers = append(resolvers, external.ResolveEndpointCredentials)
	resolvers = append(resolvers, external.ResolveContainerEndpointPathCredentials)
	resolvers = append(resolvers, external.ResolveAssumeRoleCredentials)

	return resolvers
}

func main() {
	prefs := parseArgs()

	var cfgs external.Configs
	var err error

	if len(prefs.AwsProfile) > 0 {
		cfgs = append(cfgs, external.WithSharedConfigProfile(prefs.AwsProfile))
	}

	if len(prefs.AwsRegion) > 0 {
		cfgs = append(cfgs, external.WithRegion(prefs.AwsRegion))
	}

	if cfgs, err = cfgs.AppendFromLoaders(external.DefaultConfigLoaders); err != nil {
		log.Fatal(err)
	}

	resolvers := getAwsConfigResolvers(prefs.UseEc2Role)

	awsCfg, err := cfgs.ResolveAWSConfig(resolvers)
	if err != nil {
		log.Fatal(err)
	}

	execCmd(prefs, awsCfg)
}

func execCmd(prefs ParsedArgs, cfg aws.Config) {
	ssms := ssm.New(cfg)
	kmss := kms.New(cfg)

	fileStores := make(map[string]*FileStore, len(prefs.Filenames))
	for _, fn := range prefs.Filenames {
		fs := NewFileStore(prefs.ConfDir, fn)
		if err := fs.Load(); err != nil {
			log.Fatalf("Failed to load file store for name %s. reason: %s", fn, err)
		}
		fileStores[fn] = &fs
	}

	kmsMap := KmsMap{
		aliasesToKeys: make(map[string]string, 0),
		keysToAliases: make(map[string]string, 0)}

	ctx := CmdContext{
		Prefs:  prefs,
		Stores: fileStores,
		Ssms:   ssms,
		KmsMap: kmsMap}

	switch strings.ToLower(prefs.SsmCmd) {
	case "get":
		if !prefs.NoStoreSecureString {
			buildAliasList(kmss, &kmsMap)
		}
		doGet(&ctx)
	case "put":
		if !prefs.NoPutSecureString {
			buildAliasList(kmss, &kmsMap)
		}
		doPut(&ctx)
	case "delete":
		doDelete(&ctx)
	case "clear":
		doClear(&ctx)
	default:
		log.Fatalf("Unknown command %s", prefs.SsmCmd)
	}
}
