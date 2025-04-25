package gcs_test

import (
	"os"
	"testing"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

const (
	gcsTestHost = "127.0.0.1"
	gcsTestPort = 8081
)

var (
	fakeServer *fakestorage.Server
)

func TestMain(m *testing.M) {
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		Host:   gcsTestHost,
		Port:   gcsTestPort,
		Scheme: "http",
	})
	if err != nil {
		panic(err)
	}

	fakeServer = server
	exitCode := m.Run()

	fakeServer.Stop()
	os.Exit(exitCode)
}
