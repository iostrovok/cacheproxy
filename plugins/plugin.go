package plugins

type IPlugin interface {
	Read(file, key string) ([]byte, error)
	Save(file, key string, data []byte) error
	SetVersion(version string) error
}
