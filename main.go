package main

import (
	"bytes"
	"fmt"
	"github.com/ledongthuc/pdf"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	var excelFile string
	fmt.Println("Первая колонка должна быть ИИН")
	fmt.Print("Напишите название вашего эксель файла (.xlsx): ")
	fmt.Scanln(&excelFile)

	e := readExcel(excelFile)
	_ = e
	fmt.Println("Успешно!")
	d := processDirectory(".")
	_ = d

	outputDir := "output"
	fmt.Println("Очищаем и создаем папку:", outputDir)
	os.RemoveAll(outputDir)
	os.MkdirAll(outputDir, os.ModePerm)

	fmt.Println("Копируем...")
	countCopied := 0
	for _, eRow := range e {
		if len(d[eRow]) > 0 {
			for _, path := range d[eRow] {
				//err := os.Remove(path)

				destPath := filepath.Join(outputDir, filepath.Base(path))

				// Copy the file
				err := copyFile(path, destPath)
				if err == nil {
					countCopied++
				}
			}
		}
	}

	fmt.Println("Готово! Скопировано:", countCopied)
	fmt.Scanln()
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	err = destinationFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync destination file: %v", err)
	}

	return nil
}

func processDirectory(path string) map[string][]string {
	r := make(map[string][]string)

	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == "output" {
				fmt.Println("Пропустил парсинг:", path)
				return filepath.SkipDir
			}
			fmt.Println("Парсинг:", path)
			return nil
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".pdf") {
			content, err := readPdf(path)
			if err != nil {
				return nil
			}
			iin, err := regExIIN(content)
			if err != nil {
				return nil
			}
			r[iin] = append(r[iin], path)
		}

		return nil
	})
	return r
}

func regExIIN(content string) (string, error) {
	pattern := `ИИН\s*(\d{12})`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(content)

	if len(match) > 1 {
		return match[1], nil
	}

	return "", fmt.Errorf("IIN ne naiden")
}

func readPdf(path string) (string, error) {
	f, r, err := pdf.Open(path)
	defer f.Close()
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	return buf.String(), nil
}

func readExcel(path string) []string {
	var r []string
	f, _ := excelize.OpenFile(path)
	rows, _ := f.GetRows(f.GetSheetList()[0])
	for _, row := range rows {
		if len(row) > 0 {
			r = append(r, row[0]) // Get first column (index 0)
		}
	}
	f.Close()
	return r
}
