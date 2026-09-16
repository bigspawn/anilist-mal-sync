package main

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// boolDefaultPattern finds the first backtick-quoted `true`/`false` literal
// on a doc line — the stable part of a hand-written sentence or table cell,
// regardless of how the surrounding prose or column layout is worded.
var boolDefaultPattern = regexp.MustCompile("`(true|false)`")

// proseDefaultPattern finds a default stated in words, as README's strategy
// chains and notes write it.
var proseDefaultPattern = regexp.MustCompile(`(enabled|disabled) by default`)

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

// TestDocs_MappingSourceProse_AgreesWithRegistry catches "enabled by default"
// wording in README's strategy chains. MangaBaka shipped documented as on
// while its default is off.
func TestDocs_MappingSourceProse_AgreesWithRegistry(t *testing.T) {
	readmeData, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}

	for _, src := range mappingSources() {
		t.Run(src.name, func(t *testing.T) {
			for line := range strings.SplitSeq(string(readmeData), "\n") {
				m := proseDefaultPattern.FindStringSubmatch(line)
				if m == nil || !strings.Contains(line, src.name+" API") {
					continue
				}
				if got := m[1] == "enabled"; got != src.defaultEnabled {
					t.Errorf("README.md prose: %s API default = %v, mappingSources() says %v\nline: %q",
						src.name, got, src.defaultEnabled, strings.TrimSpace(line))
				}
			}
		})
	}
}

// TestDocs_ConfigExamples_ListEveryMappingSource guards both YAML examples:
// MangaBaka was added to the flag and env docs but never to them.
func TestDocs_ConfigExamples_ListEveryMappingSource(t *testing.T) {
	readmeData, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("reading README.md: %v", err)
	}
	exampleData, err := os.ReadFile("config.example.yaml")
	if err != nil {
		t.Fatalf("reading config.example.yaml: %v", err)
	}

	examples := map[string]map[string]any{
		"README.md config.yaml example": parseYAMLExample(t, readmeConfigExample(t, string(readmeData))),
		"config.example.yaml":           parseYAMLExample(t, string(exampleData)),
	}

	for _, src := range mappingSources() {
		key := mappingSourceYAMLKey(t, src)
		t.Run(src.name, func(t *testing.T) {
			for location, example := range examples {
				checkYAMLExampleSection(t, location, example, key, src.defaultEnabled)
			}
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

// readmeConfigExample returns the YAML body of README's "Full config.yaml
// example" code block.
func readmeConfigExample(t *testing.T, readme string) string {
	t.Helper()

	_, afterHeading, found := strings.Cut(readme, "Full `config.yaml` example:")
	if !found {
		t.Fatal("README.md: no \"Full `config.yaml` example\" section")
	}
	_, block, found := strings.Cut(afterHeading, "```yaml\n")
	if !found {
		t.Fatal("README.md: config.yaml example has no ```yaml block")
	}
	body, _, found := strings.Cut(block, "\n```")
	if !found {
		t.Fatal("README.md: config.yaml example block is not closed")
	}
	return body
}

func parseYAMLExample(t *testing.T, data string) map[string]any {
	t.Helper()

	var parsed map[string]any
	err := yaml.Unmarshal([]byte(data), &parsed)
	if err != nil {
		t.Fatalf("parsing YAML example: %v", err)
	}
	return parsed
}

// mappingSourceYAMLKey reads the yaml tag of the Config field src points at,
// so the key comes from the struct rather than a hand-written list.
func mappingSourceYAMLKey(t *testing.T, src mappingSource) string {
	t.Helper()

	var cfg Config
	target := src.configField(&cfg)
	v := reflect.ValueOf(&cfg).Elem()
	for i := range v.NumField() {
		field, ok := reflect.TypeAssert[*MappingSourceConfig](v.Field(i).Addr())
		if ok && field == target {
			return v.Type().Field(i).Tag.Get("yaml")
		}
	}
	t.Fatalf("%s: configField does not point at a Config field", src.name)
	return ""
}

// checkYAMLExampleSection asserts the example has the source's section and
// that its enabled value matches the built-in default.
func checkYAMLExampleSection(t *testing.T, location string, example map[string]any, key string, want bool) {
	t.Helper()

	section, ok := example[key].(map[string]any)
	if !ok {
		t.Errorf("%s: no %q section", location, key)
		return
	}
	got, ok := section["enabled"].(bool)
	if !ok {
		t.Errorf("%s: %q section has no boolean enabled", location, key)
		return
	}
	if got != want {
		t.Errorf("%s: %s.enabled = %v, mappingSources() says %v", location, key, got, want)
	}
}
