package rkindex

import "slices"

const (
	// Length of n-grams used for indexing and searching
	n = 3

	// Prime numbers used by hash
	prime0 uint32 = 5381
	prime1 uint32 = 1566083941
)

// Index is a search index used to quickly perform substring matches.
type Index struct {
	strings []string
	table   map[uint32][]string
}

// NewIndex builds a searchable index from all provided strings.
func NewIndex(strings []string) *Index {
	i := &Index{
		strings: strings,
		table:   make(map[uint32][]string),
	}
	for _, str := range strings {
		for s := str; len(s) >= n; s = s[1:] {
			hash := hash(s[:n])
			i.updateHash(hash, str)
		}
	}
	return i
}

// updateHash adds a string to the index under the given hash.
func (i *Index) updateHash(hash uint32, str string) {
	if strings, ok := i.table[hash]; ok {
		if !slices.Contains(strings, str) {
			i.table[hash] = append(i.table[hash], str)
		}
	} else {
		i.table[hash] = []string{str}
	}
}

// Find searches the index and returns all substring matches.
func (i *Index) Find(substr string) []string {
	if len(substr) == 0 {
		return i.strings
	}
	if len(substr) < n {
		return i.bruteForceSearch(substr)
	}

	var candidates, tmp map[string]bool

	remain := substr
	for {
		ngram := remain[:n]
		hash := hash(ngram)

		matches := i.getMatches(hash)
		if len(matches) == 0 {
			return []string{}
		}

		if candidates == nil {
			candidates = make(map[string]bool, len(matches))
			tmp = make(map[string]bool, len(matches))
			for _, str := range matches {
				candidates[str] = true
			}
		} else {
			for _, str := range matches {
				if candidates[str] {
					tmp[str] = true
				}
			}
			candidates, tmp = tmp, candidates
			clear(tmp)
			if len(candidates) == 0 {
				return []string{}
			}
		}

		remain = remain[n:]
		if len(remain) == 0 {
			break
		}

		// If the remainder is shorter than an n-gram, build the final n-gram
		// from the original substring's last n characters. This gives us some
		// extra filtering power when the length of the substring isn't evenly
		// divisible by n.
		if len(remain) < n {
			remain = substr[len(substr)-n:]
		}
	}

	result := make([]string, 0, len(candidates))
	for str := range candidates {
		if contains(str, substr) {
			result = append(result, str)
		}
	}

	return result
}

// bruteForceSearch performs a direct search through all strings. Used
// for short substring searches.
func (i *Index) bruteForceSearch(substr string) []string {
	result := make([]string, 0)
	for _, str := range i.strings {
		if contains(str, substr) {
			result = append(result, str)
		}
	}
	return result
}

// getMatches returns all strings associated with a hash.
func (i *Index) getMatches(hash uint32) []string {
	if strings, ok := i.table[hash]; ok {
		return strings
	}
	return []string{}
}

// contains checks if a string contains a substring.
func contains(str, substr string) bool {
	ssn := len(substr)
	if ssn > len(str) {
		return false
	}

	for s := str; len(s) >= ssn; s = s[1:] {
		if s[:ssn] == substr {
			return true
		}
	}

	return false
}

// hash computes a string's hash value. It uses an algorithm similar to the
// one used by pre-6.0 .NET.
func hash(str string) uint32 {
	hash1, hash2 := prime0, prime0
	for ; len(str) >= 2; str = str[2:] {
		hash1 = ((hash1 << 5) + hash1) ^ uint32(str[0])
		hash2 = ((hash2 << 5) + hash2) ^ uint32(str[1])
	}
	if len(str) > 0 {
		hash1 = ((hash1 << 5) + hash1) ^ uint32(str[0])
	}
	return hash1 + (hash2 * prime1)
}
