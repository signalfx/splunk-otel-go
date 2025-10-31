// Copyright Splunk Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package example instruments a simple HTTP server-client application.
//
// The application is configured to send spans to a local instance
// of the OpenTelemetry Collector, which propagates them to both
// Splunk Observability Cloud and to a local Jaeger instance.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"

	"github.com/signalfx/splunk-otel-go/distro"
	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
)

const address = "localhost:8080"

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() (err error) {
	// handle CTRL+C gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// initialize Splunk OTel distro
	sdk, err := distro.Run()
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, sdk.Shutdown(context.Background()))
	}()

	logger := slog.New(otelslog.NewHandler("github.com/signalfx/splunk-otel-go/example"))
	slog.SetDefault(logger)

	slog.InfoContext(ctx, "Application started", slog.String("address", address))

	// instrument http.Handler
	var handler http.Handler = http.HandlerFunc(handle)
	handler = splunkhttp.NewHandler(handler)
	handler = otelhttp.NewHandler(handler, "handle")

	l, err := (&net.ListenConfig{}).Listen(ctx, "tcp", address)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           handler,
		WriteTimeout:      time.Second,
		ReadTimeout:       time.Second,
		ReadHeaderTimeout: time.Second,
	}

	g := &errgroup.Group{}

	g.Go(func() error {
		err := srv.Serve(l) // Closing via srv.Shutdown.
		if err == http.ErrServerClosed {
			return nil
		}
		return err // Error while serving.
	})

	g.Go(func() error {
		// instrument http.Client
		client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

		call(ctx, client)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", http.NoBody)
		if err != nil {
			panic(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		resp.Body.Close()

		// When Shutdown is called, Serve immediately returns ErrServerClosed.
		return srv.Shutdown(context.Background())
	})

	return g.Wait()
}

func handle(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HTTP request:")
	dump, err := httputil.DumpRequest(req, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Print(string(dump))
}

func call(ctx context.Context, client *http.Client) {
	slog.InfoContext(ctx, "Making HTTP request", slog.String("url", "http://"+address))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+address, http.NoBody)
	if err != nil {
		panic(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("HTTP response:")
	dump, err := httputil.DumpResponse(resp, false)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Print(string(dump))
	slog.InfoContext(ctx, "HTTP request completed", slog.Int("status", resp.StatusCode))
}
