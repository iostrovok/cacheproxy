package handler

import (
	"bytes"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/sqlite"
	"github.com/iostrovok/cacheproxy/store"
)

func init() {
	sqlite.Init()
}

func handler(cfg *config.Config, w http.ResponseWriter, req *http.Request) {

	key, requestDump, err := cacheKey(req)

	fmt.Printf("\nkey: %s,\nrequestDump: %s,\nerr: %+v\n\n\n", key, string(requestDump), err)
	if err != nil {
		logError(cfg, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	req.URL.Host = cfg.URL.Host
	req.URL.Scheme = cfg.URL.Scheme

	logPrintf(cfg, "Try to get %s", req.URL.String())

	fileName := cfg.File(req.URL.String())

	if !cfg.ForceSave {
		store, err := sqlite.Select(fileName, key)
		if err != nil && err != sql.ErrNoRows {
			logError(cfg, err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		// it means value is found
		if store != nil {
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
	storeData := &store.StoreUnit{
		Request:        requestDump,
		ResponseBody:   streamToByte(resp.Body),
		ResponseHeader: resp.Header,
		StatusCode:     resp.StatusCode,
	}

	if err := sqlite.Upsert(fileName, key, storeData); err != nil {
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

func cacheKey(req *http.Request) (string, []byte, error) {

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

func logError(cfg *config.Config, err error) {
	if cfg.Verbose && err != nil {
		log.Print(err)
	}
}

func logPrintf(cfg *config.Config, tmpl string, data ...interface{}) {
	if cfg.Verbose {
		log.Printf(tmpl, data...)
	}
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
