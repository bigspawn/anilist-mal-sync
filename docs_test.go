package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// boolDefaultPattern finds the first backtick-quoted `true`/`false` literal
// on a doc line — the stable part of a hand-written sentence or table cell,
// regardless of how the surrounding prose or column layout is worded.
var boolDefaultPattern = regexp.MustCompile("`(true|false)`")

// TestDocs_MappingSourceDefaults_AgreeWithRegistry guards README.md and
// CLAUDE.md against silently drifting from mappingSources(): the class of
// bug behind the MangaBaka default fix, extended to the docs a user
// actually reads. It iterates mappingSources() itself — not a hand-written
// source list — so a future source that is registered but never documented
// fails here automatically. Offline DB is intentionally out of scope: it
// is not in mappingSources() (see mapping_sources.go).
func TestDocs_MappingSourceDefaults_AgreeWithRegistry(t *testing.T) {
	readmeData, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	claudeMDData, err := os.ReadFile("CLAUDE.md")
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}
	readme := string(readmeData)
	claudeMD := string(claudeMDData)

	for _, src := range mappingSources() {
		t.Run(src.name, func(t *testing.T) {
			checkDocDefault(t, "README.md CLI flag table", readme, backtickAnchor("--"+src.flag), src.defaultEnabled)
			checkDocDefault(t, "README.md env var list", readme, backtickAnchor(src.envPrefix+"ENABLED"), src.defaultEnabled)
			checkClaudeDefaultsTableRow(t, claudeMD, src)
		})
	}
}

// checkDocDefault finds the line in doc naming anchor (a flag or env var,
// backtick-quoted the way both docs write them) and asserts the `true`/
// `false` literal on that line matches want. A missing anchor or a missing
// literal is reported as its own distinct failure, not lumped together.
func checkDocDefault(t *testing.T, docLocation, doc, anchor string, want bool) {
	t.Helper()

	line, found := findLineContaining(doc, anchor)
	if !found {
		t.Errorf("%s: no entry for %s", docLocation, anchor)
		return
	}

	got, ok := extractBoolDefault(line)
	if !ok {
		t.Errorf("%s: entry for %s has no `true`/`false` default\nline: %q", docLocation, anchor, strings.TrimSpace(line))
		return
	}

	if got != want {
		t.Errorf("%s: %s default = %v, mappingSources() says %v\nline: %q", docLocation, anchor, got, want, strings.TrimSpace(line))
	}
}

// checkClaudeDefaultsTableRow is like checkDocDefault but additionally
// requires the flag and env var to appear on the SAME row, since CLAUDE.md
// documents both together in one table (README documents them in two
// separate places, checked individually by checkDocDefault).
func checkClaudeDefaultsTableRow(t *testing.T, doc string, src mappingSource) {
	t.Helper()

	flagAnchor := backtickAnchor("--" + src.flag)
	line, found := findLineContaining(doc, flagAnchor)
	if !found {
		t.Errorf("CLAUDE.md defaults table: no row for %s", flagAnchor)
		return
	}

	envAnchor := backtickAnchor(src.envPrefix + "ENABLED")
	if !strings.Contains(line, envAnchor) {
		t.Errorf("CLAUDE.md defaults table: row for %s does not also name %s\nrow: %q", flagAnchor, envAnchor, strings.TrimSpace(line))
		return
	}

	got, ok := extractBoolDefault(line)
	if !ok {
		t.Errorf("CLAUDE.md defaults table: row for %s has no `true`/`false` default\nrow: %q", flagAnchor, strings.TrimSpace(line))
		return
	}

	if got != src.defaultEnabled {
		t.Errorf("CLAUDE.md defaults table: %s default = %v, mappingSources() says %v\nrow: %q",
			flagAnchor, got, src.defaultEnabled, strings.TrimSpace(line))
	}
}

// backtickAnchor wraps s the way both docs quote a flag or env var name, so
// a search for "--hato-api" can't accidentally match "--hato-api-url".
func backtickAnchor(s string) string {
	return "`" + s + "`"
}

// findLineContaining returns the first line of doc containing anchor.
func findLineContaining(doc, anchor string) (string, bool) {
	for line := range strings.SplitSeq(doc, "\n") {
		if strings.Contains(line, anchor) {
			return line, true
		}
	}
	return "", false
}

// extractBoolDefault reads the first backtick-quoted true/false literal on
// line, left to right.
func extractBoolDefault(line string) (bool, bool) {
	m := boolDefaultPattern.FindStringSubmatch(line)
	if m == nil {
		return false, false
	}
	value, err := strconv.ParseBool(m[1])
	if err != nil {
		return false, false // unreachable: the pattern only matches true/false
	}
	return value, true
}
