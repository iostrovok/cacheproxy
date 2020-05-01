package config

import (
	"testing"

	. "github.com/iostrovok/check"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

// test syntax only
func (s *testSuite) Test(c *C) {

	cfg := Config{
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
	c.Assert(cfg.File("http://127.0.0.1:20200/home.php"), Equals, "/tmp/my-project/http12700120200homephp.db")

	cfg.DynamoFileName = false
	c.Assert(cfg.File("http://127.0.0.1:20200/home.php"), Equals, "/tmp/my-project/my.db")

}
