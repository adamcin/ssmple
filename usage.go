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
	"os"
	"path/filepath"
)

var argv0 = filepath.Base(os.Args[0])

func usage(operation string) {

	globalHelp := fmt.Sprintf(`GENERAL USAGE

  %[1]s [ global opts ] <operation> [ options ]

GLOBAL OPTIONS

  -h | --help                           : print this help message
  -p | --profile                        : set AWS profile
  -r | --region                         : set AWS region
       --use-ec2-role                   : allow attempt to resolve EC2 instance role credentials from host endpoint`, argv0)

	fmt.Println(globalHelp)
	fmt.Println(help(operation))
}

func help(operation string) string {
	switch operation {
	case "get":
		return helpGet()
	case "put":
		return helpPut()
	case "delete":
		return helpDelete()
	case "clear":
		return helpClear()
	default:
		return helpOperations()
	}
}

func helpGet() string {
	return fmt.Sprintf(`
OPERATION

  get                                   : Download SSM parameter values matching each specified filename, 
                                          for one or more -s path prefixes, and merge param values according 
                                          to -s declaration sequence.

    USAGE

      %[1]s get [ --no-store-secure-string ] -s <prefix> [ [ -s <prefix> ] ... ] [ -C <confDir> ] 
            -f filename [ [ -f filename ] ... ]

    OPTIONS

      -s | --starts-with                : specify an SSM parameter path prefix. When more than one -s argument is specified,
                                          they are evaluated in the order they are supplied.
      -C | --conf-dir                   : specify a base configuration directory, against which filenames are resolved relatively. Defaults to $CWD.
      -f | --filename                   : specify a configuration filename. this is resolved as a path relative to the -C confDir, and the basename of
                                          the filename (filename minus last extension) is treated as a suffix appended to each SSM param path prefix in turn.
           --no-store-secure-string     : if a parameter is of type SecureString, it will not be saved to the file.

    EXAMPLES

      1. Simplest Case

           %[1]s get -s /ep/conf -C /root/ep/conf -f ecs.properties

         Get SSM parameters matching /ep/conf/ecs/* and store them in a local file at path /root/ep/conf/ecs.properties.

      2. Multiple files
      
           %[1]s get -s /ep/conf -C /root/ep/conf -f ecs.properties -f tomcat.properties

         Get SSM parameters named /ep/conf/ecs/* and /ep/conf/tomcat/*, and store them at paths /root/ep/conf/ecs.properties 
         and /root/ep/conf/tomcat.properties, respectively.

      3. Multiple files, Multiple prefixes
      
           %[1]s get -s /ep/conf -s /ep/conf/prod -C /root/ep/conf -f ecs.properties -f tomcat.properties

         Get SSM parameters named /ep/conf/ecs/* and /ep/conf/prod/ecs/*, and /ep/conf/tomcat/* and /ep/conf/prod/tomcat/*, and store them at paths 
         /root/ep/conf/ecs.properties and /root/ep/conf/tomcat.properties, respectively.
`, argv0)
}

func helpPut() string {
	return fmt.Sprintf(`
OPERATION

  put                                   : Upload new parameter values to a single path prefix, from one or
                                          more specified filenames.

    USAGE

      %[1]s put [ --no-put-secure-string ] [ --overwrite-put | --clear-on-put ] [ --key-id-put-all <keyId|keyAlias> ]
            -s <prefix> [ -C <confDir> ] -f filename [ [ -f filename ] ... ]

    OPTIONS

      -s | --starts-with                : specify an SSM parameter path prefix. When more than one -s argument is specified,
                                          they are evaluated in the order they are supplied.
      -C | --conf-dir                   : specify a base configuration directory, against which filenames are resolved relatively. Defaults to $CWD.
      -f | --filename                   : specify a configuration filename. this is resolved as a path relative to the -C confDir, and the basename of
                                          the filename (filename minus last extension) is treated as a suffix appended to each SSM param path prefix in turn.
      -k | --key-id-put-all             : specify a KMS key ID or key alias to use to encrypt all uploaded parameters as SecureStrings.
      -o | --overwrite-put              : normally, the command will fail if you attempt to put a parameter that already exists in SSM. use this flag to
                                          overwrite any existing values in that situation.
           --clear-on-put               : convenience flag to first delete all parameters at the specified parameter path prefix.
           --no-put-secure-string       : if a property has a buddy _SecureStringKeyId property, it will not be uploaded to SSM.

    EXAMPLES

      1. Simplest Case

           %[1]s put -s /ep/conf -C /root/ep/conf -f ecs.properties

         Read values from /root/ep/conf/ecs.properties, and create SSM parameters at path prefix /ep/conf/ecs, with parameter names matching the
         property keys.

      2. Multiple files
      
           %[1]s put -s /ep/conf -C /root/ep/conf -f ecs.properties -f tomcat.properties

         Read values from /root/ep/conf/ecs.properties and /root/ep/conf/tomcat.properties, and create SSM parameters at path prefixes /ep/conf/ecs
         and /ep/conf/tomcat, respectively.
`, argv0)
}

func helpDelete() string {
	return fmt.Sprintf(`
OPERATION

  delete                                : Delete SSM parameters within a particular path prefix according to
                                          parameter names present in one or more specified filenames.

    USAGE

      %[1]s delete -s <prefix> [ -C <confDir> ] -f filename [ [ -f filename ] ... ] 

    OPTIONS

      -s | --starts-with                : specify an SSM parameter path prefix. When more than one -s argument is specified,
                                          they are evaluated in the order they are supplied.
      -C | --conf-dir                   : specify a base configuration directory, against which filenames are resolved relatively. Defaults to $CWD.
      -f | --filename                   : specify a configuration filename. this is resolved as a path relative to the -C confDir, and the basename of 
                                          the filename (filename minus last extension) is treated as a suffix appended to each SSM param path prefix in turn.

    EXAMPLES

      1. Simplest Case

           %[1]s delete -s /ep/conf -C /root/ep/conf -f ecs.properties

         Read property keys from /root/ep/conf/ecs.properties, and delete any SSM parameters at path prefix /ep/conf/ecs whose parameter names are 
         present in the file.

      2. Multiple files
      
           %[1]s put -s /ep/conf -C /root/ep/conf -f ecs.properties -f tomcat.properties

         Read property keys from /root/ep/conf/ecs.properties and /root/ep/conf/tomcat.properties, and delete any SSM parameters at path prefixes 
         /ep/conf/ecs and /ep/conf/tomcat, respectively, whose parameter names are present as keys in the associated file.
`, argv0)
}

func helpClear() string {
	return fmt.Sprintf(`
OPERATION

  clear                                 : Delete ALL SSM parameters within in the specified path prefix.

    USAGE

      %[1]s clear -s <prefix>

    OPTIONS

      -s | --starts-with                : specify an SSM parameter path prefix. When more than one -s argument is specified,
                                          they are evaluated in the order they are supplied.

    EXAMPLES

      1. Simplest Case

           %[1]s clear -s /ep/conf

         Read property keys from /root/ep/conf/ecs.properties, and delete any SSM parameters at path prefix /ep/conf/ecs whose parameter names are 
         present in the file.
`, argv0)
}

func helpOperations() string {
	return fmt.Sprintf(`
  Specify %[1]s -h <operation> to see detailed help for one of the following operations.

OPERATIONS

  get                                   : Download SSM parameter values matching each specified filename, 
                                          for one or more -s path prefixes, and merge param values according 
                                          to -s declaration sequence.

  put                                   : Upload new parameter values to a single path prefix, from one or
                                          more specified filenames.

  delete                                : Delete SSM parameters within a particular path prefix according to
                                          parameter names present in one or more specified filenames.

  clear                                 : Delete ALL SSM parameters within in the specified path prefix.
`, argv0)
}
