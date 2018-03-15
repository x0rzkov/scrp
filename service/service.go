/*
 *
 * Copyright 2015 gRPC authors.
 * & Andrew Grosser
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
 *
 */

//Example
//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	pb "github.com/dioptre/gtscrp/proto"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port = ":50551"
)

type server struct{}

// SayHello implements proto.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

// Scrape implements proto.ScaperServer
func (s *server) Scrape(ctx context.Context, in *pb.ScrapeRequest) (*pb.ScrapeReply, error) {

	// Instantiate default collector
	c := colly.NewCollector(
		// Turn on asynchronous requests
		colly.Async(true),
		// Attach a debugger to the collector
		colly.Debugger(&debug.LogDebugger{}),
	)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	var glob string
	if in.Filter == "" {
		glob = "*"
	} else {
		glob = in.Filter
	}
	c.Limit(&colly.LimitRule{
		DomainGlob:  glob,
		Parallelism: 1,
		Delay:       7 * time.Second,
	})

	// Start scraping in five threads on https://httpbin.org/delay/2
	for i := 0; i < 5; i++ {
		defer c.Visit(fmt.Sprintf("%s?n=%d", in.Url, i))
		//TODO: add to cassandra
		//Try and add a sharding value to help stop double ups
	}
	// Wait until threads are finished
	defer c.Wait()
	return &pb.ScrapeReply{Message: true}, nil
}

//TODO: get from cassandra, mix the following
//Run different clients for each domain, each at rate limited speeds
func crawl() {
	// UPDATE users
	// SET email = ‘janedoe@abc.com’
	// WHERE login = 'jdoe'
	// IF email = ‘jdoe@abc.com’;

	// 	BEGIN BATCH
	//   INSERT INTO purchases (user, balance) VALUES ('user1', -8) IF NOT EXISTS;
	//   INSERT INTO purchases (user, expense_id, amount, description, paid)
	//     VALUES ('user1', 1, 8, 'burrito', false);
	// APPLY BATCH;
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Read cert and key file
	BackendCert, _ := ioutil.ReadFile("./backend.cert")
	BackendKey, _ := ioutil.ReadFile("./backend.key")

	// Generate Certificate struct
	cert, err := tls.X509KeyPair(BackendCert, BackendKey)
	if err != nil {
		log.Fatalf("failed to parse certificate: %v", err)
	}

	// Create credentials
	creds := credentials.NewServerTLSFromCert(&cert)

	// Use Credentials in gRPC server options
	serverOption := grpc.Creds(creds)
	var s = grpc.NewServer(serverOption)
	defer s.Stop()

	pb.RegisterScraperServer(s, &server{})
	fmt.Printf("Server up on %s\n", port)
	// Register reflection service on gRPC server.
	// reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
