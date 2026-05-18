package vault

import (
	"os"
	"path/filepath"
	"strings"
)

// Walk returns all .md files under root, skipping hidden dirs and any dir
// whose name matches one of the skipDirs entries (e.g. the attachment folder).
func Walk(root string, skipDirs ...string) ([]string, error) {
	skip := make(map[string]bool, len(skipDirs))
	for _, d := range skipDirs {
		if d != "" {
			skip[d] = true
		}
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || skip[name] {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
