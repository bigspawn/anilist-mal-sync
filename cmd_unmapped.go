package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"
)

func newUnmappedCommand() *cli.Command {
	return &cli.Command{
		Name:  "unmapped",
		Usage: "Show and manage unmapped entries from last sync",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "interactively fix unmapped entries",
			},
			&cli.BoolFlag{
				Name:  "ignore-all",
				Usage: "add all unmapped entries to ignore list",
			},
		},
		Action: runUnmapped,
	}
}

func runUnmapped(_ context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")
	mappingsPath := resolveMappingsPath(configPath)

	state, err := LoadUnmappedState("")
	if err != nil {
		return fmt.Errorf("load unmapped state: %w", err)
	}

	if len(state.Entries) == 0 {
		log.Println("No unmapped entries found. Run 'sync' first.")
		return nil
	}

	if cmd.Bool("ignore-all") {
		return runUnmappedIgnoreAll(state, mappingsPath)
	}

	if cmd.Bool("fix") {
		return runUnmappedFix(state, mappingsPath)
	}

	return runUnmappedList(state)
}

func resolveMappingsPath(configPath string) string {
	if configPath == "" {
		return ""
	}
	config, err := loadConfigFromFile(configPath)
	if err != nil {
		return ""
	}
	return config.MappingsFilePath
}

func runUnmappedList(state *UnmappedState) error {
	log.Printf("Unmapped entries (%d) from last sync (%s):\n",
		len(state.Entries), state.UpdatedAt.Format("2006-01-02 15:04:05"))

	for i, entry := range state.Entries {
		printUnmappedEntry(i+1, entry)
	}

	log.Println("\nUse --fix to interactively manage these entries")
	log.Println("Use --ignore-all to add all to ignore list")
	return nil
}

func printUnmappedEntry(num int, entry UnmappedEntry) {
	mediaLabel := capitalizeFirst(entry.MediaType)
	log.Println(formatUnmappedLine(num, entry, mediaLabel))
}

func runUnmappedIgnoreAll(state *UnmappedState, mappingsPath string) error {
	mappings, err := LoadMappings(mappingsPath)
	if err != nil {
		return fmt.Errorf("load mappings: %w", err)
	}

	added := 0
	for _, entry := range state.Entries {
		if entry.AniListID > 0 && !mappings.IsIgnored(entry.AniListID, entry.Title) {
			mappings.AddIgnoreByID(entry.AniListID, entry.Title, entry.Reason)
			added++
			log.Printf("  + %q (AniList: %d)", entry.Title, entry.AniListID)
		}
	}

	if added == 0 {
		log.Println("All entries are already in the ignore list.")
		return nil
	}

	if err := mappings.Save(mappingsPath); err != nil {
		return fmt.Errorf("save mappings: %w", err)
	}

	savePath := mappingsPath
	if savePath == "" {
		savePath = getDefaultMappingsPath()
	}
	log.Printf("Added %d entries to ignore list in %s", added, savePath)
	return nil
}

func runUnmappedFix(state *UnmappedState, mappingsPath string) error {
	mappings, err := LoadMappings(mappingsPath)
	if err != nil {
		return fmt.Errorf("load mappings: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	changed := false

	for i, entry := range state.Entries {
		printFixHeader(i+1, len(state.Entries), entry)

		action := promptFixAction(reader)
		changed = applyFixAction(action, entry, mappings, reader) || changed
		if action == "q" {
			break
		}
	}

	if changed {
		if err := mappings.Save(mappingsPath); err != nil {
			return fmt.Errorf("save mappings: %w", err)
		}
		savePath := mappingsPath
		if savePath == "" {
			savePath = getDefaultMappingsPath()
		}
		log.Printf("Saved changes to %s", savePath)
	}
	return nil
}

func printFixHeader(num, total int, entry UnmappedEntry) {
	mediaLabel := capitalizeFirst(entry.MediaType)
	switch {
	case entry.AniListID > 0:
		log.Printf("\n[%d/%d] %q (AniList: %d, %s)", num, total, entry.Title, entry.AniListID, mediaLabel)
	case entry.MALID > 0:
		log.Printf("\n[%d/%d] %q (MAL: %d, %s)", num, total, entry.Title, entry.MALID, mediaLabel)
	default:
		log.Printf("\n[%d/%d] %q (%s)", num, total, entry.Title, mediaLabel)
	}
}

func promptFixAction(reader *bufio.Reader) string {
	log.Print("\nAction: [i]gnore  [m]ap to MAL ID  [s]kip  [q]uit\n> ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "s"
	}
	return strings.TrimSpace(strings.ToLower(input))
}

func applyFixAction(action string, entry UnmappedEntry, mappings *MappingsConfig, reader *bufio.Reader) bool {
	switch action {
	case "i":
		if entry.AniListID > 0 {
			mappings.AddIgnoreByID(entry.AniListID, entry.Title, entry.Reason)
			log.Printf("  -> Added AniList ID %d to ignore list", entry.AniListID)
			return true
		}
		log.Println("  -> Cannot ignore: no AniList ID available")
	case "m":
		malID, ok := promptMALID(reader)
		if ok && entry.AniListID > 0 {
			mappings.AddManualMapping(entry.AniListID, malID, entry.Title)
			log.Printf("  -> Mapped AniList %d -> MAL %d", entry.AniListID, malID)
			return true
		}
	case "q":
		log.Println("Quitting...")
	default:
		log.Println("  -> Skipped")
	}
	return false
}

func promptMALID(reader *bufio.Reader) (int, bool) {
	log.Print("  Enter MAL ID: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, false
	}
	input = strings.TrimSpace(input)
	id, err := strconv.Atoi(input)
	if err != nil || id <= 0 {
		log.Println("  Invalid MAL ID")
		return 0, false
	}
	return id, true
}
