package shodan

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

// SaveIPs ذخیره لیست IP ها در فایل
func SaveIPs(ips []string, path string, append bool) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	flag := os.O_CREATE | os.O_WRONLY
	if append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	f, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, ip := range ips {
		fmt.Fprintln(w, ip)
	}
	return w.Flush()
}

// LoadIPs خواندن لیست IP از فایل
func LoadIPs(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var ips []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" && line[0] != '#' {
			ips = append(ips, line)
		}
	}
	return ips, scanner.Err()
}
