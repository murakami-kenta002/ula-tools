// SPDX-License-Identifier: Apache-2.0
/**
 * Copyright (c) 2024  Panasonic Automotive Systems, Co., Ltd.
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
	"encoding/json"
	"flag"
	"fmt"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/ulamulticonn"
	. "ula-tools/internal/ulog"
	"io/ioutil"
	"os"
)

func readStdinCommand() (string, error) {
	var err error
	/* read json command */
	bs, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		ELog.Printf("Json Command ReadAll error: %s \n", err)
		os.Exit(1)
	}

	jsonComm := string(bs)
	DLog.Println(jsonComm)

	mJson := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonComm), &mJson)
	if err != nil {
		ELog.Printf("Unmarshal json command error: %s \n", err)
		os.Exit(1)
	}

	layoutComm, err := json.Marshal(&mJson)
	if err != nil {
		ELog.Printf("Marshal error: %s \n", err)
		os.Exit(1)
	}

	return string(layoutComm), err
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "%s [option] VScrnDefFile < json command\n", os.Args[0])

	fmt.Fprintf(os.Stderr, "[option]\n")
	flag.PrintDefaults()
}

func main() {

	flag.Usage = printUsage

	var (
		force   bool
		verbose bool
		debug   bool
	)

	flag.BoolVar(&force, "f", false, "force layout control even if some nodes are not alive.")
	flag.BoolVar(&verbose, "v", true, "verbose info log")
	flag.BoolVar(&debug, "d", false, "verbose debug log")
	flag.Parse()

	if !force {
		force = ula.GetEnvBool("ULA_FORCE", false)
	}

	if verbose == true {
		ILog.SetOutput(os.Stderr)
	}

	if debug == true {
		DLog.SetOutput(os.Stderr)
	}

	var vsdPath string
	args := flag.Args()
	if len(args) > 0 {
		DLog.Printf("ARG0:%s", flag.Arg(0))
		vsdPath = flag.Arg(0)
	}

	var err error
	err = ulamulticonn.UlaConnectionInit(force, vsdPath)
	if err != nil {
		ELog.Println(err)
		os.Exit(1)
	}

	layoutComm, err := readStdinCommand()
	if err != nil {
		ELog.Println(err)
		os.Exit(1)
	}

	err = ulamulticonn.UlaMulCon.SendLayoutCommand(layoutComm)
	if err != nil {
		ELog.Println(err)
		os.Exit(1)
	}
}
