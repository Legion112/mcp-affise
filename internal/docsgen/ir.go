package docsgen

// Operation is the intermediate representation of one Affise API endpoint
// scraped from HTML docs.
type Operation struct {
	Method     string
	Path       string
	Summary    string
	Parameters []Parameter
}

// Parameter describes a query or path parameter from the docs table.
type Parameter struct {
	Name        string
	In          string // "query" or "path"
	Type        string // Affise type string, e.g. "String", "Array[integer]"
	Description string
	Default     string
	Enum        []string
	Required    bool
}
