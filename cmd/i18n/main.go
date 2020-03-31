// i18n is a command line utility that assists in maintaining code bases that need to serve
// internationalization and localization purposes. It allows for the quick conversion between typical formats like .csv
// to application resource files (e.g .properties) and vice versa for message files found commonly in i18n applications.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var input = flag.String("i", "", "input path: either a directory or a file")

func usage() {
	fmt.Fprint(os.Stderr, `Usage: i18n -i ./path/to/my/resources -o ./my/generated/test.csv OR i18n -i ./path/to/test.csv -o /some/dir/path

provided an input path and a output path i18n determines whether or not to generate a .csv 
file or a set of .properties files.
`)
	os.Exit(2)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("i18n: ")
	flag.Usage = usage
	flag.Parse()
	if flag.NFlag() == 0 {
		usage()
	}
	if err := i18n(); err != nil {
		log.Fatal(err)
	}
}

func i18n() error {
	fi, err := os.Stat(*input)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	inputIsDir := fi.IsDir()
	if inputIsDir {
		return generateCsv(*input, cwd)
	}
	return generateResourceFiles(*input, cwd)
}

type messageTable map[string][]string

func (mt messageTable) properties(outs ...io.WriteCloser) error {
	var sb strings.Builder
	for idx, out := range outs {
		for key, ts := range mt {
			if _, err := sb.WriteString(fmt.Sprintf("%s=%s\n", key, ts[idx])); err != nil {
				return err
			}
		}
		if _, err := out.Write([]byte(sb.String())); err != nil {
			return err
		}
		out.Close()
		sb.Reset()
	}
	return nil
}

func (mt messageTable) csv(out io.Writer) error {
	var sb strings.Builder
	for key, translations := range mt {
		ts := strings.Join(translations, ",")
		if _, err := sb.WriteString(fmt.Sprintf("%s,%s\n", key, ts)); err != nil {
			return err
		}
	}
	if _, err := out.Write([]byte(sb.String())); err != nil {
		return err
	}
	return nil
}

func generateCsv(inputDir, outputFile string) error {
	files := make([]string, 0)
	languages := make([]string, 0)
	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".properties") {
			parts := strings.Split(path, "_")
			if len(parts) != 3 {
				return fmt.Errorf("expected 3 parts from split with \"_\" found in the file name: %s, found %d", path, len(parts))
			}
			languages = append(languages, strings.ReplaceAll(strings.Join(parts[1:], "_"), ".properties", ""))
			files = append(files, path)
		}
		return nil
	})
	if len(files) == 0 {
		return fmt.Errorf("no .properties files found in the input directory %s", *input)
	}
	if err != nil {
		return err
	}
	path := filepath.Join(outputFile, "uiMessages.csv")
	if err := writeCsv(languages, files, path); err != nil {
		return err
	}
	return nil
}

func generateResourceFiles(inputFile, outputDir string) error {
	mt := make(messageTable)
	langs, err := readCsv(inputFile, mt)
	if err != nil {
		return err
	}
	var outs []io.WriteCloser
	for _, lang := range langs {
		fileName := filepath.Join(outputDir, fmt.Sprintf("uiMessages_%s.properties", lang))
		fp, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		outs = append(outs, fp)
	}
	if err := mt.properties(outs...); err != nil {
		return err
	}
	return nil
}

func readResource(file string, mt messageTable) error {
	fp, err := os.Open(file)
	defer fp.Close()
	if err != nil {
		return err
	}
	scan := bufio.NewScanner(fp)
	lineNum := 0
	for scan.Scan() {
		lineNum++
		l := scan.Text()
		// if line is empty or there is a comment in the properties file
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		parts := strings.Split(l, "=")
		if len(parts) != 2 {
			return fmt.Errorf("expected 2 parts found %d on line %d", len(parts), lineNum)
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])
		if _, ok := mt[k]; !ok {
			mt[k] = make([]string, 0)
		}
		mt[k] = append(mt[k], v)
	}
	return nil
}

func readCsv(file string, mt messageTable) ([]string, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	scan := bufio.NewScanner(fp)
	scan.Scan()
	header := scan.Text()
	langs := strings.Split(header, ",")[1:]
	for scan.Scan() {
		l := scan.Text()
		if l == "" {
			continue
		}
		parts := strings.Split(l, ",")
		if _, ok := mt[parts[0]]; !ok {
			mt[parts[0]] = make([]string, len(parts)-1)
		}
		mt[parts[0]] = parts[1:]
	}
	return langs, nil
}

func writeCsv(langs, resourceFiles []string, outputFile string) error {
	mt := make(messageTable)
	for _, file := range resourceFiles {
		if err := readResource(file, mt); err != nil {
			return err
		}
	}
	fp, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE, 0755)
	defer fp.Close()
	if err != nil {
		return err
	}
	var sb strings.Builder
	ls := strings.Join(langs, ",")
	// write header
	sb.WriteString(fmt.Sprintf("%s,%s\n", "Key", ls))
	if _, err := fp.WriteString(sb.String()); err != nil {
		return err
	}
	if err := mt.csv(fp); err != nil {
		return err
	}
	return nil
}
