package pypi

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Requirement represents a Python requirement.
type Requirement struct {
	Name          string
	Specification Specifier
}

// Specifier is like this thing maybe?: https://www.python.org/dev/peps/pep-0508/
type Specifier struct {
	Comparison string
	Version    string
}

// ParseRequirements parses requirements
func ParseRequirements(filename string) (reqs []Requirement, err error) {
	f, err := os.Open(filename) // os.OpenFile has more options if you need them
	defer f.Close()
	if err != nil {
		return reqs, err
	}

	rd := bufio.NewReader(f)

	commentRegexp, err := regexp.Compile(`(^|\s+)#.*$`)
	comparisonRegexp, err := regexp.Compile(`(<|<=|!=|==|>=|>|~=|===)`)
	if err != nil {
		return reqs, err
	}
	for {
		line, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return reqs, err
		}
		line = strings.TrimSpace(line)
		line = commentRegexp.ReplaceAllString(line, "")
		if line != "" {
			comparison := comparisonRegexp.FindString(line)
			if comparison != "==" {
				return reqs, fmt.Errorf("version specifier listed (%s) not current supported: %s", comparison, line)
			}
			parts := comparisonRegexp.Split(line, -1)
			fmt.Printf("%q\n", parts)
			reqs = append(reqs, Requirement{
				Name: parts[0],
				Specification: Specifier{
					Comparison: comparison,
					Version: parts[1],
				},
			})
		}
	}
	return reqs, nil
}
