package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"
)

// escapeCSV escapes a string for CSV output
func escapeCSV(s string) string {
	// Remove newlines and extra whitespace for CSV
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	// Collapse multiple spaces
	s = strings.Join(strings.Fields(s), " ")
	return s
}

// formatCSV formats results as CSV
func formatCSV(results []*QueryResult) string {
	if len(results) == 0 {
		return ""
	}

	var output strings.Builder
	writer := csv.NewWriter(&output)

	// Collect query names (preserve order from first occurrence)
	queryNames := []string{}
	seenQueries := make(map[string]bool)

	for _, result := range results {
		if result.Query != "" && !seenQueries[result.Query] {
			queryNames = append(queryNames, result.Query)
			seenQueries[result.Query] = true
		}
	}

	// Write header
	header := []string{"file"}
	header = append(header, queryNames...)
	writer.Write(header)

	// Group results by file
	type fileData struct {
		file   string
		values map[string]string
	}

	fileMap := make(map[string]*fileData)
	var fileOrder []string

	for _, result := range results {
		if _, ok := fileMap[result.File]; !ok {
			fileMap[result.File] = &fileData{
				file:   result.File,
				values: make(map[string]string),
			}
			fileOrder = append(fileOrder, result.File)
		}

		// Get value for this query - CSV should only use Body (not the label/heading)
		var value string
		if result.Body != "" {
			value = result.Body
		}
		// For CSV, empty properties should remain empty, not show the field name

		fileMap[result.File].values[result.Query] = escapeCSV(value)
	}

	// Write rows
	for _, fileName := range fileOrder {
		fd := fileMap[fileName]
		row := []string{fd.file}

		for _, queryName := range queryNames {
			row = append(row, fd.values[queryName])
		}

		writer.Write(row)
	}

	writer.Flush()
	return strings.TrimRight(output.String(), "\n")
}

// FormatOutput formats query results for display
func FormatOutput(results []*QueryResult, opts Options) string {
	if opts.CSVOutput {
		return formatCSV(results)
	}
	if opts.JSONOutput {
		return formatJSON(results, opts)
	}
	if opts.MarkdownOutput {
		return formatMarkdown(results, opts)
	}
	return formatText(results, opts)
}

// formatMarkdown formats results as markdown, including only the sections selected by the query
func formatMarkdown(results []*QueryResult, opts Options) string {
	var output strings.Builder

	// Group results by file for better formatting
	type fileGroup struct {
		file    string
		results []*QueryResult
	}

	var groups []fileGroup
	currentFile := ""
	var currentResults []*QueryResult

	for _, result := range results {
		if result.File != currentFile {
			if currentFile != "" {
				groups = append(groups, fileGroup{currentFile, currentResults})
			}
			currentFile = result.File
			currentResults = []*QueryResult{result}
		} else {
			currentResults = append(currentResults, result)
		}
	}

	if currentFile != "" {
		groups = append(groups, fileGroup{currentFile, currentResults})
	}

	// Track if frontmatter has been added for each file
	frontmatterAdded := make(map[string]bool)

	// Format output
	for gi, group := range groups {
		hasFrontmatter := false

		// Add file comment if multiple files
		if len(groups) > 1 {
			if gi > 0 {
				output.WriteString("\n")
			}
			output.WriteString(fmt.Sprintf("<!-- File: %s -->\n\n", group.file))
		}

		// First pass: identify if there are frontmatter queries with non-empty values
		for _, result := range group.results {
			// A frontmatter query will have result.Query that doesn't start with #
			if !strings.HasPrefix(result.Query, "#") {
				// Only mark as having frontmatter if there's actual content
				if result.Body != "" {
					hasFrontmatter = true
					break
				}
			}
		}

		// Output frontmatter if present
		if hasFrontmatter && !frontmatterAdded[group.file] {
			output.WriteString("---\n")
			for _, result := range group.results {
				// Only include frontmatter fields that were queried
				if !strings.HasPrefix(result.Query, "#") {
					// Get the field name - use result.Heading if available, otherwise use result.Query
					// (when -b flag is used, result.Heading will be empty)
					fieldName := result.Heading
					if fieldName == "" {
						fieldName = result.Query
					}

					if result.Body != "" {
						// Format the value
						output.WriteString(fmt.Sprintf("%s: %s\n", fieldName, result.Body))
					} else {
						// Empty value
						output.WriteString(fmt.Sprintf("%s: \"\"\n", fieldName))
					}
				}
			}
			output.WriteString("---\n\n")
			frontmatterAdded[group.file] = true
		}

		// Output each section result
		for ri, result := range group.results {
			// Skip frontmatter fields (already handled above)
			if !strings.HasPrefix(result.Query, "#") {
				continue
			}

			// Skip empty results
			if result.Heading == "" && result.Body == "" {
				continue
			}

			// Add blank line between multiple section results
			if ri > 0 {
				output.WriteString("\n")
			}

			// Output heading if present
			if result.Heading != "" && !opts.BodyOnly {
				output.WriteString(result.Heading)
				if result.Body != "" && !opts.HeadOnly {
					output.WriteString("\n\n")
				}
			}

			// Output body if present
			if result.Body != "" && !opts.HeadOnly {
				output.WriteString(result.Body)
			}

			output.WriteString("\n")
		}
	}

	return strings.TrimRight(output.String(), "\n")
}

// formatJSON formats results as JSON
func formatJSON(results []*QueryResult, opts Options) string {
	// Object output mode: combine multiple queries per file into single objects
	if opts.ObjectOutput {
		return formatJSONObject(results)
	}

	// If only one result, output as single object
	if len(results) == 1 {
		data, err := json.MarshalIndent(results[0], "", "  ")
		if err != nil {
			return ""
		}
		return string(data)
	}

	// Multiple results, output as array
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

// formatJSONObject formats results as objects with query results as fields
func formatJSONObject(results []*QueryResult) string {
	// Group results by file
	fileResults := make(map[string]map[string]interface{})

	for _, result := range results {
		if _, ok := fileResults[result.File]; !ok {
			fileResults[result.File] = make(map[string]interface{})
			fileResults[result.File]["file"] = result.File
		}

		// Use the query string as the key
		queryKey := result.Query
		if queryKey == "" {
			continue
		}

		// For object output, just use the body value (not the heading label)
		// Empty values should remain empty, not show the field name
		var value string
		if result.Body != "" {
			value = result.Body
		}

		fileResults[result.File][queryKey] = value
	}

	// If only one file, return as single object
	if len(fileResults) == 1 {
		for _, obj := range fileResults {
			data, err := json.MarshalIndent(obj, "", "  ")
			if err != nil {
				return ""
			}
			return string(data)
		}
	}

	// Multiple files, return as array of objects
	var objects []map[string]interface{}
	for _, obj := range fileResults {
		objects = append(objects, obj)
	}

	data, err := json.MarshalIndent(objects, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

// formatText formats results as plain text
func formatText(results []*QueryResult, opts Options) string {
	var output strings.Builder

	// Raw mode: only output the found text
	if opts.RawOutput {
		for _, result := range results {
			// Skip empty results
			if result.Heading == "" && result.Body == "" {
				continue
			}

			// Output heading if present
			if result.Heading != "" && !opts.BodyOnly {
				output.WriteString(result.Heading)
				if result.Body != "" && !opts.HeadOnly {
					output.WriteString("\n")
				}
			}

			// Output body if present
			if result.Body != "" && !opts.HeadOnly {
				output.WriteString(result.Body)
			}

			output.WriteString("\n")
		}
		return strings.TrimRight(output.String(), "\n")
	}

	// Group results by file for better formatting
	type fileGroup struct {
		file    string
		results []*QueryResult
	}

	var groups []fileGroup
	currentFile := ""
	var currentResults []*QueryResult

	for _, result := range results {
		if result.File != currentFile {
			if currentFile != "" {
				groups = append(groups, fileGroup{currentFile, currentResults})
			}
			currentFile = result.File
			currentResults = []*QueryResult{result}
		} else {
			currentResults = append(currentResults, result)
		}
	}
	if currentFile != "" {
		groups = append(groups, fileGroup{currentFile, currentResults})
	}

	// Format output
	for gi, group := range groups {
		// Add file prefix if multiple files
		if len(groups) > 1 {
			if gi > 0 {
				output.WriteString("\n")
			}
			output.WriteString(fmt.Sprintf("==> %s <==\n", group.file))
		}

		// Output each result
		for ri, result := range group.results {
			// Skip empty results
			if result.Heading == "" && result.Body == "" {
				continue
			}

			// Add blank line between multiple query results (but not for single query)
			if ri > 0 && len(group.results) > 1 {
				output.WriteString("\n")
			}

			// Output heading if present
			if result.Heading != "" && !opts.BodyOnly {
				output.WriteString(result.Heading)
				if result.Body != "" && !opts.HeadOnly {
					output.WriteString("\n")
				}
			}

			// Output body if present
			if result.Body != "" && !opts.HeadOnly {
				output.WriteString(result.Body)
			}

			output.WriteString("\n")
		}
	}

	return strings.TrimRight(output.String(), "\n")
}
