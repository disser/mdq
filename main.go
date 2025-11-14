package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// parseQueryStrings splits comma-separated query strings
func parseQueryStrings(queryStr string) []string {
	parts := strings.Split(queryStr, ",")
	var queries []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			queries = append(queries, trimmed)
		}
	}
	return queries
}

func main() {
	// Define command-line flags with both short and long options
	var headOnly bool
	flag.BoolVar(&headOnly, "h", false, "Return only the heading (the matching element)")
	flag.BoolVar(&headOnly, "head", false, "Return only the heading (the matching element)")

	var bodyOnly bool
	flag.BoolVar(&bodyOnly, "b", false, "Return only the body (content before next section)")
	flag.BoolVar(&bodyOnly, "body", false, "Return only the body (content before next section)")

	var jsonOutput bool
	flag.BoolVar(&jsonOutput, "j", false, "Return results in JSON format")
	flag.BoolVar(&jsonOutput, "json", false, "Return results in JSON format")

	var noBlocks bool
	flag.BoolVar(&noBlocks, "n", false, "Omit text blocks within triple backticks")
	flag.BoolVar(&noBlocks, "no-blocks", false, "Omit text blocks within triple backticks")

	var rawOutput bool
	flag.BoolVar(&rawOutput, "r", false, "Raw output (only the found text, no filename or query)")
	flag.BoolVar(&rawOutput, "raw", false, "Raw output (only the found text, no filename or query)")

	var objectOutput bool
	flag.BoolVar(&objectOutput, "o", false, "JSON object output for multiple queries (use with -j)")
	flag.BoolVar(&objectOutput, "object", false, "JSON object output for multiple queries (use with --json)")

	var csvOutput bool
	flag.BoolVar(&csvOutput, "c", false, "CSV output format")
	flag.BoolVar(&csvOutput, "csv", false, "CSV output format")

	var markdownOutput bool
	flag.BoolVar(&markdownOutput, "m", false, "Markdown output (only the sections selected by the query)")
	flag.BoolVar(&markdownOutput, "markdown", false, "Markdown output (only the sections selected by the query)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mdq [-h|--head|-b|--body] [-j|--json] [-n|--no-blocks] QUERY [FILES...]\n\n")
		fmt.Fprintf(os.Stderr, "Query markdown files and extract information like 'jq' does for JSON.\n\n")
		fmt.Fprintf(os.Stderr, "Query syntax:\n")
		fmt.Fprintf(os.Stderr, "  #           First h1 block\n")
		fmt.Fprintf(os.Stderr, "  #[0]        First h1 block (explicit index)\n")
		fmt.Fprintf(os.Stderr, "  ##Notes     First h2 block titled \"Notes\"\n")
		fmt.Fprintf(os.Stderr, "  ##[3]       Fourth h2 in the document (0-indexed)\n")
		fmt.Fprintf(os.Stderr, "  date        \"date\" field from YAML frontmatter\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nIf no FILES are provided, reads from stdin.\n")
	}

	flag.Parse()

	// Check for conflicting flags
	if headOnly && bodyOnly {
		fmt.Fprintln(os.Stderr, "Error: -h/--head and -b/--body flags are mutually exclusive")
		os.Exit(1)
	}

	// Check for conflicting output formats
	outputFlags := 0
	if jsonOutput {
		outputFlags++
	}
	if csvOutput {
		outputFlags++
	}
	if markdownOutput {
		outputFlags++
	}
	if outputFlags > 1 {
		fmt.Fprintln(os.Stderr, "Error: -j/--json, -c/--csv, and -m/--markdown flags are mutually exclusive")
		os.Exit(1)
	}

	// Get query and files
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	queryStr := args[0]
	files := args[1:]

	// Parse comma-separated queries
	queryStrings := parseQueryStrings(queryStr)
	var queries []*Query
	for _, qs := range queryStrings {
		query, err := ParseQuery(qs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing query '%s': %v\n", qs, err)
			os.Exit(1)
		}
		queries = append(queries, query)
	}

	// Set up options
	opts := Options{
		HeadOnly:       headOnly,
		BodyOnly:       bodyOnly,
		JSONOutput:     jsonOutput,
		NoBlocks:       noBlocks,
		RawOutput:      rawOutput,
		ObjectOutput:   objectOutput,
		CSVOutput:      csvOutput,
		MarkdownOutput: markdownOutput,
	}

	var results []*QueryResult

	// Process files or stdin
	if len(files) == 0 {
		// Read from stdin
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			os.Exit(1)
		}

		doc, err := ParseDocument(string(content), "stdin", noBlocks)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing stdin: %v\n", err)
			os.Exit(1)
		}

		// Execute all queries against the document
		for _, query := range queries {
			result := ExecuteQuery(doc, query, opts)
			results = append(results, result)
		}
	} else {
		// Process each file
		for _, filePath := range files {
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
				continue
			}

			doc, err := ParseDocument(string(content), filePath, noBlocks)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", filePath, err)
				continue
			}

			// Execute all queries against the document
			for _, query := range queries {
				result := ExecuteQuery(doc, query, opts)
				results = append(results, result)
			}
		}
	}

	// Format and print output
	output := FormatOutput(results, opts)
	if output != "" {
		fmt.Println(output)
	}
}
