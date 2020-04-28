package sqlite

import (
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/cacheproxy/store"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

func tmpFile(c *C) string {

	// Create our Temp File:  This will create a filename like /tmp/prefix-123456
	// We can use a pattern of "pre-*.txt" to get an extension like: /tmp/pre-123456.txt
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	c.Assert(err, IsNil)
	out := tmpFile.Name()
	c.Assert(tmpFile.Close(), IsNil)
	os.Remove(out)

	return out

}

// test syntax only
func (s *testSuite) TestSQL_Conn(c *C) {
	fileName := tmpFile(c)

	q, err := Conn(fileName)
	c.Assert(err, IsNil)
	c.Assert(q, NotNil)

	defer os.Remove(fileName)

	c.Assert(q.Close(), IsNil)

	q, err = Conn(fileName)
	c.Assert(err, IsNil)
	c.Assert(q, NotNil)

	q2, err2 := Conn(fileName)
	c.Assert(err2, IsNil)
	c.Assert(q2, NotNil)
}

// test syntax only
func (s *testSuite) TestSQL_1(c *C) {
	fileName := tmpFile(c)

	q, err := Conn(fileName)
	c.Assert(err, IsNil)
	c.Assert(q, NotNil)

	defer os.Remove(fileName)

	unit := &store.StoreUnit{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
	}

	c.Assert(q.Upsert("test-1", unit), IsNil)

	unit2, err2 := q.Select("test-1")
	c.Assert(err2, IsNil)

	c.Assert(unit2.Request, DeepEquals, unit.Request)
	c.Assert(unit2.ResponseBody, DeepEquals, unit.ResponseBody)
	c.Assert(unit2.ResponseHeader, DeepEquals, unit.ResponseHeader)
	c.Assert(unit2.StatusCode, DeepEquals, unit.StatusCode)

	c.Assert(q.Close(), IsNil)
}
