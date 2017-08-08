package sounds

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	wordBoundary = `(?:\s|^|$)`
)

// catMatcher matches a category between curly braces. Category names must
// start with a letter, and may not contain whitespace or '}'
var catMatcher = regexp.MustCompile(`\{(?:(\d+):)?([\p{L}[^}\s]*)\}`)

// parenMatcher matches a single parenthesis and the following character
var parenMatcher = regexp.MustCompile(`\(.`)

func parenReplacer(match string) string {
	if match[1] == '?' {
		return match
	}
	return fmt.Sprintf("(?:%s", match)
}

type CompiledRule struct {
	From                             *compiledPattern
	To                               string
	Before, After, UnBefore, UnAfter *compiledPattern
}

func (cr *CompiledRule) Equal(other *CompiledRule) bool {
	if !cr.From.Equal(other.From) {
		return false
	}
	if cr.To != other.To {
		return false
	}
	if !cr.Before.Equal(other.Before) {
		return false
	}
	if !cr.After.Equal(other.After) {
		return false
	}
	if !cr.UnBefore.Equal(other.UnBefore) {
		return false
	}
	if !cr.UnAfter.Equal(other.UnAfter) {
		return false
	}
	return true
}

// A compiledPattern stores a regular expression and a mapping from the
// capturing groups of that regexp to numbered categories
type compiledPattern struct {
	*regexp.Regexp
	nc         []numCat
	categories CategoryList
}

func (cp *compiledPattern) Equal(other *compiledPattern) bool {
	if cp == nil && other != nil || cp != nil && other == nil {
		return false
	}
	if cp == nil && other == nil {
		return true
	}
	if cp.Regexp.String() != other.Regexp.String() {
		return false
	}
	if len(cp.nc) != len(other.nc) {
		return false
	}
	for i, nc := range cp.nc {
		if nc != other.nc[i] {
			return false
		}
	}
	return cp.categories.Equal(other.categories)
}

// Compile compiles a rule into a set of regular expressions that can be used
// to find matches
func (r *Rule) Compile(categories CategoryList) (*CompiledRule, error) {
	var from, before, after, unBefore, unAfter *compiledPattern
	from, err := compilePattern(r.From, categories)
	if err != nil {
		return nil, err
	}
	before, err = compilePattern(r.Before+"$", categories)
	if err != nil {
		return nil, err
	}
	after, err = compilePattern("^"+r.After, categories)
	if err != nil {
		return nil, err
	}
	if r.UnBefore != "" {
		unBefore, err = compilePattern(r.UnBefore+"$", categories)
		if err != nil {
			return nil, err
		}
	}
	if r.UnAfter != "" {
		unAfter, err = compilePattern("^"+r.UnAfter, categories)
		if err != nil {
			return nil, err
		}
	}
	return &CompiledRule{
		From:     from,
		To:       r.To,
		Before:   before,
		After:    after,
		UnBefore: unBefore,
		UnAfter:  unAfter,
	}, nil
}

// compilePattern generates a compiledPattern
func compilePattern(pattern string, categories CategoryList) (*compiledPattern, error) {
	// first, make all capturing groups non-capturing
	pattern = parenMatcher.ReplaceAllStringFunc(pattern, parenReplacer)
	// second, replace '#' with wordBoundary
	pattern = strings.Replace(pattern, "#", wordBoundary, -1)
	// third, replace categories with regular expressions
	pattern, nc, err := categories.categoryReplace(pattern)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &compiledPattern{Regexp: re, nc: nc, categories: categories}, nil
}

type numCat struct {
	num int
	cat *Category
}

func (nc numCat) Equal(other numCat) bool {
	return nc.num == other.num && nc.cat.Equal(other.cat)
}

// categoryReplace replaces all categories in a pattern with regular
// expressions that will match that category (and in the case of numbered
// categories, capture it). It also returns a list of the numbers and
// categories corresponding to each capturing group.
func (cl CategoryList) categoryReplace(pattern string) (string, []numCat, error) {
	var (
		err error
		nc  []numCat
	)
	replacer := func(match string) string {
		if err != nil {
			// if there's already an error, don't bother
			return ""
		}
		groups := catMatcher.FindStringSubmatch(match)
		cat, ok := cl[groups[2]]
		if !ok {
			err = fmt.Errorf("parse error: category %#v is not defined", groups[2])
			return ""
		}
		pat := cat.Pattern()
		if len(groups[1]) > 0 {
			// numbered group, so it should be capturing
			n, err_ := strconv.Atoi(groups[1])
			if err_ != nil {
				err = err_
				return ""
			}
			nc = append(nc, numCat{num: n, cat: cat})
			return fmt.Sprintf("(%s)", pat)
		}
		// non-capturing group
		return fmt.Sprintf("(?:%s)", pat)
	}
	return catMatcher.ReplaceAllStringFunc(pattern, replacer), nc, err
}
