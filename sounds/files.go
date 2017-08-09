package sounds

import (
	"bufio"
	"os"
)

// LoadFile loads a sound change file as a RuleList
func LoadFile(filename string) (*RuleList, error) {
	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	rl := NewRuleList()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		err = rl.ParseRuleCat(string(scanner.Text()))
		if err != nil {
			return nil, err
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return rl, nil
}

// ApplyFile loads a file and applies it to a word
func ApplyFile(filename, word string) (output string, debug []string, err error) {
	rl, err := LoadFile(filename)
	if err != nil {
		return "", nil, err
	}
	output, debug, err = rl.Apply(word)
	if err != nil {
		return "", debug, err
	}
	return output, debug, err
}
