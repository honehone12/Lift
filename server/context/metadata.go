package context

type Metadata struct {
	name    string
	version string
}

func NewMetadata(name string, version string) *Metadata {
	return &Metadata{
		name:    name,
		version: version,
	}
}

func (m *Metadata) Name() string {
	return m.name
}

func (m *Metadata) Version() string {
	return m.version
}
