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
	"context"
	"fmt"
	"time"

	"github.com/lpabon/rpcscout/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ScoutServerConfig struct {
	Name string
}
type ScoutServer struct {
	api.ScoutServer
	api.ScoutIdentityServer

	name string
}

func NewScoutServer(config *ScoutServerConfig) *ScoutServer {
	return &ScoutServer{
		name: config.Name,
	}
}

func (s *ScoutServer) Ping(
	ctx context.Context,
	req *api.ScoutPingRequest,
) (*api.ScoutPingResponse, error) {
	logrus.Infof("received from client %s", req.GetClient().GetClientName())
	return &api.ScoutPingResponse{
		Server: &api.ServerInfo{
			Id:         req.GetClient().GetId(),
			ServerName: s.name,
			Timestamp:  timestamppb.Now(),
		},
	}, nil
}

func (s *ScoutServer) StreamData(
	req *api.ScoutStreamDataRequest,
	stream api.Scout_StreamDataServer,
) error {
	logrus.Infof("received stream data request from client %s", req.GetClient().GetClientName())

	var i int
	for i = 0; i < 10; i++ {
		response := &api.ScoutStreamDataResponse{
			Server: &api.ServerInfo{
				Id:         req.GetClient().GetId(),
				Timestamp:  timestamppb.Now(),
				ServerName: s.name,
			},
			RandomNumber: int32(i),
		}
		if err := stream.Send(response); err != nil {
			logrus.Errorf("unable to send data to client %s: %v",
				req.GetClient().GetClientName(),
				err)
			return fmt.Errorf("unable to send data to client %s: %v",
				req.GetClient().GetClientName(),
				err)
		}
		time.Sleep(time.Millisecond * 10)
	}

	return nil
}

func (s *ScoutServer) List(
	ctx context.Context,
	req *api.ScoutListRequest,
) (*api.ScoutListResponse, error) {
	logrus.Infof("list request from %s", req.GetClient().GetClientName())
	return &api.ScoutListResponse{
		List: []string{"a", "b", "c"},
	}, nil
}

func (s *ScoutServer) Delete(
	ctx context.Context,
	req *api.ScoutDeleteRequest,
) (*api.ScoutDeleteResponse, error) {
	logrus.Infof("delete %s request from %s",
		req.GetDeleteId(),
		req.GetClient().GetClientName())

	return &api.ScoutDeleteResponse{}, nil
}

func (s *ScoutServer) Version(
	ctx context.Context,
	req *api.ScoutIdentityVersionRequest,
) (*api.ScoutIdentityVersionResponse, error) {
	logrus.Info("Received request for version")
	return &api.ScoutIdentityVersionResponse{
		ScoutVersion: &api.ScoutVersion{
			Major: int32(api.ScoutVersion_MAJOR),
			Minor: int32(api.ScoutVersion_MINOR),
			Patch: int32(api.ScoutVersion_PATCH),
			Version: fmt.Sprintf("%d.%d.%d",
				api.ScoutVersion_MAJOR,
				api.ScoutVersion_MINOR,
				api.ScoutVersion_PATCH),
		},
	}, nil
}
