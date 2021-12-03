package main_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"

	"github.com/ingridhq/comptest"
	"github.com/ingridhq/comptest/_example/binary/config"
)

var cfg config.Config

func TestMain(t *testing.M) {
	if os.Getenv("RUN_COMPONENT_TESTS") != "true" {
		return
	}

	envconfig.MustProcess("", &cfg)

	cleanUp := comptest.MustBuildAndRun("./main.go", "./main.bin", "./comptest.logs")
	defer cleanUp()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := comptest.WaitForMainServer(ctx); err != nil {
		cleanUp()
		log.Fatalf("wait for main server failed: %v", err)
	}

	t.Run()
}

func Test_response(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%v/", cfg.Port))
	if err != nil {
		t.Fatalf("could not do get request: %v", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	if string(body) != "Potato" {
		t.Fatalf("unexpected response: %v", body)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code : %v", resp.StatusCode)
	}
}
