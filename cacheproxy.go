package cacheproxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"sync"
)

type store struct {
	Request        []byte      `json:"request"`
	ResponseBody   []byte      `json:"response_body"`
	ResponseHeader http.Header `json:"response_header"`
	StatusCode     int         `json:"status_code"`
}

// Exists reports whether the named file or directory exists.
func existsFile(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// readStoreFile reads zip file and returns store object
func readStoreFile(cfg *Config) (map[string]*store, error) {

	filename := cfg.File()
	out := map[string]*store{}

	if !existsFile(filename) {
		return out, nil
	}

	//file, err := os.Open(filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() < 10 {
		return out, nil
	}

	fileGZip, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer fileGZip.Close()

	content, err := ioutil.ReadAll(fileGZip)
	if err != nil {
		// if we can't read file... It's wrong file!
		return out, nil
	}

	if len(content) > 0 {
		err = json.Unmarshal(content, &out)
		if err != nil {
			logPrintf(cfg, "5. readStoreFile.... %d => %v => %v", cfg.Port, err, content)
		}
	}

	return out, err
}

func updateStoreFile(cfg *Config, key string, data *store) error {

	filename := cfg.File()

	out, err := readStoreFile(cfg)
	if err != nil {
		return err
	}

	out[key] = data

	rawData, err := json.Marshal(out)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	zw, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		return err
	}

	if _, err := zw.Write(rawData); err != nil {
		return err
	}

	return zw.Close()
}

func findKey(cfg *Config, key string) (*store, error, bool) {

	data, err := readStoreFile(cfg)
	if err != nil {
		return nil, err, false
	}

	body, find := data[key]
	return body, nil, find
}

func cacheKey(cfg *Config, req *http.Request) (string, []byte, error) {

	b, err := req.URL.MarshalBinary()
	if err != nil {
		return "", nil, err
	}

	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "", nil, err
	}

	bodyParts := bytes.SplitN(dump, []byte("\r\n\r\n"), 2)
	if len(bodyParts) == 2 {
		b = append(b, bodyParts[1]...)
	}

	// convert key to human readable value
	return fmt.Sprintf("%x", md5.Sum(b)), dump, nil
}

func logError(cfg *Config, err error) {
	if cfg.Verbose && err != nil {
		log.Print(err)
	}
}

func logPrintf(cfg *Config, tmpl string, data ...interface{}) {
	if cfg.Verbose {
		log.Printf(tmpl, data...)
	}
}

func handler(cfg *Config, w http.ResponseWriter, req *http.Request) {

	locker := blocker.FileLocker(cfg.File())
	defer locker.Unlock()

	key, requestDump, err := cacheKey(cfg, req)
	if err != nil {
		logError(cfg, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	req.URL.Host = cfg.URL
	req.URL.Scheme = cfg.Scheme

	logPrintf(cfg, "Try to get %s", req.URL.String())

	if !cfg.ForceSave {
		store, err, find := findKey(cfg, key)
		if err != nil {
			logError(cfg, err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		if find {
			logPrintf(cfg, "Found at cache key: %s for %s", key, req.URL.String())
			copyHeader(w.Header(), store.ResponseHeader)
			w.WriteHeader(store.StatusCode)
			_, err := io.Copy(w, bytes.NewReader(store.ResponseBody))
			if err != nil {
				log.Print(err)
			}
			return
		}
	}

	logPrintf(cfg, "Loading from remote server.... %s", req.URL.String())

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		logError(cfg, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// >>>>>>>>> store for next using
	storeData := &store{
		Request:        requestDump,
		ResponseBody:   streamToByte(resp.Body),
		ResponseHeader: resp.Header,
		StatusCode:     resp.StatusCode,
	}

	if err := updateStoreFile(cfg, key, storeData); err != nil {
		logError(cfg, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	// <<<<<<<<<< store for next using

	// return result
	copyHeader(w.Header(), storeData.ResponseHeader)
	w.WriteHeader(storeData.StatusCode)
	w.Write(storeData.ResponseBody)
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func run(ctx context.Context, cfg *Config) {

	locker := blocker.PortLocker(cfg.Port)
	go func(ctx context.Context, cfg *Config, locker *sync.RWMutex) {
		defer locker.Unlock()

		server := &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.Port),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler(cfg, w, r)
			}),
		}

		ch := make(chan error, 1)
		if cfg.Scheme == "https" {
			select {
			case <-ctx.Done():
				// nothing
			case ch <- server.ListenAndServeTLS(cfg.PemPath, cfg.KeyPath):
				// nothing
			}
		} else {
			select {
			case <-ctx.Done():
				// nothing
			case ch <- server.ListenAndServe():
				// nothing
			}
		}

		logError(cfg, <-ch)
		logPrintf(cfg, "Force close server. Port: %d, error: %v", cfg.Port, server.Close())
	}(ctx, cfg, locker)

}
