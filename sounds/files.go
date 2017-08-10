package sounds

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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

func stringSliceConcat(slices ...[]string) []string {
	length := 0
	for _, sl := range slices {
		length += len(sl)
	}
	out := make([]string, 0, length)
	for _, sl := range slices {
		out = append(out, sl...)
	}
	return out
}

// Pairs takes a list of pairs of `.`-separated filenames, and returns the
// intermediate steps between the start and end points. The second element of
// each pair should be a relative path, starting with `.`
func Pairs(names ...string) (out []string, err error) {
	if len(names)%2 != 0 {
		// there should be an even number of strings
		return nil, fmt.Errorf("pair error: invalid number of strings")
	}
	for i := 0; i < len(names); i += 2 {
		if !strings.HasPrefix(names[i+1], ".") {
			return nil, fmt.Errorf("pair error: second element %#v"+
				"does not begin with `.`", names[i+1])
		}
		splits := splitAll(names[i+1], ".")[1:]
		if names[i] == "" {
			out = append(out, trimPrefixSlice(splits, ".")...)
		} else {
			out = append(out, prefixSlice(splits, names[i])...)
		}
	}
	return out, nil
}

// splitAll splits a string by a separator, and returns a slice containing the
// first element, the first and second elements, etc, upto the whole string.
// So, for example, splitAll("a,b,c", ",") returns ["a", "a,b", "a,b,c"]
func splitAll(s, sep string) (out []string) {
	sss := strings.Split(s, sep)
	out = make([]string, len(sss))
	for i := range sss {
		out[i] = strings.Join(sss[:i+1], sep)
	}
	return out
}

// prefixSlice prepends a prefix to each element of a slice
func prefixSlice(a []string, pre string) (out []string) {
	out = make([]string, len(a))
	for i, s := range a {
		out[i] = pre + s
	}
	return out
}

// trimPrefixSlice applies strings.TrimPrefix to each element of a slice
func trimPrefixSlice(a []string, pre string) (out []string) {
	out = make([]string, len(a))
	for i, s := range a {
		out[i] = strings.TrimPrefix(s, pre)
	}
	return out
}
