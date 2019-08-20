package cacheproxy

import (
	"bytes"
	"context"
	"fmt"
	. "gopkg.in/check.v1"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

var manager *Manager
var manager2 *Manager
var testHome string

func init() {
	testHome = os.Getenv("TEST_SOURCE_PATH")
	if testHome == "" {
		log.Fatal("Please setup the TEST_SOURCE_PATH")
	}

	manager = NewManager(18900, 18900, baseCfg())
	manager2 = NewManager(18900, 19000, baseCfg())
}

func baseCfg() *Config {
	return &Config{
		URL:       "127.0.0.1:9200", // local elastisearch
		Scheme:    "http",           // http OR https
		StorePath: filepath.Join(testHome, "cassettes"),
		Verbose:   true,
		ForceSave: false,
	}
}

func (s *testSuite) TestGet(c *C) {
	cfg := baseCfg()
	cfg.Port = 19201
	cfg.FileName = "my_test.zip"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, cfg)
	c.Assert(ctx, NotNil)

	resp, err := http.Get("http://127.0.0.1:19201/beer/beer/OQ3ur2wBHqN_LZyul2Oh")
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(len(body), Equals, 710)
}

func (s *testSuite) TestPost(c *C) {
	cfg := baseCfg()
	cfg.Port = 19200
	cfg.FileName = "my_test_2.zip"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, cfg)
	c.Assert(ctx, NotNil)

	buf := []byte(`{"from":0,"query":{"bool":{"must":[{"match":{"brands":"Abita Amber"}}]}},"size":25}`)

	resp, err := http.Post("http://localhost:19200/beer/_search", "application/json", bytes.NewReader(buf))
	c.Assert(err, IsNil)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)
	c.Assert(len(body), Equals, 814)
}

func (s *testSuite) TestWithManager(c *C) {
	wg := sync.WaitGroup{}
	for _, fileName := range []string{"test_0.zip", "test_1.zip", "test_2.zip", "test_3.zip", "test_0.zip", "test_1.zip", "test_2.zip", "test_3.zip"} {
		wg.Add(1)
		go func(c *C, fileName string) {
			defer wg.Done()

			//ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			port, cancel := manager.RunSrv(context.Background(), fileName)
			defer cancel()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/beer/beer/OQ3ur2wBHqN_LZyul2Oh", port))
			c.Assert(err, IsNil)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			c.Assert(err, IsNil)
			c.Assert(len(body), Equals, 710)

		}(c, fileName)
	}

	wg.Wait()
}

func (s *testSuite) TestWithManager2(c *C) {
	wg := sync.WaitGroup{}
	for _, fileName := range []string{"test_10.zip", "test_10.zip", "test_11.zip", "test_11.zip", "test_12.zip", "test_13.zip", "test_12.zip", "test_13.zip"} {
		wg.Add(1)
		go func(c *C, fileName string) {
			defer wg.Done()

			//ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			port, cancel := manager2.RunSrv(context.Background(), fileName)
			defer cancel()

			resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/beer/beer/OQ3ur2wBHqN_LZyul2Oh", port))
			c.Assert(err, IsNil)

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			c.Assert(err, IsNil)
			c.Assert(len(body), Equals, 710)

		}(c, fileName)
	}

	wg.Wait()
}
