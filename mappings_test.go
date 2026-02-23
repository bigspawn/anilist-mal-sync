package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
)

func TestLoadMappings_FileNotFound(t *testing.T) {
	t.Parallel()
	cfg, err := LoadMappings("/tmp/nonexistent-mappings-test.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.ManualMappings)
	assert.Empty(t, cfg.Ignore.AniListIDs)
}

func TestLoadMappings_ValidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")

	content := `manual_mappings:
- anilist_id: 100
  mal_id: 200
  comment: test mapping
ignore:
  anilist_ids:
  - 300
  - 400
  titles:
  - "Some Title"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadMappings(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Len(t, cfg.ManualMappings, 1)
	assert.Equal(t, 100, cfg.ManualMappings[0].AniListID)
	assert.Equal(t, 200, cfg.ManualMappings[0].MALID)
	assert.Equal(t, "test mapping", cfg.ManualMappings[0].Comment)
	assert.Equal(t, []int{300, 400}, cfg.Ignore.AniListIDs)
	assert.Equal(t, []string{"Some Title"}, cfg.Ignore.Titles)
}

func TestMappingsConfig_SaveAndLoad(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")

	cfg := &MappingsConfig{
		ManualMappings: []ManualMapping{
			{AniListID: 1, MALID: 2, Comment: "test"},
		},
		Ignore: IgnoreConfig{
			AniListIDs: []int{10, 20},
			Titles:     []string{"ignore me"},
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadMappings(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg.ManualMappings, loaded.ManualMappings)
	assert.Equal(t, cfg.Ignore.AniListIDs, loaded.Ignore.AniListIDs)
	assert.Equal(t, cfg.Ignore.Titles, loaded.Ignore.Titles)
}

func TestMappingsConfig_GetManualMALID(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		ManualMappings: []ManualMapping{
			{AniListID: 100, MALID: 200},
			{AniListID: 300, MALID: 400},
		},
	}

	id, ok := cfg.GetManualMALID(100)
	assert.True(t, ok)
	assert.Equal(t, 200, id)

	id, ok = cfg.GetManualMALID(300)
	assert.True(t, ok)
	assert.Equal(t, 400, id)

	_, ok = cfg.GetManualMALID(999)
	assert.False(t, ok)
}

func TestMappingsConfig_GetManualAniListID(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		ManualMappings: []ManualMapping{
			{AniListID: 100, MALID: 200},
		},
	}

	id, ok := cfg.GetManualAniListID(200)
	assert.True(t, ok)
	assert.Equal(t, 100, id)

	_, ok = cfg.GetManualAniListID(999)
	assert.False(t, ok)
}

func TestMappingsConfig_IsIgnored(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		Ignore: IgnoreConfig{
			AniListIDs: []int{100, 200},
			Titles:     []string{"Ignore Me"},
		},
	}

	assert.True(t, cfg.IsIgnored(100, "anything"))
	assert.True(t, cfg.IsIgnored(200, "anything"))
	assert.True(t, cfg.IsIgnored(0, "ignore me")) // case insensitive
	assert.True(t, cfg.IsIgnored(0, "IGNORE ME")) // case insensitive
	assert.False(t, cfg.IsIgnored(999, "not ignored"))
}

func TestMappingsConfig_AddIgnoreByID(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{}

	cfg.AddIgnoreByID(100, "Test Title", "Test Reason")
	assert.Equal(t, []int{100}, cfg.Ignore.AniListIDs)

	// Should not add duplicate
	cfg.AddIgnoreByID(100, "New Title", "New Reason")
	assert.Equal(t, []int{100}, cfg.Ignore.AniListIDs)

	cfg.AddIgnoreByID(200, "Another Title", "")
	assert.Equal(t, []int{100, 200}, cfg.Ignore.AniListIDs)
}

func TestMappingsConfig_IsIgnoredByMALID(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		Ignore: IgnoreConfig{
			MALIDs: []int{500, 600},
		},
	}

	assert.True(t, cfg.IsIgnoredByMALID(500))
	assert.True(t, cfg.IsIgnoredByMALID(600))
	assert.False(t, cfg.IsIgnoredByMALID(999))
}

func TestMappingsConfig_AddIgnoreByMALID(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{}

	cfg.AddIgnoreByMALID(500, "Test Title", "Test Reason")
	assert.Equal(t, []int{500}, cfg.Ignore.MALIDs)

	// Should not add duplicate
	cfg.AddIgnoreByMALID(500, "New Title", "New Reason")
	assert.Equal(t, []int{500}, cfg.Ignore.MALIDs)

	cfg.AddIgnoreByMALID(600, "Another Title", "")
	assert.Equal(t, []int{500, 600}, cfg.Ignore.MALIDs)

	// Check metadata is stored
	assert.NotNil(t, cfg.Ignore.malMeta)
	assert.Equal(t, IgnoreEntry{Title: "Test Title", Reason: "Test Reason"}, cfg.Ignore.malMeta[500])
	assert.Equal(t, IgnoreEntry{Title: "Another Title", Reason: ""}, cfg.Ignore.malMeta[600])
}

func TestMappingsConfig_AddManualMapping(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{}

	cfg.AddManualMapping(100, 200, "test")
	assert.Len(t, cfg.ManualMappings, 1)
	assert.Equal(t, 100, cfg.ManualMappings[0].AniListID)
	assert.Equal(t, 200, cfg.ManualMappings[0].MALID)

	// Update existing mapping
	cfg.AddManualMapping(100, 300, "updated")
	assert.Len(t, cfg.ManualMappings, 1)
	assert.Equal(t, 300, cfg.ManualMappings[0].MALID)
	assert.Equal(t, "updated", cfg.ManualMappings[0].Comment)

	// Add new mapping
	cfg.AddManualMapping(400, 500, "new")
	assert.Len(t, cfg.ManualMappings, 2)
}

func TestLoadMappings_InvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")
	if err := os.WriteFile(path, []byte("{{invalid yaml"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMappings(path)
	assert.Error(t, err)
}

func TestMappingsConfig_MarshalWithComments(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		Ignore: IgnoreConfig{
			AniListIDs: []int{100, 200},
		},
	}

	// Add metadata for the IDs
	cfg.Ignore.metadata = make(map[int]IgnoreEntry)
	cfg.Ignore.metadata[100] = IgnoreEntry{Title: "Test Anime", Reason: "Different version"}
	cfg.Ignore.metadata[200] = IgnoreEntry{Title: "Another Anime", Reason: ""}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)

	// Check that comments are present (LineComment in yaml.v3 adds inline comments)
	yamlStr := string(data)
	assert.Contains(t, yamlStr, "# Test Anime : Different version")
	assert.Contains(t, yamlStr, "# Another Anime")
}

func TestMappingsConfig_MarshalWithMALIDComments(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		Ignore: IgnoreConfig{
			MALIDs: []int{500, 600},
		},
	}

	cfg.Ignore.malMeta = map[int]IgnoreEntry{
		500: {Title: "MAL Anime", Reason: "reverse sync"},
		600: {Title: "Another MAL Anime"},
	}

	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "mal_ids")
	assert.Contains(t, yamlStr, "# MAL Anime : reverse sync")
	assert.Contains(t, yamlStr, "# Another MAL Anime")
}

func TestMappingsConfig_BackwardCompatibility(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")

	// Create old-style YAML without metadata
	content := `ignore:
  anilist_ids:
  - 300
  - 400
  titles:
  - "Some Title"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	// Load should work fine
	cfg, err := LoadMappings(path)
	assert.NoError(t, err)
	assert.Equal(t, []int{300, 400}, cfg.Ignore.AniListIDs)
	assert.Equal(t, []string{"Some Title"}, cfg.Ignore.Titles)

	// Metadata map should be nil (not initialized)
	assert.Nil(t, cfg.Ignore.metadata)
}

func TestMappingsConfig_SaveAndLoad_WithMALIDs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "mappings.yaml")

	cfg := &MappingsConfig{
		ManualMappings: []ManualMapping{
			{AniListID: 1, MALID: 2, Comment: "test"},
		},
		Ignore: IgnoreConfig{
			AniListIDs: []int{10, 20},
			MALIDs:     []int{30, 40},
			Titles:     []string{"ignore me"},
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadMappings(path)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, cfg.ManualMappings, loaded.ManualMappings)
	assert.Equal(t, cfg.Ignore.AniListIDs, loaded.Ignore.AniListIDs)
	assert.Equal(t, cfg.Ignore.MALIDs, loaded.Ignore.MALIDs)
	assert.Equal(t, cfg.Ignore.Titles, loaded.Ignore.Titles)
}

func TestMappingsConfig_MarshalYAML_CombinedIgnore(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{
		Ignore: IgnoreConfig{
			AniListIDs: []int{100},
			MALIDs:     []int{500},
		},
	}
	cfg.Ignore.metadata = map[int]IgnoreEntry{
		100: {Title: "AniList Entry"},
	}
	cfg.Ignore.malMeta = map[int]IgnoreEntry{
		500: {Title: "MAL Entry", Reason: "reverse"},
	}

	data, err := yaml.Marshal(cfg)
	assert.NoError(t, err)

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "anilist_ids")
	assert.Contains(t, yamlStr, "mal_ids")
	assert.Contains(t, yamlStr, "# AniList Entry")
	assert.Contains(t, yamlStr, "# MAL Entry : reverse")
}

func TestIgnoreConfig_Metadata(t *testing.T) {
	t.Parallel()
	cfg := &MappingsConfig{}

	// Add IDs with metadata
	cfg.AddIgnoreByID(100, "Test Title", "Test Reason")
	cfg.AddIgnoreByID(200, "Another Title", "")

	// Check IDs are stored
	assert.Equal(t, []int{100, 200}, cfg.Ignore.AniListIDs)

	// Check metadata is stored
	assert.NotNil(t, cfg.Ignore.metadata)
	assert.Equal(t, IgnoreEntry{Title: "Test Title", Reason: "Test Reason"}, cfg.Ignore.metadata[100])
	assert.Equal(t, IgnoreEntry{Title: "Another Title", Reason: ""}, cfg.Ignore.metadata[200])

	// Duplicate should not be added
	cfg.AddIgnoreByID(100, "New Title", "New Reason")
	assert.Equal(t, []int{100, 200}, cfg.Ignore.AniListIDs)
	// Original metadata should be preserved
	assert.Equal(t, IgnoreEntry{Title: "Test Title", Reason: "Test Reason"}, cfg.Ignore.metadata[100])
}
