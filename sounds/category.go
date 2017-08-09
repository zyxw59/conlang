package sounds

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// A CategoryList is a map of strings to categories
type CategoryList map[string]*Category

// Replace replaces all instances of a numbered category with the
// appropriate element of that category
func (cl CategoryList) Replace(text string, indices map[int]int) (string, error) {
	var err error
	replacer := func(match string) string {
		if err != nil {
			// if there's already an error, don't bother
			return ""
		}
		groups := catMatcher.FindStringSubmatch(match)
		cat, ok := cl[groups[2]]
		if !ok {
			err = fmt.Errorf("replacement error: category %#v is not defined", groups[2])
			return ""
		}
		if groups[1] == "" {
			// unnumbered category in replacement text, error
			err = fmt.Errorf("replacement error: unnumbered category %#v in replacement text", groups[2])
			return ""
		}
		// numbered category
		n, err_ := strconv.Atoi(groups[1])
		if err_ != nil {
			err = err_
			return ""
		}
		i := indices[n]
		if i >= cat.Len() || i < 0 {
			err = fmt.Errorf("replacement error: invalid index %#v for category %#v", groups[2])
			return ""
		}
		return cat.Get(i)
	}
	return catMatcher.ReplaceAllStringFunc(text, replacer), err
}

// Equal compares two CategoryLists by value
func (cl CategoryList) Equal(other CategoryList) bool {
	if len(cl) != len(other) {
		return false
	}
	for k, v := range cl {
		val, ok := other[k]
		if !ok {
			return false
		}
		if !v.Equal(val) {
			return false
		}
	}
	return true
}

// A Category is a set of sounds
type Category struct {
	values  []string
	sorted  []string
	indices map[string]int
	Name    string
}

// NewCategory returns a category from a []string
func NewCategory(name string, elements []string) (*Category, error) {
	c := &Category{
		values:  elements,
		sorted:  make([]string, len(elements)),
		indices: make(map[string]int),
		Name:    name,
	}
	for i, e := range elements {
		if _, ok := c.indices[e]; ok {
			return nil, fmt.Errorf("parse error: duplicate element %#v in category %#v.", e, c.Name)
		}
		c.sorted[i] = e
		c.indices[e] = i
	}
	sort.Sort(c)
	return c, nil
}

// Equal compares two Categories by value
func (c *Category) Equal(other *Category) bool {
	if c == nil {
		return true
	}
	if c == nil && other != nil || c != nil && other == nil {
		return false
	}
	if c.Name != other.Name {
		return false
	}
	if c.Len() != other.Len() {
		return false
	}
	for i, v := range c.values {
		if v != other.values[i] {
			return false
		}
	}
	return true
}

func (c *Category) Apply(word string) (output, debug string, err error) {
	return word, c.String(), nil
}

// String writes the category as it would appear in a sound change file, as a
// category name, followed by an equals sign, followed by a space separated
// list of its elements
func (c *Category) String() string {
	return fmt.Sprintf("%s = %s", c.Name, strings.Join(c.values, " "))
}

// ElemString writes the elements of the category as a space separated list
func (c *Category) ElemString() string {
	return strings.Join(c.values, " ")
}

// Pattern writes the category as a `|`-separated list, for use in a regular
// expression
func (c *Category) Pattern() string {
	return strings.Join(c.sorted, "|")
}

// Get returns the i-th element of the category
func (c *Category) Get(index int) string {
	return c.values[index]
}

// Len returns the size of the category
func (c *Category) Len() int {
	return len(c.values)
}

func (c *Category) Swap(i, j int) {
	c.sorted[i], c.sorted[j] = c.sorted[j], c.sorted[i]
}

func (c *Category) Less(i, j int) bool {
	return len(c.sorted[i]) > len(c.sorted[j])
}
