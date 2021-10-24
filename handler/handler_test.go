package handler

import (
	"github.com/iostrovok/cacheproxy/config"
	"net/url"
	"testing"

	. "github.com/iostrovok/check"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

// test syntax only
func (s *testSuite) Test(c *C) {
	c.Assert(true, Equals, true)
}

func (s *testSuite) Test_UrlAsString(c *C) {
	u, err := url.Parse("https://username:password@bing.com/search?q=dotnet")
	c.Assert(err, IsNil)

	c.Assert(urlAsString(u, true, true), EqualsMore, "search?q=dotnet")
	c.Assert(urlAsString(u, false, true), EqualsMore, "bing.com/search?q=dotnet")
	c.Assert(urlAsString(u, true, false), EqualsMore, "username:password@/search?q=dotnet")
	c.Assert(urlAsString(u, false, false), EqualsMore, "username:password@bing.com/search?q=dotnet")
}

func (s *testSuite) Test_File(c *C) {
	cfg := &config.Config{
		Host:   "http://localhost:9200",
		Scheme: "http",
		Port:   20200,
		// PemPath, KeyPath
		StorePath:      "/tmp/my-project",
		FileName:       "my.db",
		Verbose:        true,
		ForceSave:      false,
		DynamoFileName: true,
		SessionMode:    true,
	}

	c.Assert(cfg.Init(), IsNil)

	c.Assert(cfg.URL.Host, Equals, "localhost:9200")
	c.Assert(fileKey(cfg, "http://127.0.0.1:20200/home.php"), Equals, "http12700120200homephp")

	cfg.DynamoFileName = false
	c.Assert(fileKey(cfg, "http://127.0.0.1:20200/home.php"), Equals, "my.db")
}
