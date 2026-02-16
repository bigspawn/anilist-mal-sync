package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// ManualMapping represents a user-defined mapping between AniList and MAL IDs.
type ManualMapping struct {
	AniListID int    `yaml:"anilist_id"`
	MALID     int    `yaml:"mal_id"`
	Comment   string `yaml:"comment,omitempty"`
}

// IgnoreEntry stores metadata for ignored entries (for comments).
type IgnoreEntry struct {
	Title  string
	Reason string
}

// IgnoreConfig holds lists of entries to ignore during sync.
type IgnoreConfig struct {
	AniListIDs []int               `yaml:"anilist_ids,omitempty"`
	MALIDs     []int               `yaml:"mal_ids,omitempty"`
	Titles     []string            `yaml:"titles,omitempty"`
	metadata   map[int]IgnoreEntry // unexported, not in YAML (AniList IDs)
	malMeta    map[int]IgnoreEntry // unexported, not in YAML (MAL IDs)
}

// MappingsConfig holds user-editable mappings and ignore rules.
type MappingsConfig struct {
	ManualMappings []ManualMapping `yaml:"manual_mappings,omitempty"`
	Ignore         IgnoreConfig    `yaml:"ignore,omitempty"`
}

// LoadMappings loads mappings from YAML file. Creates empty file if it doesn't exist.
func LoadMappings(path string) (*MappingsConfig, error) {
	if path == "" {
		path = getDefaultMappingsPath()
	}

	data, err := os.ReadFile(path) // #nosec G304 - Path from user config
	if err != nil {
		if os.IsNotExist(err) {
			return &MappingsConfig{}, nil
		}
		return nil, fmt.Errorf("read mappings file: %w", err)
	}

	var cfg MappingsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse mappings file: %w", err)
	}

	return &cfg, nil
}

// Save writes the mappings config to a YAML file.
func (m *MappingsConfig) Save(path string) error {
	if path == "" {
		path = getDefaultMappingsPath()
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G301 - Config directory
		return fmt.Errorf("create mappings directory: %w", err)
	}

	file, err := os.Create(path) // #nosec G304 - Path from config
	if err != nil {
		return fmt.Errorf("create mappings file: %w", err)
	}
	defer func() { _ = file.Close() }() // #nosec ErrCheck - Error handling not needed for read-only file

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2) // Match default yaml.v2 indent
	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("marshal mappings: %w", err)
	}

	return nil
}

// GetManualMALID returns the MAL ID for a given AniList ID from manual mappings.
func (m *MappingsConfig) GetManualMALID(anilistID int) (int, bool) {
	for _, mapping := range m.ManualMappings {
		if mapping.AniListID == anilistID {
			return mapping.MALID, true
		}
	}
	return 0, false
}

// GetManualAniListID returns the AniList ID for a given MAL ID from manual mappings.
func (m *MappingsConfig) GetManualAniListID(malID int) (int, bool) {
	for _, mapping := range m.ManualMappings {
		if mapping.MALID == malID {
			return mapping.AniListID, true
		}
	}
	return 0, false
}

// IsIgnored checks if an entry should be ignored by AniList ID or title.
func (m *MappingsConfig) IsIgnored(anilistID int, title string) bool {
	for _, id := range m.Ignore.AniListIDs {
		if id == anilistID {
			return true
		}
	}
	lowerTitle := strings.ToLower(title)
	for _, t := range m.Ignore.Titles {
		if strings.ToLower(t) == lowerTitle {
			return true
		}
	}
	return false
}

// AddIgnoreByID adds an AniList ID to the ignore list with optional metadata.
func (m *MappingsConfig) AddIgnoreByID(anilistID int, title string, reason string) {
	for _, id := range m.Ignore.AniListIDs {
		if id == anilistID {
			return
		}
	}
	m.Ignore.AniListIDs = append(m.Ignore.AniListIDs, anilistID)

	if m.Ignore.metadata == nil {
		m.Ignore.metadata = make(map[int]IgnoreEntry)
	}
	m.Ignore.metadata[anilistID] = IgnoreEntry{Title: title, Reason: reason}
}

// IsIgnoredByMALID checks if an entry should be ignored by MAL ID.
func (m *MappingsConfig) IsIgnoredByMALID(malID int) bool {
	for _, id := range m.Ignore.MALIDs {
		if id == malID {
			return true
		}
	}
	return false
}

// AddIgnoreByMALID adds a MAL ID to the ignore list with optional metadata.
func (m *MappingsConfig) AddIgnoreByMALID(malID int, title string, reason string) {
	for _, id := range m.Ignore.MALIDs {
		if id == malID {
			return
		}
	}
	m.Ignore.MALIDs = append(m.Ignore.MALIDs, malID)

	if m.Ignore.malMeta == nil {
		m.Ignore.malMeta = make(map[int]IgnoreEntry)
	}
	m.Ignore.malMeta[malID] = IgnoreEntry{Title: title, Reason: reason}
}

// AddManualMapping adds a manual mapping if not already present.
func (m *MappingsConfig) AddManualMapping(anilistID, malID int, comment string) {
	for i, mapping := range m.ManualMappings {
		if mapping.AniListID == anilistID {
			m.ManualMappings[i].MALID = malID
			m.ManualMappings[i].Comment = comment
			return
		}
	}
	m.ManualMappings = append(m.ManualMappings, ManualMapping{
		AniListID: anilistID,
		MALID:     malID,
		Comment:   comment,
	})
}

func getDefaultMappingsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(configDir, "anilist-mal-sync", "mappings.yaml")
}

// buildIgnoreIDsNode creates a YAML sequence node for ignore IDs with inline comments.
func buildIgnoreIDsNode(ids []int, meta map[int]IgnoreEntry) *yaml.Node {
	seqNode := &yaml.Node{Kind: yaml.SequenceNode}
	for _, id := range ids {
		idNode := &yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", id)}
		if meta, ok := meta[id]; ok {
			var parts []string
			if meta.Title != "" {
				parts = append(parts, meta.Title)
			}
			if meta.Reason != "" {
				parts = append(parts, meta.Reason)
			}
			if len(parts) > 0 {
				idNode.LineComment = "# " + strings.Join(parts, " : ")
			}
		}
		seqNode.Content = append(seqNode.Content, idNode)
	}
	return seqNode
}

// MarshalYAML implements custom YAML marshaling with inline comments.
func (m *MappingsConfig) MarshalYAML() (interface{}, error) { //nolint:unparam // Error return required by yaml.Marshaler interface
	node := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	// Handle manual_mappings (standard serialization)
	if len(m.ManualMappings) > 0 {
		manualMappingsNode := &yaml.Node{Kind: yaml.SequenceNode}
		for _, mapping := range m.ManualMappings {
			mappingNode := &yaml.Node{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "anilist_id"},
					{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", mapping.AniListID)},
					{Kind: yaml.ScalarNode, Value: "mal_id"},
					{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", mapping.MALID)},
				},
			}
			if mapping.Comment != "" {
				mappingNode.Content = append(mappingNode.Content,
					[]*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "comment"},
						{Kind: yaml.ScalarNode, Value: mapping.Comment},
					}...,
				)
			}
			manualMappingsNode.Content = append(manualMappingsNode.Content, mappingNode)
		}
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "manual_mappings"},
			manualMappingsNode,
		)
	}

	// Handle ignore section
	hasIgnore := len(m.Ignore.AniListIDs) > 0 || len(m.Ignore.MALIDs) > 0 || len(m.Ignore.Titles) > 0
	if hasIgnore {
		ignoreNode := &yaml.Node{Kind: yaml.MappingNode}

		if len(m.Ignore.AniListIDs) > 0 {
			ignoreNode.Content = append(ignoreNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "anilist_ids"},
				buildIgnoreIDsNode(m.Ignore.AniListIDs, m.Ignore.metadata),
			)
		}

		if len(m.Ignore.MALIDs) > 0 {
			ignoreNode.Content = append(ignoreNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "mal_ids"},
				buildIgnoreIDsNode(m.Ignore.MALIDs, m.Ignore.malMeta),
			)
		}

		if len(m.Ignore.Titles) > 0 {
			titlesNode := &yaml.Node{Kind: yaml.SequenceNode}
			for _, title := range m.Ignore.Titles {
				titlesNode.Content = append(titlesNode.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: title},
				)
			}
			ignoreNode.Content = append(ignoreNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "titles"},
				titlesNode,
			)
		}

		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "ignore"},
			ignoreNode,
		)
	}

	return node, nil
}
