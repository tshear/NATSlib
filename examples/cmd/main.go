// Copyright 2012-2019 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/json-iterator/go"
	"github.com/nats-io/nats.go"
)

type NATSResponseHeader struct {
	Created      bool   `json:"created,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	Path         string `json:"path,omitempty"`
	Doc          string `json:"docId,omitempty"`
	DocVersion   string `json:"docVersion,omitempty"`
	Status       int    `json:"status"`
	ErrorStr     string `json:"error_str,omitempty"`
	ServerID     string `json:"serverID,omitempty"`
	EncryptedHdr []byte `json:"encrypted_hdr,omitempty"`
}

type NATSReqHeader struct {
	Mode          string                 `json:"mode"`
	Path          string                 `json:"path"`
	Flags         map[string]interface{} `json:"flags"`
	Authorization string                 `json:"authorization"`
}

type NATSRequest struct {
	Header NATSReqHeader `json:"header"`
	Body   []byte        `json:"body"`
}

type NATSResponse struct {
	Header   NATSResponseHeader `json:"header"`
	Response string             `json:"response"`
}

// NOTE: Can test with demo servers.
// nats-req -s demo.nats.io <subject> <msg>
// nats-req -s demo.nats.io:4443 <subject> <msg> (TLS version)

func usage() {
	log.Printf("Usage: micro-runner [-s server] [-creds file] <service-queue> <requestor> <reqpasscode <data-file>\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}

func main() {
	urls := flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
	userCreds := flag.String("creds", "", "User Credentials File")
	identity := flag.String("identity", "", "Microservice ID")
	requestor := flag.String("requestor", "", "Requestor ID")
	passCode := flag.String("passcode", "", "Requestor PassCode")
	dataFile := flag.String("dataFile", "", "Request Data File")
	showHelp := flag.Bool("h", false, "Show help message")

	flag.Parse()
	log.SetFlags(0)
	flag.Usage = usage

	if *showHelp {
		showUsageAndExit(0)
	}

	args := flag.Args()
	//if len(args) < 2 {
	//	showUsageAndExit(1)
	//}

	// Connect Options.
	opts := []nats.Option{nats.Name("DISP-NATS Microservice-runner")}

	// Use UserCredentials
	if *userCreds != "" {
		opts = append(opts, nats.UserCredentials(*userCreds))
	}

	// Connect to NATS
	nc, err := nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
	var json = jsoniter.ConfigCompatibleWithStandardLibrary
	subj := args[0]
	var token string

	file, err := os.Open(*dataFile) // For read access.
	if err != nil {
		fmt.Printf("open err %v\n", err)
		os.Exit(1)
	}
	dataBuffer := make([]byte, 4096)
	_, err = file.Read(dataBuffer)
	if err != nil {
		fmt.Printf("read err %v\n", err)
		os.Exit(1)
	}
	rbody := bytes.NewBuffer(dataBuffer)

	type User struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Expires  uint64 `json:"expires"`
	}

	type Token struct {
		Token string `json:"token"`
	}

	tk := &Token{}

	tbodyS := &User{}
	tbodyS.Username = *requestor
	tbodyS.Password = *passCode
	tbodyS.Expires = 999999999999
	tbody, err := json.Marshal(tbodyS)

	thdr := NATSReqHeader{
		Mode: "POST",
		Path: "/api/login",
	}
	trec := &NATSRequest{
		Header: thdr,
		Body:   tbody,
	}
	payload, err := json.Marshal(trec)
	msg, err := nc.Request(subj, payload, 2*time.Second)
	if err == nil {
		var response = &NATSResponse{}
		err = json.Unmarshal(msg.Data, response)
		err = json.Unmarshal([]byte(response.Response), tk)
		token = string(tk.Token)
		fmt.Printf("token '%v'\n", token)
	}

	rflags := make(map[string]interface{})
	rflags["identity"] = *identity

	rhdr := NATSReqHeader{
		Mode:          "POST",
		Path:          "/relation/register",
		Flags:         rflags,
		Authorization: token,
	}
	rrec := &NATSRequest{
		Header: rhdr,
		Body:   rbody.Bytes(),
	}

	var RDID string
	payload, err = json.Marshal(rrec)
	msg, err = nc.Request(subj, payload, 2*time.Second)
	if err == nil {
		var response = &NATSResponse{}
		err = json.Unmarshal(msg.Data, response)
		if response.Header.Status != 200 {
			fmt.Printf("RDID status %v error \"%s\"\n", response.Header.Status, response.Header.ErrorStr)
		} else {
			fmt.Printf("RDID status %v RDID %v\n", response.Header.Status, response.Response)
			RDID = response.Response
		}
	}

	dflags := make(map[string]interface{})
	dflags["entityAccess"] = "public"
	dflags["withHeader"] = true
	//dflags["requestor"] = "Testing1"
	//dflags["passCode"] = "g9BKjHkk0T0"
	dflags["domain"] = "tester18"
	dflags["entity"] = "clients"
	dflags["token"] = RDID
	dflags["aspect"] = "claims"
	dflags["k"] = "annotations"
	//dflags["v"] = "0"
	dflags["timestamp"] = "latest"

	mode := args[1]
	dhdr := NATSReqHeader{
		Mode: mode,
		Path: fmt.Sprintf("/%v/%v/%v/%v", dflags["domain"],
			dflags["entity"], dflags["token"], dflags["aspect"]),
		Flags:         dflags,
		Authorization: token,
	}
	drec := &NATSRequest{
		Header: dhdr,
		Body:   rbody.Bytes(),
	}

	start := time.Now()
	payload, err = json.Marshal(drec)

	msg, err = nc.Request(subj, payload, 2*time.Second)
	if err != nil {
		if nc.LastError() != nil {
			log.Fatalf("%v for request", nc.LastError())
		}
		log.Fatalf("%v for request", err)
	}
	var response = &NATSResponse{}
	err = json.Unmarshal(msg.Data, response)
	//log.Printf("struct %v\n",response)
	//log.Printf("Published [%s] : '%s'", subj, payload)
	if response.Header.Status != 200 {
		log.Printf("Received  [%v] : error_str '%v' %v", msg.Subject, response.Header.ErrorStr, response.Response)
	} else {
		log.Printf("Received  [%v] : %v", msg.Subject, response.Response)
	}
	log.Printf("Received response, elapsed %v", time.Since(start))
	nc.Close()

}
