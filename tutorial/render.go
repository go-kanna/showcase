//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := flag.String("dir", "tutorial", "tutorial source directory")
	flag.Parse()
	if err := render(*dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func render(dir string) error {
	files := []struct {
		name, destination string
		append            bool
	}{
		{"post", "model/post.go", false},
		{"scopes", "scopes/scopes.go", false},
		{"post_repository", "repository/post.go", false},
		{"input", "input.go", false},
		{"post_mapper_base", "mapper/mapper.go", false},
		{"api", "api.go", false},
		{"main", "main.go", false},
		{"comment", "model/comment.go", false},
		{"comment_repository", "repository/comment.go", false},
		{"comment_input", "comment_input.go", false},
		{"comment_response", "mapper/mapper.go", true},
		{"comments", "comments.go", false},
		{"repository_test", "repository/repository_test.go", false},
	}
	output := filepath.Join(dir, "..", "tmp", "tutorial-tapes")
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("create tape directory: %w", err)
	}
	for _, file := range files {
		source, err := os.ReadFile(filepath.Join(dir, "files", file.name+".go.txt"))
		if err != nil {
			return fmt.Errorf("read %s: %w", file.name, err)
		}
		tape, err := scene(string(source), file.destination, file.append)
		if err != nil {
			return fmt.Errorf("render %s: %w", file.name, err)
		}
		if err := os.WriteFile(filepath.Join(output, file.name+".tape"), []byte(tape), 0o644); err != nil {
			return fmt.Errorf("write %s tape: %w", file.name, err)
		}
	}
	return nil
}

func scene(source, destination string, appendSource bool) (string, error) {
	var tape strings.Builder
	fmt.Fprintf(&tape, "Type \"vim %s\"\nEnter\n", destination)
	fmt.Fprintf(&tape, "Wait+Screen@10s /%s/\n", strings.ReplaceAll(destination, "/", `\/`))
	if appendSource {
		tape.WriteString("Type \"Go\"\n")
	} else {
		tape.WriteString("Type \"i\"\n")
	}
	tape.WriteString("Wait+Screen@10s /INSERT/\n")
	speeds := [...]int{60, 83, 71, 90, 68}
	index := 0
	for line := range strings.Lines(source) {
		line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
		content := strings.TrimLeft(line, "\t")
		if indentation := len(line) - len(content); indentation > 0 {
			fmt.Fprintf(&tape, "Tab@60ms %d\n", indentation)
		}
		content = expandTabs(content)
		if content != "" {
			quoted, err := quote(content)
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&tape, "Type@%dms %s\n", speeds[index%len(speeds)], quoted)
		}
		tape.WriteString("Enter\n")
		if content == "" {
			tape.WriteString("Sleep 400ms\n")
		} else if content == "}" {
			tape.WriteString("Sleep 700ms\n")
		}
		index++
	}
	tape.WriteString("Escape\nType \"gg\"\nSleep 2s\nType \":wq\"\nEnter\nWait+Line@10s /blog>/\n")
	return tape.String(), nil
}

func quote(text string) (string, error) {
	for _, delimiter := range []string{"`", "'", `"`} {
		if !strings.Contains(text, delimiter) {
			return delimiter + text + delimiter, nil
		}
	}
	return "", fmt.Errorf("cannot quote a VHS line: %q", text)
}

func expandTabs(text string) string {
	var expanded strings.Builder
	column := 0
	for _, char := range text {
		if char == '\t' {
			spaces := 4 - column%4
			expanded.WriteString(strings.Repeat(" ", spaces))
			column += spaces
		} else {
			expanded.WriteRune(char)
			column++
		}
	}
	return expanded.String()
}
