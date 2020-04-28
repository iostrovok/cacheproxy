package cacheproxy

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
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

var manager *Manager
var manager2 *Manager
var testHome string

//Run once when the suite starts running.
func (s *testSuite) SetUpSuite(c *C) {
	s.globalCtx, s.globalCancel = context.WithCancel(context.Background())
	cfg := baseCfg()
	cfg.Port = StandaloneServerPort
	err := Server(s.globalCtx, cfg)
	c.Assert(err, IsNil)
}

//Run before each test or benchmark starts running.
func (s *testSuite) SetUpTest(c *C) {}

//Run after each test or benchmark runs.
func (s *testSuite) TearDownTest(c *C) {}

//Run once after all tests or benchmarks have finished running.
func (s *testSuite) TearDownSuite(c *C) {
	s.globalCancel()
}

func init() {
	testHome = os.Getenv("TEST_SOURCE_PATH")
	if testHome == "" {
		log.Fatal("Please setup the TEST_SOURCE_PATH")
	}

	manager = NewManager(18801, 18899, baseCfg())
	manager2 = NewManager(19001, 19001, baseCfg())
}

func baseCfg() *config.Config {
	return &config.Config{
		Host:      "http://127.0.0.1:9200", // local elastisearch
		Scheme:    "http",                  // http OR https
		StorePath: filepath.Join(testHome, "cassettes"),
		Verbose:   true,
		ForceSave: false,
	}
}

func (s *testSuite) TestGet(c *C) {

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

func (s *testSuite) TestPost(c *C) {
	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		fmt.Fprintln(w, "Hello, client - TestPost")
	}))
	defer ts.Close()

	cfg := baseCfg()
	cfg.Host = ts.URL
	cfg.Port = 19200
	cfg.FileName = "my_test_2.db"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler.Start(ctx, cfg)
	c.Assert(ctx, NotNil)

	buf := []byte(`{"from":0,"query":{"bool":{"must":[{"match":{"brands":"Abita Amber"}}]}},"size":25}`)

	resp, err := http.Post("http://localhost:19200/beer/_search", "application/json", bytes.NewReader(buf))
	c.Assert(err, IsNil)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(string(body), DeepEquals, "Hello, client - TestPost\n")
	c.Assert(counter, Equals, 1)

	resp2, err2 := http.Post("http://localhost:19200/beer/_search", "application/json", bytes.NewReader(buf))
	c.Assert(err2, IsNil)
	defer resp2.Body.Close()
	body2, err2 := ioutil.ReadAll(resp2.Body)
	c.Assert(err2, IsNil)
	c.Assert(string(body2), DeepEquals, "Hello, client - TestPost\n")
	c.Assert(counter, Equals, 1)
}

func (s *testSuite) TestWithManager(c *C) {
	c.Skip("TestPost")

	wg := sync.WaitGroup{}
	for _, fileName := range []string{"test_0.db", "test_1.db", "test_2.db", "test_3.db", "test_0.db", "test_1.db", "test_2.db", "test_3.db"} {
		wg.Add(1)
		go func(c *C, fileName string) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())
			port, err := manager.RunSrv(ctx, fileName)
			c.Assert(port, NotNil)
			defer cancel()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/beer/beer/OQ3ur2wBHqN_LZyul2Oh", port))
			c.Assert(err, IsNil)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			c.Assert(err, IsNil)
			c.Assert(len(body), Equals, 710)

			//c.Assert(stopFunc(), IsNil)

		}(c, fileName)
	}

	wg.Wait()
}

func (s *testSuite) TestWithManager2(c *C) {
	c.Skip("TestPost")

	wg := sync.WaitGroup{}
	for _, fileName := range []string{"test_10.db"} {
		wg.Add(1)
		go func(c *C, fileName string) {
			defer wg.Done()

			ctx, cancel := context.WithCancel(context.Background())
			port, err := manager2.RunSrv(ctx, fileName)
			c.Assert(err, IsNil)
			defer cancel()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/beer/beer/OQ3ur2wBHqN_LZyul2Oh", port))
			c.Assert(err, IsNil)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			c.Assert(err, IsNil)
			c.Assert(len(body), Equals, 710)
			//c.Assert(stopFunc(), IsNil)
		}(c, fileName)
	}

	wg.Wait()
}

func (s *testSuite) TestWithStandaloneServer(c *C) {
	c.Skip("TestPost")

	wg := sync.WaitGroup{}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(c *C) {
			defer wg.Done()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/beer/beer/Og1VtGwBHqN_LZyuhGNG", StandaloneServerPort))
			c.Assert(err, IsNil)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)

			c.Assert(err, IsNil)
			c.Assert(len(body), Equals, 1166)
		}(c)
	}

	wg.Wait()
}
