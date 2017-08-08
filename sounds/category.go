package sounds

import (
	"fmt"
	"sort"
	"strings"
)

// A CategoryList is a map of strings to categories
type CategoryList map[string]*Category

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

func (c *Category) String() string {
	return strings.Join(c.values, " ")
}

func (c *Category) Pattern() string {
	return strings.Join(c.sorted, "|")
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
