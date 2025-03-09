package rkindex

import (
	"reflect"
	"sort"
	"testing"
)

func TestNewIndex(t *testing.T) {
	// Test empty strings list
	emptyIndex := NewIndex([]string{})
	if len(emptyIndex.strings) != 0 || len(emptyIndex.table) != 0 {
		t.Errorf("Expected empty index, got strings: %d, hashes: %d",
			len(emptyIndex.strings), len(emptyIndex.table))
	}

	// Test with actual strings
	testStrings := []string{"hello", "world", "hello world"}
	idx := NewIndex(testStrings)

	// Check strings are stored
	if !reflect.DeepEqual(idx.strings, testStrings) {
		t.Errorf("Expected strings %v, got %v", testStrings, idx.strings)
	}

	// Check index is built (should have some hash entries)
	if len(idx.table) == 0 {
		t.Error("Expected non-empty hashToStrings map")
	}
}

func makeString(len int) string {
	runes := make([]rune, len)
	for i := 0; i < len; i++ {
		runes[i] = 'a'
	}
	return string(runes)
}

func TestAddToIndex(t *testing.T) {
	idx := &Index{
		table:   make(map[uint32][]string),
		strings: []string{},
	}

	// Test adding a string for a new hash
	hash1 := uint32(12345)
	str1 := "test1"
	idx.updateHash(hash1, str1)

	if strings, exists := idx.table[hash1]; !exists || len(strings) != 1 || strings[0] != str1 {
		t.Errorf("Expected new hash entry with string %s, got %v", str1, strings)
	}

	// Test adding a different string with the same hash
	str2 := "test2"
	idx.updateHash(hash1, str2)

	if strings, exists := idx.table[hash1]; !exists || len(strings) != 2 ||
		strings[0] != str1 || strings[1] != str2 {
		t.Errorf("Expected hash entry with strings %s and %s, got %v", str1, str2, strings)
	}

	// Test adding a duplicate string with the same hash (should not add duplicate)
	idx.updateHash(hash1, str1)

	if strings, exists := idx.table[hash1]; !exists || len(strings) != 2 {
		t.Errorf("Expected hash entry to still have 2 strings, got %v", strings)
	}
}

func TestFind(t *testing.T) {
	cases := []struct {
		name      string
		strings   []string
		substring string
		expected  []string
	}{
		{
			name:      "No matches (short)",
			strings:   []string{"hello", "world", "hi there"},
			substring: "x",
			expected:  []string{},
		},
		{
			name:      "No matches",
			strings:   []string{"hello", "world", "hi there"},
			substring: "xyzxyzxyz",
			expected:  []string{},
		},
		{
			name:      "Empty index",
			strings:   []string{},
			substring: "test",
			expected:  []string{},
		},
		{
			name:      "Substring shorter than n-gram size",
			strings:   []string{"hello", "world", "hi there"},
			substring: "hi",
			expected:  []string{"hi there"},
		},
		{
			name:      "All matches",
			strings:   []string{"hello", "world", "hi there"},
			substring: "",
			expected:  []string{"hello", "world", "hi there"},
		},
		{
			name:      "Multiple matches",
			strings:   []string{"hello world", "world of code", "hello code"},
			substring: "world",
			expected:  []string{"hello world", "world of code"},
		},
		{
			name:      "Substring exactly n-gram size",
			strings:   []string{"abcdef", "xyzabc", "abcxyz"},
			substring: string(make([]byte, n)),
			expected:  []string{},
		},
		{
			name:      "Substring longer than n-gram size",
			strings:   []string{"hello world", "world hello", "hello there world"},
			substring: "hello world",
			expected:  []string{"hello world"},
		},
		{
			name:      "Partial word match",
			strings:   []string{"testing", "est", "test"},
			substring: "tes",
			expected:  []string{"testing", "test"},
		},
		{
			name:      "Partial word match",
			strings:   []string{"testing", "est", "test"},
			substring: "est",
			expected:  []string{"testing", "est", "test"},
		},
		{
			name:      "Partial word match",
			strings:   []string{"testing", "est", "test"},
			substring: "sti",
			expected:  []string{"testing"},
		},
		{
			name:      "Partial word match",
			strings:   []string{"testing", "est", "test"},
			substring: "tin",
			expected:  []string{"testing"},
		},
		{
			name:      "Partial word match",
			strings:   []string{"testing", "est", "test"},
			substring: "ing",
			expected:  []string{"testing"},
		},
		{
			name:      "Multiple n-gram matches that fail",
			strings:   []string{"abcde", "defg"},
			substring: "abcdef",
			expected:  []string{},
		},
		{
			name:      "Case sensitivity",
			strings:   []string{"Hello", "HELLO", "hello"},
			substring: "hello",
			expected:  []string{"hello"},
		},
		{
			name:      "Noncontiguous n-grams",
			strings:   []string{"abcXdefXghi", "XabcXdefX", "defabc"},
			substring: "abcdef",
			expected:  []string{},
		},
		{
			name:      "Noncontiguous n-grams",
			strings:   []string{"xabcdef"},
			substring: "defxabc",
			expected:  []string{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			idx := NewIndex(c.strings)
			result := idx.Find(c.substring)

			// Sort both slices for consistent comparison
			sort.Strings(result)
			sort.Strings(c.expected)

			if !reflect.DeepEqual(result, c.expected) {
				t.Errorf("Expected %v, got %v", c.expected, result)
			}
		})
	}
}

func TestGetStringsByHash(t *testing.T) {
	idx := &Index{
		table:   make(map[uint32][]string),
		strings: []string{},
	}

	// Add some test data
	hash1 := uint32(12345)
	strs1 := []string{"test1", "test2"}
	idx.table[hash1] = strs1

	// Test getting strings for an existing hash
	result := idx.getMatches(hash1)
	if !reflect.DeepEqual(result, strs1) {
		t.Errorf("Expected %v, got %v", strs1, result)
	}

	// Test getting strings for a non-existent hash
	nonExistentHash := uint32(99999)
	result = idx.getMatches(nonExistentHash)
	if len(result) != 0 {
		t.Errorf("Expected empty slice for non-existent hash, got %v", result)
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", true},
		{"hello world", "ello w", true},
		{"hello world", "o w", true},
		{"hello world", "o", true},
		{"hello world", " ", true},
		{"hello world", "", true},
		{"hello world", "worlds", false},
		{"hello", "hello world", false},
		{"hello", "", true},  // Empty substring is contained in any string
		{"", "hello", false}, // Non-empty substring not in empty string
		{"", "", true},       // Empty substring in empty string
	}

	for _, tc := range testCases {
		result := contains(tc.str, tc.substr)
		if result != tc.expected {
			t.Errorf("contains(%q, %q): expected %v, got %v",
				tc.str, tc.substr, tc.expected, result)
		}
	}
}

func TestCalculateHash(t *testing.T) {
	// Test same string with different lengths
	str := "abcdefg"
	hash1 := hash(str[:3])
	hash2 := hash(str[:5])
	if hash1 == hash2 {
		t.Errorf("Expected different hashes for different lengths, got %d and %d", hash1, hash2)
	}

	// Test same hash for same input
	hash3 := hash(str[:3])
	if hash1 != hash3 {
		t.Errorf("Expected same hash for same input, got %d and %d", hash1, hash3)
	}

	// Test different strings with same prefix
	str1 := "abcdef"
	str2 := "abcxyz"
	hashStr1 := hash(str1[:3]) // hash of "abc"
	hashStr2 := hash(str2[:3]) // also hash of "abc"
	if hashStr1 != hashStr2 {
		t.Errorf("Expected same hash for same prefix, got %d and %d", hashStr1, hashStr2)
	}

	// Test full strings should have different hashes
	hashFull1 := hash(str1)
	hashFull2 := hash(str2)
	if hashFull1 == hashFull2 {
		t.Errorf("Expected different hashes for different strings, got %d and %d", hashFull1, hashFull2)
	}
}

// TestEdgeCases tests various edge cases that might not be covered elsewhere
func TestEdgeCases(t *testing.T) {
	// Test with string containing only repeated characters
	repeated := "aaaaa"
	idx := NewIndex([]string{repeated})
	result := idx.Find("aaa")
	if len(result) != 1 || result[0] != repeated {
		t.Errorf("Failed to find substring in repeated characters string")
	}

	// Test with Unicode strings
	unicodeStrings := []string{"こんにちは世界", "你好世界", "안녕하세요 세계"}
	idx = NewIndex(unicodeStrings)
	result = idx.Find("世界")
	if len(result) != 2 ||
		(result[0] != unicodeStrings[0] && result[0] != unicodeStrings[1]) ||
		(result[1] != unicodeStrings[0] && result[1] != unicodeStrings[1]) {
		t.Errorf("Failed to find Unicode substring, got %v", result)
	}

	// Test with very long string
	longString := makeString(1000)
	pattern := "xyzxyz"
	pos := 990
	for i := 0; i < len(pattern); i++ {
		longString = longString[:pos+i] + string(pattern[i]) + longString[pos+i+1:]
	}

	idx = NewIndex([]string{longString})
	result = idx.Find(pattern)
	if len(result) != 1 || result[0] != longString {
		t.Errorf("Failed to find pattern in very long string")
	}
}

// Benchmark the index building
func BenchmarkNewIndex(b *testing.B) {
	testStrings := []string{
		"hello world",
		"goodbye world",
		"hello there",
		"general kenobi",
		"lorem ipsum dolor sit amet",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewIndex(testStrings)
	}
}

// Benchmark the find operation
func BenchmarkFind(b *testing.B) {
	testStrings := []string{
		"hello world",
		"goodbye world",
		"hello there",
		"general kenobi",
		"lorem ipsum dolor sit amet",
	}

	idx := NewIndex(testStrings)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Find("world")
	}
}
