package store

import (
	"net/http"
	"testing"

	. "github.com/iostrovok/check"
)

type testSuite struct{}

var _ = Suite(&testSuite{})

func TestService(t *testing.T) { TestingT(t) }

func (s *testSuite) TestToZip(c *C) {

	unit := &Item{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
		StatusCode: 200,
	}

	zip, err := unit.ToZip()

	c.Assert(err, IsNil)
	c.Assert(len(zip), Equals, 105)
}

func (s *testSuite) TestFromZip_Error(c *C) {

	_, err := FromZip([]byte{})
	c.Assert(err, NotNil)

	_, err = FromZip([]byte{12, 34, 56})
	c.Assert(err, NotNil)
}

func (s *testSuite) TestFromZip(c *C) {

	in := &Item{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
		StatusCode: 200,
	}

	zip, err := in.ToZip()
	c.Assert(err, IsNil)

	out, err := FromZip(zip)
	c.Assert(err, IsNil)

	c.Assert(out.StatusCode, Equals, 200)
	c.Assert(out.Request, DeepEquals, []byte{100})
	c.Assert(out.ResponseBody, DeepEquals, []byte{101})
	c.Assert(out.ResponseHeader["HEADER-1"], DeepEquals, []string{"VALUE-1", "VALUE-2"})
}
