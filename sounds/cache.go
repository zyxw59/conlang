package sounds

import (
	"os"
	"time"
)

type Cache struct {
	files map[string]cachedFile
}

type cachedFile struct {
	modTime time.Time
	name    string
	rl      *RuleList
}

func NewCache() *Cache {
	return &Cache{
		files: make(map[string]cachedFile),
	}
}

// LoadFile loads a file and caches its contents, or returns the cached
// contents if they are as new as the file
func (c *Cache) LoadFile(filename string) (rl *RuleList, err error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if cf, ok := c.files[filename]; ok {
		if !cf.modTime.Before(info.ModTime()) {
			// cached RuleList is not older than the file, it's
			// good enough
			return cf.rl, nil
		}
	}
	// here, either the cached RuleList is older than the file, or it
	// doesn't exist
	rl, err = LoadFile(filename)
	if err != nil {
		return nil, err
	}
	c.files[filename] = cachedFile{
		modTime: info.ModTime(),
		name:    filename,
		rl:      rl,
	}
	return rl, nil
}

// LoadFiles loads multiple files and caches their contents, or returns the
// cached contents if they are as new as the relevant file
func (c *Cache) LoadFiles(files ...string) (rls []*RuleList, err error) {
	rls = make([]*RuleList, len(files))
	for i, f := range files {
		rls[i], err = c.LoadFile(f)
		if err != nil {
			return nil, err
		}
	}
	return rls, nil
}

// ApplyFile applies a sound change file to a word
func (c *Cache) ApplyFile(word, filename string) (output string, debug []string, err error) {
	rl, err := c.LoadFile(filename)
	if err != nil {
		return "", nil, err
	}
	return rl.Apply(word)
}

// ApplyFiles applies a series of files to a word
func (c *Cache) ApplyFiles(word string, files ...string) (output string, debug []string, err error) {
	debugs := make([][]string, len(files))
	output = word
	for i, f := range files {
		var db []string
		output, db, err = c.ApplyFile(output, f)
		if err != nil {
			return "", debug, err
		}
		debugs[i] = make([]string, 1, len(db)+1)
		debugs[i][0] = f
		debugs[i] = append(debugs[i], db...)
	}
	return output, stringSliceConcat(debugs...), nil
}

// LoadPairs loads multiple files and caches their contents, using a prefix for
// all filenames. It returns cached content if the cache is as recent as the
// files
func (c *Cache) LoadPairs(prefix string, names ...string) ([]*RuleList, error) {
	pairs, err := Pairs(names...)
	if err != nil {
		return nil, err
	}
	files := prefixSlice(pairs, prefix)
	return c.LoadFiles(files...)
}

// ApplyPairs applies a series of sound changes to a word, using a prefix for
// all filenames
func (c *Cache) ApplyPairs(word, prefix string, names ...string) (string, []string, error) {
	pairs, err := Pairs(names...)
	if err != nil {
		return "", nil, err
	}
	files := prefixSlice(pairs, prefix)
	return c.ApplyFiles(word, files...)
}
