package actions

import (
	"os"
	"testing"
	"time"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Simple FileInfo mock for testing
type mockFileInfo struct {
	size     int64
	modTime  time.Time
	isDir    bool
	fileName string
}

func (m *mockFileInfo) Name() string       { return m.fileName }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestGetNoteInfo(t *testing.T) {
	mockVault := new(mocks.VaultManager)
	mockNote := new(mocks.NoteManager)

	testContent := `---
title: Test Note
tags: [test, example]
---

# Test Note

This is a test note with some #tag1 and #tag2.

[[linked-note]]
`

	// Helper function to create a simplified mock os.Stat implementation
	// Save the original os.Stat function
	originalOsStat := osStat
	t.Cleanup(func() {
		osStat = originalOsStat // Restore original function
	})
	
	// Create a mock file info for our test
	mockFileInfo := &mockFileInfo{
		size:     1024,
		modTime:  time.Date(2023, 9, 1, 12, 0, 0, 0, time.UTC),
		isDir:    false,
		fileName: "test.md",
	}
	
	// Replace the os.Stat function with our mock
	osStat = func(name string) (os.FileInfo, error) {
		return mockFileInfo, nil
	}

	t.Run("basic info without tags or links", func(t *testing.T) {
		mockVault.On("Path").Return("/vault/path", nil)
				
		// Set up mocks
		params := InfoParams{
			FilePath:    "test.md",
			IncludeTags: false,
			IncludeLinks: false,
			VaultPath:   "/vault/path",
		}
		
		// Get note contents
		mockNote.On("GetContents", "/vault/path", "test.md").Return(testContent, nil)
		
		info, err := GetNoteInfo(mockVault, mockNote, params)
		
		assert.NoError(t, err)
		assert.Equal(t, "Test Note", info["title"])
		assert.Equal(t, "test.md", info["path"])
		assert.Equal(t, 18, info["word_count"])
		
		// No tags or links included
		_, tagsExist := info["tags"]
		assert.False(t, tagsExist)
		
		_, linksExist := info["links"]
		assert.False(t, linksExist)
		
		mockVault.AssertExpectations(t)
		mockNote.AssertExpectations(t)
	})

	t.Run("with tags", func(t *testing.T) {
		mockVault.On("Path").Return("/vault/path", nil)
		
		// Set up mocks
		params := InfoParams{
			FilePath:    "test.md",
			IncludeTags: true,
			IncludeLinks: false,
			VaultPath:   "/vault/path",
		}
		
		// Get note contents
		mockNote.On("GetContents", "/vault/path", "test.md").Return(testContent, nil)
		
		info, err := GetNoteInfo(mockVault, mockNote, params)
		
		assert.NoError(t, err)
		assert.Equal(t, "Test Note", info["title"])
		
		tags, ok := info["tags"].([]string)
		assert.True(t, ok)
		assert.Contains(t, tags, "test")
		assert.Contains(t, tags, "example")
		assert.Contains(t, tags, "tag1")
		assert.Contains(t, tags, "tag2")
		
		mockVault.AssertExpectations(t)
		mockNote.AssertExpectations(t)
	})

	t.Run("with links", func(t *testing.T) {
		mockVault.On("Path").Return("/vault/path", nil)
		
		// Set up mocks
		params := InfoParams{
			FilePath:    "test.md",
			IncludeTags: false,
			IncludeLinks: true,
			VaultPath:   "/vault/path",
		}
		
		// Get note contents
		mockNote.On("GetContents", "/vault/path", "test.md").Return(testContent, nil)
		
		// Get all notes for backlinks
		mockNote.On("GetNotesList", "/vault/path").Return([]string{"test.md", "other.md"}, nil)
		
		// Get content for checking backlinks
		mockNote.On("GetContents", "/vault/path", "other.md").Return("[[test]]", nil)
		
		info, err := GetNoteInfo(mockVault, mockNote, params)
		
		assert.NoError(t, err)
		
		links, ok := info["links"].([]string)
		assert.True(t, ok)
		assert.Contains(t, links, "linked-note")
		
		backlinks, ok := info["backlinks"].([]string)
		assert.True(t, ok)
		assert.Contains(t, backlinks, "other.md")
		
		mockVault.AssertExpectations(t)
		mockNote.AssertExpectations(t)
	})

	t.Run("humanizeSize works correctly", func(t *testing.T) {
		assert.Equal(t, "100 B", humanizeSize(100))
		assert.Equal(t, "1.0 KB", humanizeSize(1024))
		assert.Equal(t, "1.5 KB", humanizeSize(1536))
		assert.Equal(t, "1.0 MB", humanizeSize(1024*1024))
		assert.Equal(t, "1.0 GB", humanizeSize(1024*1024*1024))
	})

	t.Run("calculateReadingTime works correctly", func(t *testing.T) {
		assert.Equal(t, 0.0, calculateReadingTime(0))
		assert.Equal(t, 1.0, calculateReadingTime(200))
		assert.Equal(t, 2.5, calculateReadingTime(500))
	})
}