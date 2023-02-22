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
	"github.com/lpabon/rpcscout/pkg/loop"
	pkgopts "github.com/lpabon/rpcscout/pkg/opts"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Opts *pkgopts.Opts
}

type Client struct {
	config *Config
	loops  *loop.Loop
}

func New(config *Config) *Client {
	return &Client{
		config: config,
		loops:  loop.NewLoop(config.Opts.MaxPingDuration),
	}
}

func (c *Client) Start() {
	for _, address := range c.opts().GrpcAddresses {
		if address == "" {
			continue
		}
		c.grpc(address)
	}
	for _, address := range c.opts().RestAddresses {
		if address == "" {
			continue
		}
		c.rest(address)
	}

	c.loops.Start()
}

func (c *Client) Stop() {
	c.loops.Stop()
	logrus.Info("Shutting down...")
}

func (c *Client) Wait() {
	c.loops.Wait()
}

func (c *Client) opts() *pkgopts.Opts {
	return c.config.Opts
}
