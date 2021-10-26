package cacheproxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	dir, err := ioutil.TempDir(os.TempDir(), "prefix")
	c.Assert(err, IsNil)
	testHome = dir
	s.globalCtx, s.globalCancel = context.WithCancel(context.Background())
}

//Run before each test or benchmark starts running.
func (s *testSuite) SetUpTest(c *C) {}

//Run after each test or benchmark runs.
func (s *testSuite) TearDownTest(c *C) {}

// Run once after all tests or benchmarks have finished running.
func (s *testSuite) TearDownSuite(c *C) {
	s.globalCancel()
	tmpFiles := []string{"my_post_test.db", "my_get_test.db", "test_0.db", "test_1.db", "test_2.db",
		"test_3.db", "test_0.db", "test_1.db", "test_2.db", "test_3.db"}
	for _, fileName := range tmpFiles {
		os.RemoveAll(filepath.Join(testHome, fileName))
	}
}

func baseCfg(url, fileName string, port int) *config.Config {
	return &config.Config{
		Scheme:    "http", // http OR https
		StorePath: testHome,
		Verbose:   true,
		ForceSave: false,
		Host:      url, // like "http://127.0.0.1:9200"
		Port:      port,
		FileName:  fileName,
	}
}

func (s *testSuite) TestGet(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		fmt.Fprintln(w, "Hello, client - TestGet")
	}))
	defer ts.Close()

	cfg := baseCfg(ts.URL, "my_get_test.db", 19201)

	handler.Start(ctx, cfg)
	c.Assert(ctx, NotNil)

	// first request
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://127.0.0.1:19201/beer/beer/OQ3ur2wBHqN_LZyul2Oh")
		c.Assert(err, IsNil)
		body, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, IsNil)
		c.Assert(body, EqualsMore, "Hello, client - TestGet\n")
		c.Assert(counter, Equals, 1)
		resp.Body.Close()
	}

	c.Assert(counter, Equals, 1)
}

func (s *testSuite) TestPost(c *C) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		fmt.Fprintln(w, "Hello, client - TestPost")
	}))
	defer ts.Close()

	cfg := baseCfg(ts.URL, "my_post_test.db", 19200)

	handler.Start(ctx, cfg)
	c.Assert(ctx, NotNil)

	buf := []byte(`{"from":0,"query":{"bool":{"must":[{"match":{"brands":"Abita Amber"}}]}},"size":25}`)

	for i := 0; i < 10; i++ {
		resp, err := http.Post("http://localhost:19200/beer/_search", "application/json", bytes.NewReader(buf))
		c.Assert(err, IsNil)
		body, err := ioutil.ReadAll(resp.Body)
		c.Assert(err, IsNil)
		c.Assert(body, EqualsMore, "Hello, client - TestPost\n")
		resp.Body.Close()
	}

	c.Assert(counter, Equals, 1)
}
