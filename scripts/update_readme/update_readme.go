package main

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func main() {
	b := bytes.NewBuffer(nil)
	cmd := exec.Command("mmake", "-h")
	cmd.Stdout = b
	cmd.Stderr = b

	if err := cmd.Run(); err != nil {
		panic(err)
	}

	f, err := os.Open("README.md")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// find '## Usage'
	// find '```'
	// find '```'
	var out []string
	scanner := bufio.NewScanner(f)
	var inUsage bool
	var inCodeBlock bool
	for scanner.Scan() {
		txt := scanner.Text()

		if inUsage && inCodeBlock {
			if txt == "```" {
				inCodeBlock = false
				inUsage = false
				out = append(out, strings.TrimSuffix(b.String(), "\n\n"))
			} else {
				continue // ignore text between
			}
		}
		if txt == "## Usage" {
			inUsage = true
		}
		if txt == "```" && inUsage {
			inCodeBlock = true
		}

		out = append(out, txt)
	}

	f.Close()

	f, err = os.Create("README.md")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range out {
		w.WriteString(line)
		w.WriteString("\n")
	}
	w.Flush()
	f.Close()
}
