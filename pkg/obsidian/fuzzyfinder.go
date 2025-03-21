package obsidian

import (
	"errors"
	"strings"
	"github.com/ktr0731/go-fuzzyfinder"
)

type FuzzyFinder struct{}

type FuzzyFinderManager interface {
	Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error)
}

func (f *FuzzyFinder) Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error) {
	items, ok := slice.([]string)
	if !ok {
		return -1, errors.New("invalid slice type, expected []string")
	}

	index, err := fuzzyfinder.Find(items, func(i int) string {
		return itemFunc(i)
	})
	if err != nil {
		return -1, errors.New(NoteDoesNotExistError)
	}
	return index, nil
}

// FuzzyMatch performs simplified fuzzy matching for file paths
// Returns true if pattern is found within the file path (case insensitive)
func FuzzyMatch(pattern, path string) bool {
	if pattern == "" {
		return false
	}
	
	// Case insensitive matching
	lowerPattern := strings.ToLower(pattern)
	lowerPath := strings.ToLower(path)
	
	// Simple contains check - a more sophisticated algorithm could be used 
	// for better fuzzy matching, but this works for basic search needs
	return strings.Contains(lowerPath, lowerPattern)
}
