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
package main

import (
	"flag"
	"strings"

	"github.com/libopenstorage/grpc-framework/pkg/util"
	"github.com/lpabon/lputils"
	clientcmd "github.com/lpabon/rpcscout/cmd/client"
	pkgopts "github.com/lpabon/rpcscout/pkg/opts"
)

var (
	opts          pkgopts.Opts
	grpcAddresses string
	restAddresses string
)

func init() {
	flag.StringVar(&opts.Name, "name", "", "Name for this scout. If no name provided it will create one")
	flag.BoolVar(&opts.UseTls, "usetls", false, "Connect to server using TLS. Loads CA from the system")
	flag.StringVar(&opts.Token, "token", "", "Authorization token if any")
	flag.StringVar(&grpcAddresses, "grpc-addresses", "127.0.0.1:9009", "Comma separated addresses to Scout gRPC servers as <address>:<port>")
	flag.StringVar(&restAddresses, "rest-addresses", "127.0.0.1:9010", "Comma separated addresses to Scout REST servers as <address>:<port>")
	flag.IntVar(&opts.MaxPingDuration, "max-ping-duration", 10, "Maximum ping loop duration in seconds")
}

func main() {
	flag.Parse()
	opts.GrpcAddresses = strings.Split(grpcAddresses, ",")
	opts.RestAddresses = strings.Split(restAddresses, ",")

	// Setup name for client
	if opts.Name == "" {
		opts.Name = lputils.GenUUID()[:6]
	}

	c := clientcmd.New(&clientcmd.Config{
		Opts: &opts,
	})

	// Setup CTRL-C handler
	signal_handler := util.NewSigIntManager(func() {
		c.Stop()
	})
	signal_handler.Start()

	c.Start()

	// Wait for the signal handler to stop the program
	c.Wait()
}
