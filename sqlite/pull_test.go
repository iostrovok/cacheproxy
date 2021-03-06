package sqlite

import (
	"net/http"
	"os"
	"sync"
	"time"

	. "github.com/iostrovok/check"

	"github.com/iostrovok/cacheproxy/store"
)

// test syntax only
func (s *testSuite) TestSQL_Pull_New(c *C) {
	fileName := tmpFile(c)
	defer os.Remove(fileName)

	p := New(false)
	q, err := p.Get(fileName)
	c.Assert(err, IsNil)
	c.Assert(q, NotNil)

	c.Assert(p.Close(), IsNil)
}

// test syntax only
func (s *testSuite) TestSQL_Pull_Add(c *C) {
	fileName := tmpFile(c)
	defer os.Remove(fileName)

	q, err := (New(false)).Add(fileName)
	c.Assert(err, IsNil)
	c.Assert(q, NotNil)
}

// test syntax only
func (s *testSuite) TestSQL_Pull_2(c *C) {
	fileName := tmpFile(c)
	defer os.Remove(fileName)

	unit := &store.Item{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
	}

	key := "TestSQL_Pull_2"

	p := New(false)
	p.Upsert(fileName, key, unit)

	c.Assert(p.Upsert(fileName, key, unit), IsNil)

	unit2, err2 := p.Select(fileName, key)
	c.Assert(err2, IsNil)

	c.Assert(unit2.Request, DeepEquals, unit.Request)
	c.Assert(unit2.ResponseBody, DeepEquals, unit.ResponseBody)
	c.Assert(unit2.ResponseHeader, DeepEquals, unit.ResponseHeader)
	c.Assert(unit2.StatusCode, DeepEquals, unit.StatusCode)

	c.Assert(p.Close(), IsNil)
}

// test syntax only
func (s *testSuite) TestSQL_Pull_Multi(c *C) {
	fileName := tmpFile(c)
	defer os.Remove(fileName)

	key := "TestSQL_Pull_2"
	wg := sync.WaitGroup{}

	p := New(false)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			raceSubTest(c, p, fileName, key)
		}()
	}
	wg.Wait()
	c.Assert(p.Close(), IsNil)
}

// test syntax only
func (s *testSuite) TestSQL_Pull_SsessionMode(c *C) {

	countFiles := 7
	files := make([]string, countFiles)
	for i := 0; i < countFiles; i++ {
		files[i] = tmpFile(c)
	}

	defer func() {
		for _, f := range files {
			os.Remove(f)
		}
	}()
	keys := []string{"TestSQL_Pull_Global-1", "TestSQL_Pull_Global-2", "TestSQL_Pull_Global-3", "TestSQL_Pull_Global-4", "TestSQL_Pull_Global-5"}
	keys2 := []string{"2-TestSQL_Pull_Global-1", "2-TestSQL_Pull_Global-2", "2-TestSQL_Pull_Global-3",
		"2-TestSQL_Pull_Global-4", "2-TestSQL_Pull_Global-5", "2-TestSQL_Pull_Global-6"}

	wg := sync.WaitGroup{}
	p := New(false)

	for i := 0; i < 6*countFiles; i++ {
		wg.Add(1)
		go func(file, key string) {
			defer wg.Done()
			raceSubTest(c, p, file, key)
		}(files[i%len(files)], keys[i%len(keys)])
	}
	wg.Wait()

	c.Assert(p.Close(), IsNil)
	time.Sleep(1 * time.Second)

	p = New(true)

	wg = sync.WaitGroup{}
	for i := 0; i < 6*countFiles; i++ {
		wg.Add(1)
		go func(file, key string) {
			defer wg.Done()
			raceSubTest(c, p, file, key)
		}(files[i%len(files)], keys2[i%len(keys)])
	}
	wg.Wait()

	count, err := p.DeleteOld()
	c.Assert(err, IsNil)
	c.Assert(int(count), Equals, countFiles*len(keys))

	c.Assert(p.Close(), IsNil)
}

func raceSubTest(c *C, p *Pull, fileName, key string) {
	unit := &store.Item{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
	}

	for i := 0; i < 100; i++ {
		p.Upsert(fileName, key, unit)

		c.Assert(p.Upsert(fileName, key, unit), IsNil)

		unit2, err2 := p.Select(fileName, key)
		c.Assert(err2, IsNil)

		c.Assert(unit2.Request, DeepEquals, unit.Request)
		c.Assert(unit2.ResponseBody, DeepEquals, unit.ResponseBody)
		c.Assert(unit2.ResponseHeader, DeepEquals, unit.ResponseHeader)
		c.Assert(unit2.StatusCode, DeepEquals, unit.StatusCode)
	}
}

// test syntax only
func (s *testSuite) TestSQL_Pull_Global(c *C) {
	fileName1 := tmpFile(c)
	defer os.Remove(fileName1)

	fileName2 := tmpFile(c)
	defer os.Remove(fileName2)

	keys := []string{"TestSQL_Pull_Global-1", "TestSQL_Pull_Global-2", "TestSQL_Pull_Global-3"}
	files := []string{fileName1, fileName2}
	wg := sync.WaitGroup{}

	// make global object
	Init(false)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(file, key string) {
			defer wg.Done()
			globalRaceSubTest(c, file, key)
		}(files[i%len(files)], keys[i%len(keys)])
	}
	wg.Wait()
	// global closing
	c.Assert(Close(), IsNil)
}

func globalRaceSubTest(c *C, fileName, key string) {
	unit := &store.Item{
		Request:      []byte{100},
		ResponseBody: []byte{101},
		ResponseHeader: http.Header{
			"HEADER-1": []string{"VALUE-1", "VALUE-2"},
		},
	}

	for i := 0; i < 100; i++ {
		c.Assert(Upsert(fileName, key, unit), IsNil)

		unit2, err2 := Select(fileName, key)
		c.Assert(err2, IsNil)

		c.Assert(unit2.Request, DeepEquals, unit.Request)
		c.Assert(unit2.ResponseBody, DeepEquals, unit.ResponseBody)
		c.Assert(unit2.ResponseHeader, DeepEquals, unit.ResponseHeader)
		c.Assert(unit2.StatusCode, DeepEquals, unit.StatusCode)
	}
}
