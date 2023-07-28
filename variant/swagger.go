package variant

// copy from magodo/azure-rest-api-index

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// CollectSpecs collects all Swagger specs based on the effective tags in each RP's readme.md.
func CollectSpecs(rootdir string) ([]string, error) {
	var speclist []string

	if err := filepath.WalkDir(rootdir,
		func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.EqualFold(d.Name(), "data-plane") {
					return filepath.SkipDir
				}
				if strings.EqualFold(d.Name(), "examples") {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() != "readme.md" {
				return nil
			}
			content, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("reading file %s: %v", p, err)
			}
			l, err := SpecListFromReadmeMD(content)
			if err != nil {
				return fmt.Errorf("retrieving spec list from %s: %v", p, err)
			}
			for _, relp := range l {
				speclist = append(speclist, filepath.Join(filepath.Dir(p), relp))
			}
			return filepath.SkipDir
		}); err != nil {
		return nil, err
	}
	sort.Slice(speclist, func(i, j int) bool { return speclist[i] < speclist[j] })
	return speclist, nil
}

type TagInfo struct {
	InputFile []string `yaml:"input-file"`
}

func SpecListFromReadmeMD(b []byte) ([]string, error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	specSet := map[string]struct{}{}
	var isEnter bool
	var ymlContent string
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		// Some starting line has empty space between "```" and "yaml $(tag)"
		if strings.HasPrefix(trimmedLine, "```") && strings.HasPrefix(strings.TrimSpace(strings.TrimPrefix(trimmedLine, "```")), "yaml $(tag)") {
			isEnter = true
			continue
		}
		if trimmedLine == "```" {
			var info TagInfo
			if err := yaml.Unmarshal([]byte(ymlContent), &info); err != nil {
				return nil, fmt.Errorf("decoding yaml %q: %v", ymlContent, err)
			}
			for _, p := range info.InputFile {
				p = filepath.Clean(strings.Replace(p, "$(this-folder)", ".", -1))

				// Some poor readme defines the spec path in Windows path format, convert them then..
				if !strings.Contains(p, "/") && strings.Contains(p, `\`) {
					p = strings.Replace(p, `\`, "/", -1)
				}

				specSet[p] = struct{}{}
			}
			// rest the states
			isEnter = false
			ymlContent = ""
			continue
		}
		if !isEnter {
			continue
		}
		ymlContent += line + "\n"
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %v", err)
	}
	var out []string
	for p := range specSet {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i] < out[j]
	})
	return out, nil
}
