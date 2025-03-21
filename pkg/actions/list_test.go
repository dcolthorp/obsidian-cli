package actions

import (
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/stretchr/testify/assert"
)

func TestListFiles_NoInputs(t *testing.T) {
	expectedNotes := []string{"note1.md", "note2.md", "note3.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{Notes: expectedNotes}

	result, err := ListFiles(mockVault, mockNote, ListParams{})

	assert.NoError(t, err)
	assert.Equal(t, expectedNotes, result)
}

func TestListFiles_FileInput(t *testing.T) {
	allNotes := []string{"note1.md", "folder/note2.md", "folder/note3.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{Notes: allNotes}

	// Test matching file path
	inputs := []ListInput{{
		Type:  InputTypeFile,
		Value: "folder",
	}}

	result, err := ListFiles(mockVault, mockNote, ListParams{Inputs: inputs})

	assert.NoError(t, err)
	assert.Equal(t, []string{"folder/note2.md", "folder/note3.md"}, result)
}

func TestListFiles_TagInput(t *testing.T) {
	allNotes := []string{"note1.md", "note2.md", "note3.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{
		Notes: allNotes,
		Contents: map[string]string{
			"note1.md": "---\ntags: [test]\n---\nContent",
			"note2.md": "Content with #test tag",
			"note3.md": "Just content",
		},
	}

	// Test matching tags
	inputs := []ListInput{{
		Type:  InputTypeTag,
		Value: "test",
	}}

	result, err := ListFiles(mockVault, mockNote, ListParams{Inputs: inputs})

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"note1.md", "note2.md"}, result)
}

func TestListFiles_FindInput(t *testing.T) {
	allNotes := []string{"note1.md", "test-note.md", "folder/test-file.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{Notes: allNotes}

	// Test fuzzy finding
	inputs := []ListInput{{
		Type:  InputTypeFind,
		Value: "test",
	}}

	result, err := ListFiles(mockVault, mockNote, ListParams{Inputs: inputs})

	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"test-note.md", "folder/test-file.md"}, result)
}

func TestListFiles_MultipleInputs(t *testing.T) {
	allNotes := []string{"note1.md", "folder/test-note.md", "folder/other.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{
		Notes: allNotes,
		Contents: map[string]string{
			"note1.md":          "Content with #important tag",
			"folder/test-note.md": "---\ntags: [test, important]\n---\nContent",
			"folder/other.md":   "Just content",
		},
	}

	// Combined file path and tag inputs
	inputs := []ListInput{
		{
			Type:  InputTypeFile,
			Value: "folder",
		},
		{
			Type:  InputTypeTag,
			Value: "important",
		},
	}

	result, err := ListFiles(mockVault, mockNote, ListParams{Inputs: inputs})

	assert.NoError(t, err)
	// Should find the intersection: only folder items with the important tag
	assert.ElementsMatch(t, []string{"folder/test-note.md", "note1.md"}, result)
}

func TestListFiles_OnMatch(t *testing.T) {
	allNotes := []string{"note1.md", "note2.md"}
	mockVault := &mocks.MockVault{PathVal: "/path/to/vault"}
	mockNote := &mocks.MockNote{Notes: allNotes}

	var matchedFiles []string
	onMatch := func(file string) {
		matchedFiles = append(matchedFiles, file)
	}

	_, err := ListFiles(mockVault, mockNote, ListParams{OnMatch: onMatch})

	assert.NoError(t, err)
	assert.Equal(t, allNotes, matchedFiles)
}
