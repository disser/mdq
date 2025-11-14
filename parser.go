package main

import (
	"bufio"
	"bytes"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseDocument parses a markdown file and extracts frontmatter and sections
func ParseDocument(content string, filePath string, noBlocks bool) (*Document, error) {
	doc := &Document{
		FilePath:    filePath,
		Frontmatter: make(map[string]interface{}),
		Sections:    []Section{},
	}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return doc, nil
	}

	// Parse frontmatter if present
	lineIdx := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		lineIdx = 1
		frontmatterLines := []string{}
		for lineIdx < len(lines) {
			if strings.TrimSpace(lines[lineIdx]) == "---" {
				lineIdx++
				break
			}
			frontmatterLines = append(frontmatterLines, lines[lineIdx])
			lineIdx++
		}

		if len(frontmatterLines) > 0 {
			frontmatterContent := strings.Join(frontmatterLines, "\n")
			yaml.Unmarshal([]byte(frontmatterContent), &doc.Frontmatter)
		}
	}

	// Parse sections
	levelCounts := make(map[int]int) // Track count of each heading level
	var currentSection *Section
	var bodyLines []string

	for lineIdx < len(lines) {
		line := lines[lineIdx]

		// Check if this is a heading
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			// Save the previous section if it exists
			if currentSection != nil {
				currentSection.Body = strings.TrimRight(strings.Join(bodyLines, "\n"), "\n")
				doc.Sections = append(doc.Sections, *currentSection)
				bodyLines = []string{}
			}

			// Parse the new heading
			level := 0
			trimmed := strings.TrimSpace(line)
			for i := 0; i < len(trimmed) && trimmed[i] == '#'; i++ {
				level++
			}

			title := strings.TrimSpace(trimmed[level:])

			levelCounts[level]++

			currentSection = &Section{
				Level:   level,
				Title:   title,
				Heading: line,
				Index:   levelCounts[level] - 1,
			}
		} else {
			// This is body content
			bodyLines = append(bodyLines, line)
		}

		lineIdx++
	}

	// Save the last section
	if currentSection != nil {
		currentSection.Body = strings.TrimRight(strings.Join(bodyLines, "\n"), "\n")
		doc.Sections = append(doc.Sections, *currentSection)
	}

	// Apply --no-blocks filter if requested
	if noBlocks {
		for i := range doc.Sections {
			doc.Sections[i].Body = removeCodeBlocks(doc.Sections[i].Body)
		}
	}

	return doc, nil
}

// removeCodeBlocks removes triple-backtick code blocks from text
func removeCodeBlocks(text string) string {
	var result strings.Builder
	scanner := bufio.NewScanner(bytes.NewBufferString(text))
	inCodeBlock := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if !inCodeBlock {
			result.WriteString(line)
			result.WriteString("\n")
		}
	}

	return strings.TrimRight(result.String(), "\n")
}
