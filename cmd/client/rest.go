/*
Copyright 2022 Luis Pabon

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
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lpabon/lputils"
	"github.com/lpabon/rpcscout/api"
	"github.com/sirupsen/logrus"
)

type ServerInfo struct {
	Id         string `json:"id,omitempty"`
	ServerName string `json:"serverName,omitempty"`
}

func (c *Client) rest(address string) {
	// Setup http
	clientName := "rest-" + c.opts().Name + "-" + lputils.GenUUID()[:6]
	log := logrus.New().WithFields(logrus.Fields{
		"type":   "REST",
		"client": clientName,
	})
	log.Info("REST client started")

	host := fmt.Sprintf("http://%s", address)

	// Delete
	c.loops.Add(func() error {
		type ScoutDeleteResponse struct {
			Server ServerInfo `json:"server,omitempty"`
		}

		var scoutResponse ScoutDeleteResponse
		jsonBuffer, err := json.Marshal(&api.ScoutDeleteRequest{
			Client: &api.ClientInfo{
				ClientName: clientName,
			},
			DeleteId: "deleteid",
		})
		if err != nil {
			log.Fatalf("unable to marsh ping request: %v", err)
		}

		req, err := http.NewRequest("DELETE", host+"/apis/scout/v1/data", bytes.NewBuffer(jsonBuffer))
		if err != nil {
			log.Fatalf("unable to setup request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		httpClient := http.Client{
			Timeout: time.Second * 3,
		}

		log.Infof("delete")
		start_time := time.Now()
		resp, err := httpClient.Do(req)
		end_time := time.Now()

		if err != nil {
			return fmt.Errorf("failed request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http request failed: %v", lputils.GetErrorFromResponse(resp))
		}

		err = lputils.GetJsonFromResponse(resp, &scoutResponse)
		if err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}

		log.WithFields(logrus.Fields{
			"latency": end_time.Sub(start_time),
			"server":  scoutResponse.Server.ServerName,
		}).Info("delete response received")

		return nil
	})

	// Ping
	c.loops.Add(func() error {
		type ScoutPingResponse struct {
			Server ServerInfo `json:"server,omitempty"`
		}

		var ping ScoutPingResponse
		jsonBuffer, err := json.Marshal(&api.ScoutPingRequest{
			Client: &api.ClientInfo{
				ClientName: clientName,
			},
		})
		if err != nil {
			log.Fatalf("unable to marsh ping request: %v", err)
		}

		req, err := http.NewRequest("POST", host+"/apis/scout/v1/ping", bytes.NewBuffer(jsonBuffer))
		if err != nil {
			log.Fatalf("unable to setup request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		httpClient := http.Client{
			Timeout: time.Second * 3,
		}

		log.Infof("PING >>")
		start_time := time.Now()
		resp, err := httpClient.Do(req)
		end_time := time.Now()

		if err != nil {
			return fmt.Errorf("failed request: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("http request failed: %v", lputils.GetErrorFromResponse(resp))
		}

		err = lputils.GetJsonFromResponse(resp, &ping)
		if err != nil {
			return fmt.Errorf("failed to parse response: %v", err)
		}

		log.WithFields(logrus.Fields{
			"latency": end_time.Sub(start_time),
			"server":  ping.Server.ServerName,
		}).Info("<< PONG")

		return nil
	})
}
