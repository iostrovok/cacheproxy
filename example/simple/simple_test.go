package simple

/*
	It is an example of tests of requests to elasticsearch (ES).
	We will keep requests to ES in /my-project/cassettes directory.
*/

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/cacheproxy"
	cacheproxyConfig "github.com/iostrovok/cacheproxy/config"
)

var elasticsearchURL = ""

type testSuite struct {
	globalCtx    context.Context
	globalCancel context.CancelFunc
}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

// Run once when the suite starts running.
func (s *testSuite) SetUpSuite(c *C) {
	s.globalCtx, s.globalCancel = context.WithCancel(context.Background())

	tmpPort := 19200
	schema := "http"

	elasticsearchUrl := "http://127.0.0.1:9200"
	URL, err := url.Parse(elasticsearchUrl)
	if err != nil {
		log.Fatal(err)
	}

	// prepare the config for cacheproxy
	cfg := &cacheproxyConfig.Config{
		Host:      elasticsearchUrl,
		Scheme:    schema,
		StorePath: "/my-project/cassettes", // Absolute path.
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
	elasticsearchURL = URL.String()
}

//Run before each test or benchmark starts running.
func (s *testSuite) SetUpTest(c *C) {
}

//Run after each test or benchmark runs.
func (s *testSuite) TearDownTest(c *C) {}

//Run once after all tests or benchmarks have finished running.
func (s *testSuite) TearDownSuite(c *C) {
	// shutdown the cacheproxy servers after tests
	s.globalCancel()
}

// Test uses Elasticsearch
func (s *testSuite) Test_First(c *C) {
	// create connection to ES and others
	resp, err := http.Get(elasticsearchURL + "/index/_search?q=*:*&track_total_hits=true&size=1")

	// check result
	c.Assert(err, IsNil)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	c.Assert(body, Not(DeepEquals), []string{"1", "2"})
}
