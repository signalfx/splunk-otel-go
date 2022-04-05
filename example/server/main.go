package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/signalfx/splunk-otel-go/distro"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
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
			panic(err)
		}
	}()

	r := mux.NewRouter()

	// instrument gorilla/mux
	r.Use(otelmux.Middleware("mux-server"))

	// instrument http.Handler
	otelHandler := otelhttp.NewHandler(r, "http-server")

	r.HandleFunc("/hello", func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "Hello there.")
	}).Methods("GET")

	srv := &http.Server{
		Addr:    ":8080",
		Handler: otelHandler,
	}
	srvErrCh := make(chan error)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			srvErrCh <- err
		} else {
			srvErrCh <- nil
		}
	}()

	<-ctx.Done()
	stop() // stop receiving signal notifications; next interrupt signal should kill the application

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalln(err)
	}
	if err := <-srvErrCh; err != nil {
		log.Fatalln(err)
	}
}
