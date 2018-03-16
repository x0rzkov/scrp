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

//TODO: AG
//Use GO_REUSEPORT listener
//Run a separate server instance per CPU core with GOMAXPROCS=1 (it appears during benchmarks that there is a lot more context switches with Traefik than with nginx)

//Example
//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto
package main

//os.Getenv("MACHINE_NAME")

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	pb "github.com/dioptre/gtscrp/proto"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/satori/go.uuid"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	port     = ":50551"
	internal = false
)

type server struct{}

// SayHello implements proto.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

// Scrape implements proto.ScaperServer
func (s *server) Scrape(ctx context.Context, in *pb.ScrapeRequest) (*pb.ScrapeReply, error) {

	//Notify the dispatcher of a new URL
	if in.Id == "" {
		u, _ := uuid.NewV1()
		in.Id = u.String()
	}
	received <- in
	return &pb.ScrapeReply{Message: true}, nil
}

func scrape(in *pb.ScrapeRequest) {
	//ctx := context.Background()
	// Instantiate default collector
	c := colly.NewCollector(
		colly.Async(true), // Turn on asynchronous requests
		// Attach a debugger to the collector
		colly.Debugger(&debug.LogDebugger{}),
	)

	if in.Filter != "" && in.Filter != "_" {
		filterStrings := strings.Split(in.Filter, "|")
		filters := make([]*regexp.Regexp, len(filterStrings))
		for i := range filterStrings {
			filters[i] = regexp.MustCompile(filterStrings[i])
		}
		colly.URLFilters(filters...)(c)
	}

	if in.Domain != "" && in.Domain != "_" {
		colly.AllowedDomains(in.Domain)(c)
	}

	// Limit the number of threads started by colly to one
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		Parallelism: 1,
		//Delay:       7 * time.Second,
		RandomDelay: 14 * time.Second,
	})

	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if internal {
			received <- &pb.ScrapeRequest{Id: in.Id, Url: e.Attr("href"), Domain: in.Domain, Filter: in.Filter}
		} else {
			Rescrape(&pb.ScrapeRequest{Id: in.Id, Url: e.Attr("href"), Domain: in.Domain, Filter: in.Filter})
		}
	})

	ScrapeDetail(c)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(fmt.Sprintf("%s", in.Url))
	// Wait until threads are finished
	c.Wait()
}

var received = make(chan *pb.ScrapeRequest, 10000)

//Provide order to the system and limit amount of connections per crawler
//Think about using the leaky bucket, and a worker pool
//https://gobyexample.com/worker-pools
//& a bursty limiter
//https://gobyexample.com/rate-limiting
//TODO: Could also optimize memory allocation of collys with QueryId
func dispatch() {
FOLLOW:
	var in = <-received
	//Write as fast as we want to cassandra
	fmt.Printf("%s %s %s %s\n", in.Id, in.Url, in.Domain, in.Filter)
	//TODO: Check if we've already got it in cassandra - else scrape!!!!
	//time.Sleep(1 * time.Second)
	// c := Cassandra{}
	// c.Description()

	//TODO: Rate limit the scrape by ScrapeRequest.Id
	go scrape(in)
	goto FOLLOW //TODO: Test whether this as a single thread is ok

}

func main() {
	go dispatch()

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
