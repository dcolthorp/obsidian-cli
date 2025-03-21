package actions

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// InfoParams contains parameters for the info command
type InfoParams struct {
	FilePath    string // Path to the file
	IncludeTags bool   // Whether to include tags in output
	IncludeLinks bool  // Whether to include linked notes in output
	VaultPath   string // Path to the vault
}

// Variable for mocking os.Stat in tests
var osStat = os.Stat

// GetNoteInfo retrieves and formats information about an Obsidian note
func GetNoteInfo(vault obsidian.VaultManager, note obsidian.NoteManager, params InfoParams) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	
	// Resolve the full file path within the vault
	filePath := params.FilePath
	absolutePath := filepath.Join(params.VaultPath, filePath)
	
	// Check if file exists
	fileInfo, err := osStat(absolutePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}
	
	// Get basic file information
	result["path"] = filePath
	result["absolute_path"] = absolutePath
	result["size"] = fileInfo.Size()
	result["size_human"] = humanizeSize(fileInfo.Size())
	result["modified"] = fileInfo.ModTime().Format(time.RFC3339)
	result["modified_human"] = fileInfo.ModTime().Format("Jan 02, 2006 15:04:05")
	
	// Get file content
	content, err := note.GetContents(params.VaultPath, filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %s", err)
	}
	
	// Extract frontmatter
	frontmatter, err := obsidian.ExtractFrontmatter(content)
	if err != nil {
		// Just log the error and continue
		if Debug {
			fmt.Fprintf(os.Stderr, "Warning: Error parsing frontmatter: %v\n", err)
		}
	}
	
	// Title - Try to get from frontmatter first, then filename
	var title string
	if frontmatter != nil {
		if t, ok := frontmatter["title"].(string); ok && t != "" {
			title = t
		}
	}
	if title == "" {
		// Use filename without extension
		title = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}
	result["title"] = title
	
	// Add frontmatter data
	if frontmatter != nil {
		result["frontmatter"] = frontmatter
	}
	
	// Word count and reading time
	wordCount := countWords(content)
	result["word_count"] = wordCount
	result["reading_time_minutes"] = calculateReadingTime(wordCount)
	
	// Get tags if requested
	if params.IncludeTags {
		// Get tags from frontmatter
		var frontmatterTags []string
		if frontmatter != nil {
			if tags, ok := frontmatter["tags"]; ok {
				switch t := tags.(type) {
				case []string:
					frontmatterTags = t
				case []interface{}:
					for _, tag := range t {
						if tagStr, ok := tag.(string); ok {
							frontmatterTags = append(frontmatterTags, tagStr)
						}
					}
				}
			}
		}
		
		// Extract inline hashtags
		inlineTags := obsidian.ExtractHashtags(content)
		// Clean hashtags (remove the # prefix)
		for i, tag := range inlineTags {
			inlineTags[i] = strings.TrimPrefix(tag, "#")
			inlineTags[i] = strings.TrimSpace(inlineTags[i])
		}
		
		// Combine and deduplicate tags
		allTags := uniqueTags(append(frontmatterTags, inlineTags...))
		result["tags"] = allTags
	}
	
	// Get wikilinks if requested
	if params.IncludeLinks {
		links, backlinks, err := getLinksInfo(note, params.VaultPath, filePath, content)
		if err == nil {
			result["links"] = links
			result["backlinks"] = backlinks
		} else if Debug {
			fmt.Fprintf(os.Stderr, "Warning: Error getting links: %v\n", err)
		}
	}
	
	return result, nil
}

// humanizeSize converts bytes to a human-readable string
func humanizeSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// countWords counts the number of words in a string
func countWords(s string) int {
	// Remove frontmatter
	s = obsidian.RemoveFrontmatter(s)
	
	// Remove code blocks
	s = obsidian.ExtractNonCodeContent(s)
	
	// Count words
	return len(strings.Fields(s))
}

// calculateReadingTime estimates reading time in minutes
func calculateReadingTime(wordCount int) float64 {
	// Average reading speed: 200-250 words per minute
	const wordsPerMinute = 200.0
	minutes := float64(wordCount) / wordsPerMinute
	// Round to one decimal place
	return roundToDecimal(minutes, 1)
}

// roundToDecimal rounds a float to the specified number of decimal places
func roundToDecimal(val float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Round(val*shift) / shift
}

// uniqueTags returns a slice of unique tags
func uniqueTags(tags []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, tag := range tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag != "" && !seen[normalizedTag] {
			seen[normalizedTag] = true
			result = append(result, tag) // Add the original tag, not the normalized one
		}
	}
	
	return result
}

// getLinksInfo gets information about links in a note
func getLinksInfo(note obsidian.NoteManager, vaultPath, filePath, content string) ([]string, []string, error) {
	// Extract wikilinks from the content
	links := obsidian.ExtractWikilinks(content)
	
	// Get all notes in the vault
	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return nil, nil, err
	}
	
	// Find backlinks - notes that link to this note
	var backlinks []string
	fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	
	for _, notePath := range allNotes {
		if notePath == filePath {
			continue // Skip self
		}
		
		noteContent, err := note.GetContents(vaultPath, notePath)
		if err != nil {
			continue // Skip if can't read
		}
		
		// Check if this note links to our target note
		noteLinks := obsidian.ExtractWikilinks(noteContent)
		for _, link := range noteLinks {
			linkName := strings.TrimSuffix(link, filepath.Ext(link))
			if strings.EqualFold(linkName, fileName) {
				backlinks = append(backlinks, notePath)
				break
			}
		}
	}
	
	return links, backlinks, nil
}