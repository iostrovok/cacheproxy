# CacheProxy [![Build Status](https://travis-ci.org/iostrovok/cacheproxy.svg?branch=master)](https://travis-ci.org/github/iostrovok/cacheproxy)

CacheProxy is a simple way to test over http offline.

#### Using

Example for test with Elasticsearch:
```go

package mypackage

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"testing"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/cacheproxy"
	"github.com/iostrovok/cacheproxy/config"
)

type testSuite struct {
	globalCtx    context.Context
	globalCancel context.CancelFunc
}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

// Run once when the suite starts running.
func (s *testSuite) SetUpSuite(c *C) {

	s.globalCtx, s.globalCancel = context.WithCancel(context.Background())

	/*
	    Out application uses "ELASTICSEARCH_URL" for reading elasticsearch url like
        ELASTICSEARCH_URL = "http://127.0.0.1:9200"
	    We keep request to ES in /my-project/cassettes directory.
	*/

	tmpPort := 19200
	schema := "http"

	cassettesDir := "/my-project/cassettes" // Absolute path. 
	elasticsearchUrl := os.Getenv("ELASTICSEARCH_URL")
	URL, err := url.Parse(elasticsearchUrl)
	if err != nil {
		log.Fatal(err)
	}

	cfg := &config.Config{
		Host:      elasticsearchUrl,
		Scheme:    schema,
		StorePath: cassettesDir,
		Verbose:   true,
		ForceSave: false,
		Port:      tmpPort,

        // This option provides deleting records which weren't requested during tests.
		SessionMode: true,
	}

	// start the cacheproxy servers
	err = cacheproxy.Server(s.globalCtx, cfg)
	c.Assert(err, IsNil)

	URL.Scheme = schema
	URL.Host = fmt.Sprintf("127.0.0.1:%d", tmpPort)
	if err := os.Setenv("ELASTICSEARCH_URL", URL.String()); err != nil {
		log.Fatal(err)
	}
}

//Run before each test or benchmark starts running.
func (s *testSuite) SetUpTest(c *C) {
}

//Run after each test or benchmark runs.
func (s *testSuite) TearDownTest(c *C) {}

//Run once after all tests or benchmarks have finished running.
func (s *testSuite) TearDownSuite(c *C) {
    // shotdown the cacheproxy servers
	s.globalCancel()
}

// Test uses Elasticsearch
func (s *testSuite) TestFirst(c *C) {
	// create connection to ES and others
	server := InitServer(c)

	// code which get data from ES
	data, err := server.GetDataFromES()

	// check result
	c.Assert(err, IsNil)
	c.Assert(data, DeepEquals, []string{"1", "2"})
}





```

