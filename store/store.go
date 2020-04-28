package store

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"net/http"
)

type StoreUnit struct {
	Request        []byte      `json:"request"`
	ResponseBody   []byte      `json:"response_body"`
	ResponseHeader http.Header `json:"response_header"`
	StatusCode     int         `json:"status_code"`
}

func (s *StoreUnit) ToZip() ([]byte, error) {

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

func FromZip(body []byte) (*StoreUnit, error) {
	b := bytes.Buffer{}
	r, err := zlib.NewReader(bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	r.Close()

	s := &StoreUnit{}
	err = json.Unmarshal(b.Bytes(), &s)
	return s, err
}
