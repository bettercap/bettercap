package util

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/mgutz/str"
)

// PackageName determines the package name from sourceFile if it is within $GOPATH
func PackageName(sourceFile string) (string, error) {
	if filepath.Ext(sourceFile) != ".go" {
		return "", errors.New("sourcefile must end with .go")
	}
	sourceFile, err := filepath.Abs(sourceFile)
	if err != nil {
		Panic("util", "Could not convert to absolute path: %s", sourceFile)
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", errors.New("Environment variable GOPATH is not set")
	}
	paths := strings.Split(gopath, string(os.PathListSeparator))
	for _, path := range paths {
		srcDir := filepath.Join(path, "src")
		srcDir, err := filepath.Abs(srcDir)
		if err != nil {
			continue
		}

		//log.Printf("srcDir %s sourceFile %s\n", srcDir, sourceFile)
		rel, err := filepath.Rel(srcDir, sourceFile)
		if err != nil {
			continue
		}
		return filepath.Dir(rel), nil
	}
	return "", errors.New("sourceFile not reachable from GOPATH")
}

// Template reads a go template and writes it to dist given data.
func Template(src string, dest string, data map[string]interface{}) {
	content, err := ioutil.ReadFile(src)
	if err != nil {
		Panic("template", "Could not read file %s\n%v\n", src, err)
	}

	tpl := template.New("t")
	tpl, err = tpl.Parse(string(content))
	if err != nil {
		Panic("template", "Could not parse template %s\n%v\n", src, err)
	}

	f, err := os.Create(dest)
	if err != nil {
		Panic("template", "Could not create file for writing %s\n%v\n", dest, err)
	}
	defer f.Close()
	err = tpl.Execute(f, data)
	if err != nil {
		Panic("template", "Could not execute template %s\n%v\n", src, err)
	}
}

// StrTemplate reads a go template and writes it to dist given data.
func StrTemplate(src string, data map[string]interface{}) (string, error) {
	tpl := template.New("t")
	tpl, err := tpl.Parse(src)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// PartitionKV partitions a reader then parses key-value meta using an assignment string.
//
// Example
//
// PartitionKV(buf.NewBufferString(`
//   --@ key=SelectUser
//   SELECT * FROM users;
// `, "--@", "=") => [{"_kind": "key", "key": "SelectUser", "_body": "SELECT * FROM users;"}]
func PartitionKV(r io.Reader, prefix string, assignment string) ([]map[string]string, error) {
	scanner := bufio.NewScanner(r)
	var buf bytes.Buffer
	var kv string
	var text string
	var result []map[string]string
	collect := false

	parseKV := func(kv string) {
		argv := str.ToArgv(kv)
		body := buf.String()
		for i, arg := range argv {
			m := map[string]string{}
			var key string
			var value string
			if strings.Contains(arg, assignment) {
				parts := strings.Split(arg, assignment)
				key = parts[0]
				value = parts[1]
			} else {
				key = arg
				value = ""
			}
			m[key] = value
			m["_body"] = body
			if i == 0 {
				m["_kind"] = key
			}
			result = append(result, m)
		}
	}

	for scanner.Scan() {
		text = scanner.Text()
		if strings.HasPrefix(text, prefix) {
			if kv != "" {
				parseKV(kv)
			}
			kv = text[len(prefix):]
			collect = true
			buf.Reset()
			continue
		}
		if collect {
			buf.WriteString(text)
			buf.WriteRune('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if kv != "" && buf.Len() > 0 {
		parseKV(kv)
	}

	if collect {
		return result, nil
	}

	return nil, nil
}
