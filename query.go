package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParseQuery parses a query string into a Query object
func ParseQuery(queryStr string) (*Query, error) {
	query := &Query{
		Index: 0, // Default to first match
	}

	// Check if it's a section query (starts with #)
	if strings.HasPrefix(queryStr, "#") {
		query.Type = "section"

		// Count the heading level
		level := 0
		for i := 0; i < len(queryStr) && queryStr[i] == '#'; i++ {
			level++
		}
		query.Level = level

		// Get the rest after the # symbols
		rest := queryStr[level:]

		// Check for index in brackets: [N]
		indexPattern := regexp.MustCompile(`^(.*?)\[(\d+)]$`)
		if matches := indexPattern.FindStringSubmatch(rest); matches != nil {
			title := strings.TrimSpace(matches[1])
			index, _ := strconv.Atoi(matches[2])
			query.Title = title
			query.Index = index
		} else {
			query.Title = strings.TrimSpace(rest)
			query.Index = 0
		}

		return query, nil
	}

	// Otherwise, it's a frontmatter field query
	query.Type = "frontmatter"
	query.Field = queryStr

	return query, nil
}

// ExecuteQuery executes a query against a document
func ExecuteQuery(doc *Document, query *Query, opts Options) *QueryResult {
	result := &QueryResult{
		File:  doc.FilePath,
		Query: formatQuery(query),
	}

	if query.Type == "frontmatter" {
		// Query frontmatter
		if value, ok := doc.Frontmatter[query.Field]; ok {
			// Handle nil values (empty YAML fields) as empty strings
			var bodyStr string
			if value != nil {
				bodyStr = fmt.Sprintf("%v", value)
			}

			if !opts.HeadOnly {
				result.Body = bodyStr
			}
			// In raw mode, don't set heading for frontmatter
			if !opts.BodyOnly && !opts.RawOutput {
				result.Heading = query.Field
			}
		}
		return result
	}

	// Query sections
	matchIndex := 0
	for _, section := range doc.Sections {
		// Check if level matches
		if section.Level != query.Level {
			continue
		}

		// Check if title matches (if specified)
		if query.Title != "" && section.Title != query.Title {
			continue
		}

		// Check if this is the right index
		if matchIndex == query.Index {
			if !opts.HeadOnly {
				result.Body = section.Body
			}
			if !opts.BodyOnly {
				result.Heading = section.Heading
			}
			return result
		}

		matchIndex++
	}

	return result
}

// formatQuery converts a Query back to a string representation
func formatQuery(q *Query) string {
	if q.Type == "frontmatter" {
		return q.Field
	}

	// Section query
	var sb strings.Builder
	for i := 0; i < q.Level; i++ {
		sb.WriteString("#")
	}
	sb.WriteString(q.Title)
	if q.Index > 0 {
		sb.WriteString(fmt.Sprintf("[%d]", q.Index))
	}
	return sb.String()
}
