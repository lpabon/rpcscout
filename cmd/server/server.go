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
package server

import (
	"os"
	"strings"

	"github.com/lpabon/rpcscout/api"
	scoutserver "github.com/lpabon/rpcscout/pkg/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	grpcfwserver "github.com/libopenstorage/grpc-framework/server"
	pkgopts "github.com/lpabon/rpcscout/pkg/opts"
)

type Config struct {
	Opts *pkgopts.Opts
}

type Server struct {
	config       *Config
	serverConfig *grpcfwserver.ServerConfig
	grpcfwServer *grpcfwserver.Server
}

const (
	scoutSocket = "/tmp/scout-server.sock"
)

func New(config *Config) *Server {
	return &Server{
		config: config,
	}
}

func (s *Server) Start() {
	scout := scoutserver.NewScoutServer(&scoutserver.ScoutServerConfig{
		Name: s.opts().Name,
	})
	config := &grpcfwserver.ServerConfig{
		Name:         "scout",
		Address:      s.opts().GrpcListen,
		Socket:       scoutSocket,
		AuditOutput:  os.Stdout,
		AccessOutput: os.Stdout,
	}
	config.
		RegisterGrpcServers(func(gs *grpc.Server) {
			api.RegisterScoutServer(gs, scout)
			api.RegisterScoutIdentityServer(gs, scout)
		}).WithDefaultRateLimiters()

	// Enable REST if requested
	if s.opts().RestListen != "" {
		parts := strings.Split(s.opts().RestListen, ":")
		port := parts[0]
		if len(parts) > 1 {
			port = parts[1]
		}

		config.
			WithDefaultRestServer(port).
			RegisterRestHandlers(
				api.RegisterScoutHandler,
				api.RegisterScoutIdentityHandler,
			)
	}

	// Create grpc framework server
	os.Remove(scoutSocket)
	var err error
	s.grpcfwServer, err = grpcfwserver.New(config)
	if err != nil {
		logrus.Fatalf("Unable to create server: %v", err)
	}

	// Start server
	err = s.grpcfwServer.Start()
	if err != nil {
		logrus.Fatalf("Unable to start server: %v", err)
	}

	// Wait. The signal handler will exit cleanly
	logrus.Info("Scout server running")
}

func (s *Server) Stop() {
	s.grpcfwServer.Stop()
	os.Remove(scoutSocket)
}

func (s *Server) opts() *pkgopts.Opts {
	return s.config.Opts
}
