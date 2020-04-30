package cacheproxy

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/handler"
)

var StandaloneServerPort = 35000

type testSuite struct {
	globalCtx    context.Context
	globalCancel context.CancelFunc
}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

var testHome string

//Run once when the suite starts running.
func (s *testSuite) SetUpSuite(c *C) {
	testHome = os.TempDir()
	s.globalCtx, s.globalCancel = context.WithCancel(context.Background())
	//cfg := baseCfg()
	//cfg.Port = StandaloneServerPort
	//err := Server(s.globalCtx, cfg)
	//c.Assert(err, IsNil)
}

//Run before each test or benchmark starts running.
func (s *testSuite) SetUpTest(c *C) {}

//Run after each test or benchmark runs.
func (s *testSuite) TearDownTest(c *C) {}

// Run once after all tests or benchmarks have finished running.
func (s *testSuite) TearDownSuite(c *C) {
	s.globalCancel()
}

func baseCfg() *config.Config {
	return &config.Config{
		// host will be set up later
		//Host:      "http://127.0.0.1:9200", // local elastisearch

		Scheme:    "http", // http OR https
		StorePath: testHome,
		Verbose:   true,
		ForceSave: false,
	}
}

func (s *testSuite) TestGet(c *C) {
	//c.Skip("TestPost")

	counter := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		fmt.Fprintln(w, "Hello, client - TestGet")
	}))
	defer ts.Close()

	cfg := baseCfg()
	cfg.Host = ts.URL
	cfg.Port = 19201
	cfg.FileName = "my_test.db"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler.Start(ctx, cfg)
	c.Assert(ctx, NotNil)

	// first request
	resp, err := http.Get("http://127.0.0.1:19201/beer/beer/OQ3ur2wBHqN_LZyul2Oh")
	c.Assert(err, IsNil)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), DeepEquals, "Hello, client - TestGet\n")
	c.Assert(counter, Equals, 1)

	// second request
	resp2, err2 := http.Get("http://127.0.0.1:19201/beer/beer/OQ3ur2wBHqN_LZyul2Oh")
	c.Assert(err2, IsNil)
	defer resp2.Body.Close()
	body2, err2 := ioutil.ReadAll(resp2.Body)
	c.Assert(err2, IsNil)
	c.Assert(string(body2), DeepEquals, "Hello, client - TestGet\n")
	c.Assert(counter, Equals, 1)
}
