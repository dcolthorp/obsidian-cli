package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/spf13/cobra"
)

var (
	includeTags  bool
	includeLinks bool
	outputFormat string
)

var infoCmd = &cobra.Command{
	Use:   "info [file]",
	Short: "Display information about an Obsidian note",
	Long: `Display detailed information about an Obsidian note.
Includes metadata, word count, reading time, and optional tag and link information.

Examples:
  obsidian-cli info "Notes/Ideas.md"
  obsidian-cli info --tags --links "Daily Notes/2023-09-01.md"
  obsidian-cli info --format json "Project/README.md"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// If no vault name is provided, get the default vault name
		if vaultName == "" {
			vault := &obsidian.Vault{}
			defaultName, err := vault.DefaultName()
			if err != nil {
				log.Fatal(err)
			}
			vaultName = defaultName
		}

		vault := obsidian.Vault{Name: vaultName}
		note := obsidian.Note{}

		// Get vault path
		vaultPath, err := vault.Path()
		if err != nil {
			log.Fatal(err)
		}

		// Get note info
		info, err := actions.GetNoteInfo(&vault, &note, actions.InfoParams{
			FilePath:    args[0],
			IncludeTags: includeTags,
			IncludeLinks: includeLinks,
			VaultPath:   vaultPath,
		})

		if err != nil {
			log.Fatal(err)
		}

		// Format and print output
		switch outputFormat {
		case "json":
			jsonOutput, err := json.MarshalIndent(info, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(jsonOutput))
		case "text":
			// Simple text output
			fmt.Printf("Title: %s\n", info["title"])
			fmt.Printf("Path: %s\n", info["path"])
			fmt.Printf("Size: %s\n", info["size_human"])
			fmt.Printf("Modified: %s\n", info["modified_human"])
			fmt.Printf("Word count: %d\n", info["word_count"])
			fmt.Printf("Reading time: %.1f minutes\n", info["reading_time_minutes"])

			if frontmatter, ok := info["frontmatter"].(map[string]interface{}); ok && len(frontmatter) > 0 {
				fmt.Println("\nFrontmatter:")
				for k, v := range frontmatter {
					fmt.Printf("  %s: %v\n", k, v)
				}
			}

			if tags, ok := info["tags"].([]string); ok && len(tags) > 0 {
				fmt.Println("\nTags:")
				for _, tag := range tags {
					fmt.Printf("  - %s\n", tag)
				}
			}

			if links, ok := info["links"].([]string); ok && len(links) > 0 {
				fmt.Println("\nLinks:")
				for _, link := range links {
					fmt.Printf("  - %s\n", link)
				}
			}

			if backlinks, ok := info["backlinks"].([]string); ok && len(backlinks) > 0 {
				fmt.Println("\nBacklinks:")
				for _, link := range backlinks {
					fmt.Printf("  - %s\n", link)
				}
			}
		}
	},
}

func init() {
	infoCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	infoCmd.Flags().BoolVarP(&includeTags, "tags", "t", false, "include tags information")
	infoCmd.Flags().BoolVarP(&includeLinks, "links", "l", false, "include wikilinks information")
	infoCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "output format (text, json)")
	rootCmd.AddCommand(infoCmd)
}