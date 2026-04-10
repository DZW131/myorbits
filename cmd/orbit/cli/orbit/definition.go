package orbit

// Definition is a single orbit file definition loaded from .orbit/orbits.
type Definition struct {
	ID          string   `yaml:"id"`
	Description string   `yaml:"description"`
	Include     []string `yaml:"include"`
	Exclude     []string `yaml:"exclude"`
	SourcePath  string   `yaml:"-"`
}

// RepositoryConfig groups the loaded global config and orbit definitions.
type RepositoryConfig struct {
	Global GlobalConfig
	Orbits []Definition
}

// OrbitByID returns the orbit definition for the requested identifier.
func (config RepositoryConfig) OrbitByID(id string) (Definition, bool) {
	for _, definition := range config.Orbits {
		if definition.ID == id {
			return definition, true
		}
	}

	return Definition{}, false
}
