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

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"golang.org/x/sync/errgroup"

	"github.com/signalfx/splunk-otel-go/distro"
	"github.com/signalfx/splunk-otel-go/instrumentation/net/http/splunkhttp"
)

func main() {
	exitCode := 0
	defer func() {
		os.Exit(exitCode)
	}()

	// handle CTRL+C gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// initialize Splunk OTel distro
	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			log.Println(err)
			exitCode = 1
		}
	}()

	router := mux.NewRouter()

	// instrument http.Handler
	router.Use(otelmux.Middleware("mux-server"))
	router.Use(func(h http.Handler) http.Handler { return splunkhttp.NewHandler(h) })

	router.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println() // add a blank line between requests
		fmt.Println("HTTP request:")
		dump, err := httputil.DumpRequest(req, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Print(string(dump))
	}).Methods("GET")

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		WriteTimeout: time.Second,
		ReadTimeout:  time.Second,
	}
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	g.Go(func() error {
		// instrument http.Client
		client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(10 * time.Second):
				call(ctx, client)
			}
		}
	})

	<-ctx.Done() // wait for CTRL+C or server crash
	stop()       // stop receiving signal notifications; next interrupt signal should kill the application

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Println(err)
		exitCode = 1
		return
	}
	if err := g.Wait(); err != nil {
		log.Println(err)
		exitCode = 1
		return
	}
}

func call(ctx context.Context, client *http.Client) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/hello", http.NoBody)
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
}
