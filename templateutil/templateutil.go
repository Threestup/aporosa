package templateutil

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"text/template"
)

var (
	// TemplatesMessages path names to template
	TemplatesMessages = map[string]*template.Template{}
)

// LoadFromDir load templates from a directory
func LoadFromDir(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	// for all file if they are not dirs, and finish with a .tpl extension
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tpl") {
			name := "/" + strings.TrimSuffix(file.Name(), ".tpl")

			// read file
			b, err := ioutil.ReadFile(path.Join(dir, file.Name()))
			if err != nil {
				return err
			}
			tpl, err := template.New(name).Parse(string(b))
			if err != nil {
				return fmt.Errorf("unable to parse message template: %v", err)
			}

			TemplatesMessages[name] = tpl

		} else {
			fmt.Printf("ignored file or directory: %v", file.Name())
		}
	}

	return nil
}
