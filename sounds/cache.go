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

func (c *Cache) ApplyFile(word, filename string) (output string, debug []string, err error) {
	rl, err := c.LoadFile(filename)
	if err != nil {
		return "", nil, err
	}
	return rl.Apply(word)
}

func (c *Cache) ApplyFiles(word string, files ...string) (output string, debug []string, err error) {
	debugs := make([][]string, len(files))
	output = word
	for i, f := range files {
		output, debugs[i], err = c.ApplyFile(output, f)
		if err != nil {
			return "", debug, err
		}
	}
	return output, stringSliceConcat(debugs...), nil
}
