package main

// Document represents a parsed markdown document
type Document struct {
	FilePath    string
	Frontmatter map[string]interface{}
	Sections    []Section
}

// Section represents a markdown section (heading + content)
type Section struct {
	Level   int    // 1 for h1, 2 for h2, etc.
	Title   string // Title text without the # symbols
	Heading string // The full heading line including #
	Body    string // Content until next section of same or higher level
	Index   int    // Index among sections of the same level
}

// QueryResult represents the result of a query
type QueryResult struct {
	File    string `json:"file"`
	Query   string `json:"-"`
	Heading string `json:"heading,omitempty"`
	Body    string `json:"body,omitempty"`
}

// Query represents a parsed query
type Query struct {
	Type  string // "frontmatter" or "section"
	Level int    // For section queries: heading level (1, 2, 3, etc.)
	Title string // For section queries: title to match (empty for any)
	Index int    // Index to match (-1 for first/default)
	Field string // For frontmatter queries: field name
}

// Options represents command-line options
type Options struct {
	HeadOnly     bool
	BodyOnly     bool
	JSONOutput   bool
	NoBlocks     bool
	RawOutput    bool
	ObjectOutput bool
	CSVOutput    bool
}
