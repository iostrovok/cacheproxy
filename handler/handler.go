package handler

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"github.com/iostrovok/cacheproxy/config"
	"github.com/iostrovok/cacheproxy/store"
)

var re = regexp.MustCompile(`[^-_a-zA-Z0-9]+`)

func handler(cfg *config.Config, w http.ResponseWriter, req *http.Request) {
	err := finger(cfg, w, req)
	if err != nil {
		logError(cfg, err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
}

func logRequestPrintf(cfg *config.Config, forceSave bool, urlStr string, requestDump []byte) {
	if !cfg.Verbose {
		return
	}

	debugStr := string(requestDump)
	if len(debugStr) > 100 {
		debugStr = debugStr[:100]
	}

	logPrintf(cfg, "[ForceSave: %t] Try to get %s Request: [%s]", forceSave, urlStr, debugStr)
}

func finger(cfg *config.Config, w http.ResponseWriter, req *http.Request) error {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return err
	}

	key, err := cacheKey(cfg, req, requestDump)
	if err != nil {
		return err
	}

	req.URL.Host = cfg.URL.Host
	req.URL.Scheme = cfg.URL.Scheme
	urlStr := req.URL.String()

	logRequestPrintf(cfg, cfg.ForceSave, urlStr, requestDump)

	fileName := fileKey(cfg, urlAsString(req.URL, cfg.NoUseDomain, cfg.NoUseUserData))
	if !cfg.ForceSave {
		cfg.Logger.Printf("read file: %s, key: %s", fileName, key)
		body, err := cfg.Keeper.Read(fileName, key)
		if err != nil {
			return err
		}

		// it means value is found in cache
		if body != nil && len(body) > 0 {
			item := &store.Item{}
			if item, err = store.FromZip(body); err == nil {
				logPrintf(cfg, "Found at cache key: %s for %s", key, urlStr)
				copyHeader(w.Header(), item.ResponseHeader)
				w.WriteHeader(item.StatusCode)
				if _, err = io.Copy(w, bytes.NewReader(item.ResponseBody)); err != nil {
					if !cfg.Verbose { // always save errors
						log.Print(err)
					}
				}
			}

			return err
		}

		logPrintf(cfg, "NOT Found at cache key: %s for %s", key, urlStr)
	}

	logPrintf(cfg, "Loading from remote server.... %s", urlStr)

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// >>>>>>>>> store for next using
	storeData := &store.Item{
		Request:        requestDump,
		ResponseBody:   streamToByte(resp.Body),
		ResponseHeader: resp.Header,
		StatusCode:     resp.StatusCode,
	}

	body, err := storeData.ToZip()
	if err != nil {
		return err
	}

	cfg.Logger.Printf("save file: %s, key: %s", fileName, key)
	if err := cfg.Keeper.Save(fileName, key, body); err != nil {
		return err
	}
	// <<<<<<<<<< store for next using

	logPrintf(cfg, "Result of loading: StatusCode: %d, Response Length: %d",
		storeData.StatusCode, len(storeData.ResponseBody))

	// return result
	copyHeader(w.Header(), storeData.ResponseHeader)
	w.WriteHeader(storeData.StatusCode)
	w.Write(storeData.ResponseBody)

	return nil
}

func cloneUrl(in *url.URL) *url.URL {
	var user *url.Userinfo
	if in.User != nil {
		user = &url.Userinfo{}
		if p, find := in.User.Password(); find {
			user = url.UserPassword(in.User.Username(), p)
		} else {
			user = url.User(in.User.Username())
		}
	}

	return &url.URL{
		Scheme:      in.Scheme,
		Opaque:      in.Opaque,
		User:        user,
		Host:        in.Host,
		Path:        in.Path,
		RawPath:     in.RawPath,
		ForceQuery:  in.ForceQuery,
		RawQuery:    in.RawQuery,
		Fragment:    in.Fragment,
		RawFragment: in.RawFragment,
	}
}

func urlAsString(in *url.URL, noUseDomain, noUseUserData bool) string {
	u := cloneUrl(in)
	u.Scheme = ""
	if noUseDomain {
		u.Host = ""
	}
	if noUseUserData {
		u.User = nil
	}

	return strings.TrimLeft(u.String(), "/")
}

func cacheKey(cfg *config.Config, req *http.Request, dump []byte) (string, error) {
	b := []byte(urlAsString(req.URL, cfg.NoUseDomain, cfg.NoUseUserData))

	bodyParts := bytes.SplitN(dump, []byte("\r\n\r\n"), 2)
	if len(bodyParts) == 2 {
		b = append(b, bodyParts[1]...)
	}

	// convert key to human-readable value
	return fmt.Sprintf("%x", md5.Sum(b)), nil
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

func fileKey(cfg *config.Config, urlPath string) string {
	if !cfg.DynamoFileName && cfg.FileName != "" {
		return cfg.FileName
	}

	return re.ReplaceAllString(urlPath, "")
}
