package core_http_server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunRejectsOccupiedAddress(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	defer listener.Close()

	server := newTestServer(t, &fakeReadinessChecker{
		ping: func(context.Context) error {
			return nil
		},
		timeout: time.Second,
	})
	server.config.Addr = listener.Addr().String()
	if err := server.Run(context.Background()); err == nil {
		t.Fatal("Run() accepted an occupied address")
	}
}

func TestServerWaitsForActiveRequestOnShutdown(t *testing.T) {
	server := newTestServer(t, &fakeReadinessChecker{
		ping: func(context.Context) error {
			return nil
		},
		timeout: time.Second,
	})
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	defer func() {
		select {
		case <-releaseRequest:
		default:
			close(releaseRequest)
		}
	}()
	server.mux.HandleFunc("GET /slow", func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		w.WriteHeader(http.StatusOK)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.serve(ctx, listener)
	}()

	requestDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String() + "/slow")
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			err = response.Body.Close()
		}
		requestDone <- err
	}()
	select {
	case <-requestStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not reach the server")
	}

	cancel()
	select {
	case err := <-serveDone:
		t.Fatalf("server stopped before active request finished: %v", err)
	case <-time.After(20 * time.Millisecond):
	}
	close(releaseRequest)
	select {
	case err := <-requestDone:
		if err != nil {
			t.Fatalf("active request failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("active request did not finish")
	}
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("server shutdown failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not stop")
	}
}
