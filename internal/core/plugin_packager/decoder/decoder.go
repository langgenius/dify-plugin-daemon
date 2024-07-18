package decoder

type PluginDecoder interface {
	Open() error
	Walk(fn func(filename string, dir string) error) error
	ReadFile(filename string) ([]byte, error)
	Close() error
}
