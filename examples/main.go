/*
Copyright © 2019 Dataparency, LLC mailto:dev@dataparency.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	_ "net/http"
	"time"

	"github.com/nats-io/nats.go"
)

var topic = "micro-master"

type NATSUpdateEHeader struct {
	Mode 		string	`json:"mode"`
	Entity 		string	`json:"entity"`
	EntityPath	string	`json:"entity_path"`
	ServerID	string	`json:"serverID,omitempty"`
}

type NATSEntityUpdate struct {
	Header 		NATSUpdateEHeader `json:"header"`
	Buffer		[]byte			`json:"buffer"`
}

type NATSUpdateDHeader struct {
	Created     bool   	`json:"created,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
	Path       	string  `json:"path,omitempty"`
	Doc        	string  `json:"docId,omitempty"`
	DocVersion 	string  `json:"docVersion,omitempty"`
	Expiry		int64	`json:"expiry"`
	ServerID	string	`json:"serverID,omitempty"`
	Entity		string	`json:"entity"`
	EntityPath	string	`json:"entityPath,omitempty"`
	DocPath		string	`json:"docPath,omitempty"`
	TArray  	[]string `json:"tarray,omitempty"`
	EntityAccess string	`json:"entityAccess"`
	RDID		string	`json:"rdid,omitempty"`
}

type NATSDataUpdate struct {
	Header 		NATSUpdateDHeader `json:"header"`
	Buffer		[]byte			`json:"buffer"`
}

type NATSResponseHeader struct {
	Created     bool   	`json:"created,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
	Path       string   `json:"path,omitempty"`
	Doc        string   `json:"docId,omitempty"`
	DocVersion string   `json:"docVersion,omitempty"`
	Status		int		`json:"status"`
	ErrorStr	string	`json:"error_str,omitempty"`
	ServerID	string 	`json:"serverID,omitempty"`
	Chunks		int		`json:"chunks,omitempty"`
	EncryptedHdr	[]byte	`json:"encrypted_hdr,omitempty"`
}

type NATSReqHeader struct {
	Mode 		string		`json:"mode"`
	Path		string		`json:"path"`
	Flags 		map[string]interface{}	`json:"flags"`
	Authorization 	string	`json:"authorization"`
	Accept		string		`json:"accept"`
}

type NATSRequest struct {
	Header		NATSReqHeader `json:"header"`
	Body		[]byte	`json:"body"`
}

type NATSResponse struct {
	Header 		NATSResponseHeader 	`json:"header"`
	Response	string				`json:"response"`
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}

func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'\n", i, m.Subject, string(m.Data))
}

func processReq(msg *nats.Msg, conn *nats.Conn) {
	//printMsg(msg, i)

	req := &NATSRequest{}
	_ = json.Unmarshal(msg.Data, &req)
	fmt.Printf("req %v\n",req)
	/*start := time.Now()
	handler := GetRoute(nr, req.Header)
	if handler != nil {
		resp := handler(req)
		log.Printf(
			"%s\t%s\t%s\t%s",
			req.Header.Mode,
			req.Header.Path,
			"disp-requests",
			time.Since(start),
		)
		if len(resp.Response) > int(conn.MaxPayload()) {
			resp.Header.ErrorStr = "max payload exceeded"
			resp.Header.Status = 422
			fmt.Printf("must chunk sz %v\n",len(resp.Response))
			resp.Header.Chunks = 2
			resp.Response = "error"
			response, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("json marshall err %v\n", err)
			}
			_ = msg.Respond(response)
		}

	 */
		resp := &NATSResponse{}
		resp.Response = "master micro response"
		response, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("json marshall err %v\n", err)
		}
		_ = msg.Respond(response)
	//} else {
	//	_ = msg.Respond([]byte("method not found"))
	//}
	_ = conn.Flush()
}

func connectNATS() {
	// Connect Options.
	opts := []nats.Option{nats.Name("D_ISP Responder")}
	opts = setupConnOptions(opts)
	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL, opts...)
	if err != nil {
		log.Fatal(err)
	}
	mp := nc.MaxPayload()
	log.Printf("Maximum payload is %v bytes", mp)
	i := 0
	sc, err := nc.QueueSubscribeSync(topic, "disp-micro")
	if err != nil {
		fmt.Printf("Oh shit %v\n",err)
	}

	msg,err := sc.NextMsg(2 * time.Hour)
	if err != nil {
		fmt.Printf("sub err %v\n",err)
	} else {
		i++
		processReq(msg,nc)
	}

}

func main() {
	log.Printf("Dataparency\u2122 dpmicroserver started on topic %v\n",topic)
	for {
		connectNATS()
	}

}