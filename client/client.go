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

package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	pb "github.com/dioptre/gtscrp/proto"
	"github.com/fatih/color"
	// "github.com/pkg/errors"
	// "github.com/spf13/cobra"
	// "github.com/spf13/viper"
	// "github.com/tj/go-gracefully"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	address       = "frontend.local:4443"
	defaultURL    = "https://httpbin.org/delay/2"
	defaultFilter = "*httpbin.*"
)

func main() {
	FrontendCert, _ := ioutil.ReadFile("./frontend.cert")

	// Create CertPool
	roots := x509.NewCertPool()
	roots.AppendCertsFromPEM(FrontendCert)

	// Create credentials
	credsClient := credentials.NewClientTLSFromCert(roots, "")

	// Dial with specific Transport (with credentials)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(credsClient))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewScraperClient(conn)

	// Contact the server and print out its response.
	url := defaultURL
	if len(os.Args) > 1 {
		url = os.Args[1]
	}
	filter := defaultFilter
	if len(os.Args) > 2 {
		filter = os.Args[2]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := client.Scrape(ctx, &pb.ScrapeRequest{Url: url, Filter: filter})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(color.New(color.Bold, color.FgHiBlack).SprintfFunc()("Scraper start ack: %t", r.Message))
}
