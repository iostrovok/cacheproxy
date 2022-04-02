package plugins

type IPlugin interface {
	// Read reads date from storage
	Read(file, key string) ([]byte, error)

	// Save saves date from storage
	Save(file, key string, data []byte) error

	// SetVersion sets the version of the data being used. The GIT branch name is the first candidate.
	SetVersion(version string) error
}

// ILogger is simple interface to output filename and key.
type ILogger interface {
	// Printf prints the filename and key
	Printf(format string, v ...interface{})
}
