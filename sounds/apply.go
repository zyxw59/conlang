package sounds

import (
	"strings"
)

type Match struct {
	Start   int
	End     int
	Indices map[int]int
}

func (m Match) Equal(other Match) bool {
	if m.Start != other.Start || m.End != other.End {
		return false
	}
	for k, v := range m.Indices {
		vv, ok := other.Indices[k]
		if !ok {
			return false
		}
		if v != vv {
			return false
		}
	}
	for k, v := range other.Indices {
		vv, ok := m.Indices[k]
		if !ok {
			return false
		}
		if v != vv {
			return false
		}
	}
	return true
}

// Apply applies the rule to the string, and returns its new value
func (cr *CompiledRule) Apply(word string) (string, error) {
	// first, get matches:
	matches := cr.FindMatches(word)
	parts := make([]string, 2*len(matches)+1)
	parts[0] = word[:matches[0].Start]
	for i, m := range matches {
		repl, err := cr.Categories.Replace(cr.To, m.Indices)
		if err != nil {
			return "", err
		}
		parts[2*i+1] = repl
		if i == len(matches)-1 {
			parts[2*i+2] = word[m.End:]
		} else {
			parts[2*i+2] = word[m.End:matches[i+1].Start]
		}
	}
	return strings.Join(parts, ""), nil
}

// FindMatches finds and returns a list of all valid matches of the rule in the
// word
func (cr *CompiledRule) FindMatches(word string) []Match {
	// First, match on the From field
	initialMatches := cr.From.FindAllStringIndex(word, -1)
	// initialize final matches array to have length zero, but enough
	// capacity to fit all initial matches
	finalMatches := make([]Match, 0, len(initialMatches))
	// now, check each match for validity
	for _, m := range initialMatches {
		indices := cr.From.categoryMatch(word[m[0]:m[1]], nil)
		// If the match fails to match numbered categories, discard
		if indices == nil {
			continue
		}
		// Search up to the initial match. The Before pattern will
		// always end with `$`, so it must match the end of the string,
		// i.e., right before the initial match
		indices = cr.Before.categoryMatch(word[:m[0]], indices)
		if indices == nil {
			continue
		}
		// Search starting at the end of the initial match. The After
		// pattern will always start with `^`, so it must match the
		// begining of the string, i.e., right after the initial match
		indices = cr.After.categoryMatch(word[m[1]:], indices)
		if indices == nil {
			continue
		}
		// If UnBefore matches, discard
		if cr.UnBefore.categoryMatch(word[:m[0]], indices) != nil {
			continue
		}
		// If UnAfter matches, discard
		if cr.UnAfter.categoryMatch(word[m[1]:], indices) != nil {
			continue
		}
		// if we've made it this far, we've got a match
		finalMatches = append(finalMatches, Match{
			Start:   m[0],
			End:     m[1],
			Indices: indices,
		})
	}
	return finalMatches
}

// categoryMatch checks whether a string matches a compiledPattern, and if it
// does, returns a map of the indices corresponding to each numbered category,
// for the first match in the string. If the string is not a match, return nil.
// For instance, if a pattern `{0:C}` matched the third element of category
// `C`, this function would return map[int]int{0: 3}
func (cp *compiledPattern) categoryMatch(word string, indices map[int]int) map[int]int {
	// if this pattern is nil, it can't match anything
	if cp == nil {
		return nil
	}
	// match has type []string
	match := cp.FindStringSubmatch(word)
	idxs := make(map[int]int)
	if indices != nil {
		// make local copy of indices
		for k, v := range indices {
			idxs[k] = v
		}
	}
	// skip the first element of match, which is the whole word
	for i, sm := range match[1:] {
		// cp.nc[i] is the numbered category corresponding to capturing
		// group i
		n := cp.nc[i].num
		// the index of the submatch in the category
		idx := cp.nc[i].cat.indices[sm]
		// check if idx matches previous instances of this number
		prev, ok := idxs[n]
		if ok && prev != idx {
			// the same number matched different indices
			return nil
		}
		// otherwise set the index (if it's already set, we're merely
		// resetting it with the same value, so no need to check ok)
		idxs[n] = idx
	}
	// assuming we made it out of the loop alive, idxs is fully
	// populated from the match (and any previous indices)
	return idxs
}
