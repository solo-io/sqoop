package generate

type Dependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Repository string `json:"repository"`
}

type DependencyList struct {
	Dependencies []Dependency `json:"dependencies"`
}
