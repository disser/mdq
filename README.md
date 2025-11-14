# mdq - Markdown Query Tool

A command-line tool to query markdown files and extract information like `jq` does for JSON files.

## Installation

```bash
go install
```

The binary will be installed to `$GOPATH/bin/mdq` (typically `~/go/bin/mdq`).

## Usage

```bash
mdq [-h|-b] [-j] [--no-blocks] QUERY [FILES...]
```

If no FILES are provided, mdq reads from stdin.

## Query Syntax

### Section Queries

- `#` or `#[0]` - First h1 block
- `##Notes` - First h2 block titled "Notes"
- `##Notes[1]` - Second h2 block titled "Notes" (0-indexed)
- `##[3]` - Fourth h2 in the document (0-indexed)
- `###` - First h3 block

### Frontmatter Queries

Query YAML frontmatter fields by name:

- `date` - Returns the "date" field from frontmatter
- `title` - Returns the "title" field from frontmatter
- Any other frontmatter field name

### Multiple Queries

Query multiple fields at once using comma-separated queries:

- `"date, title"` - Returns both date and title fields
- `"amount, ##Notes"` - Returns amount field and Notes section
- `"date, title, author"` - Returns three frontmatter fields

## Options

- `-h` - Return only the heading (the matching element itself)
- `-b` - Return only the body (content before the next section)
- `-j` - Return results in JSON format
- `-r` - Raw output (only the found text, no filename or field label)
- `-o` - JSON object output for multiple queries (use with `-j`)
- `--csv` - CSV output format
- `--no-blocks` - Omit text blocks within triple backticks

**Note:** `-h` and `-b` are mutually exclusive. If neither is specified, both heading and body are returned.

## Examples

### Query frontmatter

```bash
# Get the date field
mdq date notes.md

# Get date in raw format (just the value)
mdq -r date notes.md

# Get title from multiple files in JSON format
mdq -j title file1.md file2.md

# Get raw values from multiple files (one per line)
mdq -r date file1.md file2.md
```

### Multiple queries

```bash
# Get multiple fields at once
mdq "date, title" notes.md
# Output:
# date
# 2025-11-13
#
# title
# My Document

# Get multiple fields in raw format
mdq -r "date, title" notes.md
# Output:
# 2025-11-13
# My Document

# Get multiple fields as JSON array
mdq -j "date, title" notes.md
# Output:
# [
#   {"file": "notes.md", "heading": "date", "body": "2025-11-13"},
#   {"file": "notes.md", "heading": "title", "body": "My Document"}
# ]

# Get multiple fields as JSON object (with -o flag)
mdq -j -o "date, title" notes.md
# Output:
# {
#   "file": "notes.md",
#   "date": "2025-11-13",
#   "title": "My Document"
# }

# Mix frontmatter and section queries
mdq "amount, ##Notes" notes.md
```

### Query sections

```bash
# Get first h1 heading and body
mdq "#" notes.md

# Get only the heading of the first h1
mdq -h "#" notes.md

# Get only the body of the Notes section
mdq -b "##Notes" notes.md

# Get the third h2 section
mdq "##[2]" notes.md

# Get second Notes section (when there are multiple with same title)
mdq "##Notes[1]" notes.md
```

### Filter code blocks

```bash
# Get Notes section without code blocks
mdq --no-blocks "##Notes" notes.md
```

### CSV output

```bash
# Get multiple fields as CSV
mdq --csv "date, title, author" *.md
# Output:
# file,date,title,author
# file1.md,2025-11-13,My Document,John Doe
# file2.md,2025-11-14,Another Doc,Jane Smith

# CSV is great for importing into spreadsheets or databases
mdq --csv "amount, date, ##Notes" trips/*.md > expenses.csv

# Combine with other tools
mdq --csv "date, title" *.md | tail -n +2 | sort -t, -k2
```

### JSON output

```bash
# Get results in JSON format
mdq -j "##Summary" notes.md

# Output (single file):
{
  "file": "notes.md",
  "query": "##Summary",
  "heading": "## Summary",
  "body": "This is the summary content."
}

# Multiple files return an array
mdq -j "date" file1.md file2.md
```

### Read from stdin

```bash
# Pipe content to mdq
cat notes.md | mdq "##Notes"

# Use with clipboard on macOS
pbpaste | mdq -j "title"

# Chain with other commands
curl https://example.com/README.md | mdq "#Installation"
```

### Query multiple files

```bash
# Get the same section from multiple files
mdq "##Notes" *.md

# With file headers in output
mdq "date" file1.md file2.md
# Output:
# ==> file1.md <==
# date
# 2025-11-13
# 
# ==> file2.md <==
# date
# 2025-11-14
```

## Example Markdown File

```markdown
---
title: My Document
date: 2025-11-13
author: John Doe
---

# Introduction

This is the introduction.

## Background

Some background information.

## Notes

Important notes here.

```javascript
console.log("code example");
```

More notes.

## Summary

Final summary.
```

### Query Examples with Above File

```bash
# Get frontmatter date
mdq date example.md
# Output:
# date
# 2025-11-13

# Get first h1
mdq "#" example.md
# Output:
# # Introduction
# 
# This is the introduction.

# Get Notes section without code blocks
mdq --no-blocks "##Notes" example.md
# Output:
# ## Notes
# 
# Important notes here.
# 
# More notes.

# Get only Notes heading
mdq -h "##Notes" example.md
# Output:
# ## Notes
```

## Project Structure

```
mdq/
├── main.go       # CLI entry point and argument parsing
├── types.go      # Data structures (Document, Section, Query, etc.)
├── parser.go     # Markdown and YAML frontmatter parser
├── query.go      # Query parser and executor
├── output.go     # Output formatters (text and JSON)
├── go.mod        # Go module definition
└── README.md     # This file
```

## License

This project is open source and available under standard terms.
