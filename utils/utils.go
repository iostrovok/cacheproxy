package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/iostrovok/cacheproxy/sqlite"
)

func CleanDirByTime(path string, timestamp int) (map[string]int64, error) {

	out := map[string]int64{}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".db") {
			return nil
		}

		conn, err := sqlite.Conn(path)
		if err != nil {
			return err
		}

		count, err := conn.DeleteOldByTime(timestamp)
		if err != nil {
			return err
		}

		out[path] = count

		return conn.Close()
	})

	return out, err
}
