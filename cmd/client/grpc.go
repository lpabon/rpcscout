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
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"time"

	"github.com/lpabon/lputils"
	"github.com/lpabon/rpcscout/api"
	pkgopts "github.com/lpabon/rpcscout/pkg/opts"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/libopenstorage/grpc-framework/pkg/correlation"
	grpcfwclient "github.com/libopenstorage/grpc-framework/pkg/grpc/client"
	"github.com/sirupsen/logrus"
)

type ScoutToken struct {
	opts *pkgopts.Opts
}

func (t *ScoutToken) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "bearer " + t.opts.Token,
	}, nil
}

func (t *ScoutToken) RequireTransportSecurity() bool {
	return t.opts.UseTls
}

func (c *Client) grpc(address string) {
	clientName := "grpc-" + c.opts().Name + "-" + lputils.GenUUID()[:6]
	log := logrus.New().WithFields(logrus.Fields{
		"type":   "gRPC",
		"client": clientName,
	})
	log.Info("gRPC client started")
	// There are two ways to setup a token:
	//   - One is to setup a client interceptor which adds the token
	//     to every call automatically using grpc.WithPerRPCCredentials().
	//   - Second way is just to add it to the context directly as follows:
	//   import "google.golang.org/grpc/metadata"
	//   md := metadata.New(map[string]string{
	//		"authorization": "bearer" + token,
	//	 })
	//   ctx := metadata.NewOutgoingContext(context.Background(), md)
	//
	//   We will be using the more complicated first model to show how it can be done
	//
	//   To accomplish this, we first need to create an object that satisfies the
	//   interface needed by grpc.WithPerRPCCredentials(..)
	contextToken := &ScoutToken{
		opts: c.config.Opts,
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if c.opts().UseTls {
		// Setup a connection
		capool, err := x509.SystemCertPool()
		if err != nil {
			log.Fatalf("Failed to load system certs: %v\n", err)
		}
		dialOptions = []grpc.DialOption{grpc.WithTransportCredentials(
			credentials.NewClientTLSFromCert(capool, ""),
		)}
	}

	if len(c.opts().Token) != 0 {
		// Add token interceptor
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(contextToken))
	}

	// Connect to server
	conn, err := grpcfwclient.Connect(address, dialOptions)
	if err != nil {
		log.Fatalf("Could not connect to server %s: %v", address, err)
	}
	c.loops.AddTeardown(func() {
		conn.Close()
	})

	// LIST
	c.loops.Add(func() error {
		// Setup ping client message
		scoutping := api.NewScoutClient(conn)
		request := api.ScoutListRequest{
			Client: &api.ClientInfo{
				ClientName: clientName,
				Id:         lputils.GenUUID()[:4],
				Timestamp:  timestamppb.Now(),
			},
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		log.WithFields(logrus.Fields{
			"id":        request.GetClient().GetId(),
			"timestamp": request.GetClient().GetTimestamp().AsTime(),
		}).Info("list request")

		start_time := time.Now()
		resp, err := scoutping.List(ctx, &request)
		end_time := time.Now()

		if err != nil {
			return fmt.Errorf("message error: %v", err)
		}

		server := resp.GetServer()
		log.WithFields(logrus.Fields{
			"id":        server.GetId(),
			"server":    server.GetServerName(),
			"timestamp": server.GetTimestamp().AsTime(),
			"latency":   end_time.Sub(start_time),
		}).Infof("list response: %+v", resp.GetList())

		return nil
	})

	// PING
	c.loops.Add(func() error {
		// setup correlation
		ctx := correlation.WithCorrelationContext(context.Background(), correlation.Component("rpcscout-client"))

		// Setup ping client message
		scoutping := api.NewScoutClient(conn)
		request := api.ScoutPingRequest{
			Client: &api.ClientInfo{
				ClientName: clientName,
				Id:         lputils.GenUUID()[:4],
				Timestamp:  timestamppb.Now(),
			},
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second)

		log.WithFields(logrus.Fields{
			"id":        request.GetClient().GetId(),
			"timestamp": request.GetClient().GetTimestamp().AsTime(),
		}).Info("PING ==>")
		start_time := time.Now()
		resp, err := scoutping.Ping(ctx, &request)
		end_time := time.Now()
		cancel()

		if err != nil {
			return fmt.Errorf("message errored: %v", err)
		}
		server := resp.GetServer()
		log.WithFields(logrus.Fields{
			"id":        server.GetId(),
			"server":    server.GetServerName(),
			"timestamp": server.GetTimestamp().AsTime(),
			"latency":   end_time.Sub(start_time),
		}).Info("<== PONG")

		return nil
	})

	c.loops.Add(func() error {
		// setup correlation
		ctx := correlation.WithCorrelationContext(context.Background(), correlation.Component("rpcscout-client"))

		// Setup ping client message
		scout := api.NewScoutClient(conn)
		request := &api.ScoutStreamDataRequest{
			Client: &api.ClientInfo{
				ClientName: clientName,
				Id:         lputils.GenUUID()[:4],
				Timestamp:  timestamppb.Now(),
			},
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*60)
		defer cancel()

		log.WithFields(logrus.Fields{
			"id":        request.GetClient().GetId(),
			"timestamp": request.GetClient().GetTimestamp().AsTime(),
		}).Info("starting stream")

		stream, err := scout.StreamData(ctx, request)
		if err != nil {
			return fmt.Errorf("unable to call StreamData: %v", err)
		}

		messages := 0
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("unable to receive from StreamData stream: %v", err)
			}
			messages++

			server := resp.GetServer()
			serverTime := server.GetTimestamp().AsTime()
			log.WithFields(logrus.Fields{
				"messages_received": messages,
				"id":                server.GetId(),
				"server":            server.GetServerName(),
				"timestamp":         server.GetTimestamp().AsTime(),
				"latency":           time.Since(serverTime),
			}).Infof("stream received random number: %d", resp.GetRandomNumber())
		}

		return nil
	})
}
