package sqlite

import (
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

// test syntax only
func (s *testSuite) Test_fullFileName(c *C) {
	sq := &Sqlite{
		storePath: "/tmp/my-test",
	}

	c.Assert(sq.fullFileName(""), EqualsMore, "/tmp/my-test/.db")
	c.Assert(sq.fullFileName(" "), EqualsMore, "/tmp/my-test/ .db")
	c.Assert(sq.fullFileName("  "), EqualsMore, "/tmp/my-test/  .db")
	c.Assert(sq.fullFileName("123"), EqualsMore, "/tmp/my-test/123.db")
	c.Assert(sq.fullFileName("my-super-file"), EqualsMore, "/tmp/my-test/my-super-file.db")
}
