package store

import (
	"bytes"
	"compress/zlib"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Item struct {
	// Request as is
	Request []byte `json:"request"`

	// Response Body as is
	ResponseBody []byte `json:"response_body"`

	// Response headers Body as is
	ResponseHeader http.Header `json:"response_header"`

	// status code
	StatusCode int `json:"status_code"`

	// for compare and debug goals
	// it's not stored in files
	Hash string `json:"-"`
}

func (s *Item) ToZip() ([]byte, error) {
	body, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	w := zlib.NewWriter(&b)
	if _, err := w.Write(body); err != nil {
		return nil, err
	}
	w.Close()

	return b.Bytes(), nil
}

func FromZip(body []byte, needHash ...bool) (*Item, error) {
	b := bytes.Buffer{}
	r, err := zlib.NewReader(bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	r.Close()

	s := &Item{}
	err = json.Unmarshal(b.Bytes(), &s)

	if len(needHash) > 0 && needHash[0] {
		s.Hash = fmt.Sprintf("%x", md5.Sum(body))
	}

	return s, err
}
