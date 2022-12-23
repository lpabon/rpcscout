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
	"fmt"
	"os"

	"github.com/libopenstorage/grpc-framework/pkg/util"
	"github.com/libopenstorage/grpc-framework/server"

	"github.com/lpabon/rpcscout/api"
	scoutserver "github.com/lpabon/rpcscout/pkg/server"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	scoutSocket = "/tmp/scout-server.sock"
)

func main() {
	scout := scoutserver.NewScoutServer(&scoutserver.ScoutServerConfig{
		Name: "scoutserver",
	})
	config := &server.ServerConfig{
		Name:         "scout",
		Address:      ":9009",
		Socket:       scoutSocket,
		AuditOutput:  os.Stdout,
		AccessOutput: os.Stdout,
	}
	config.
		WithDefaultRestServer("9010").
		RegisterGrpcServers(func(gs *grpc.Server) {
			api.RegisterScoutServer(gs, scout)
			api.RegisterScoutIdentityServer(gs, scout)
		}).
		RegisterRestHandlers(
			api.RegisterScoutHandler,
			api.RegisterScoutIdentityHandler,
		).WithDefaultRateLimiters()

	// Create grpc framework server
	os.Remove(scoutSocket)
	s, err := server.New(config)
	if err != nil {
		fmt.Printf("Unable to create server: %v", err)
		os.Exit(1)
	}

	// Setup a signal handler
	signal_handler := util.NewSigIntManager(func() {
		s.Stop()
		os.Remove(scoutSocket)
		os.Exit(0)
	})
	signal_handler.Start()

	// Start server
	err = s.Start()
	if err != nil {
		fmt.Printf("Unable to start server: %v", err)
		os.Exit(1)
	}

	// Wait. The signal handler will exit cleanly
	logrus.Info("Scout server running")
	select {}
}
