package cli

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SelectType presents supported document types and reads an id or numeric choice.
func SelectType(r io.Reader, w io.Writer, supported []string) (string, error) {
	return SelectTypeFromScanner(bufio.NewScanner(r), w, supported)
}

// SelectTypeFromInput presents supported document types using a shared guided input reader.
func SelectTypeFromInput(input *Input, w io.Writer, supported []string) (string, error) {
	return SelectTypeFromScanner(input.scanner, w, supported)
}

func SelectTypeFromScanner(scanner *bufio.Scanner, w io.Writer, supported []string) (string, error) {
	if len(supported) == 0 {
		return "", fmt.Errorf("no document types are available")
	}
	fmt.Fprintln(w, "Supported document types:")
	for idx, typ := range supported {
		fmt.Fprintf(w, "%d. %s\n", idx+1, typ)
	}
	fmt.Fprint(w, "Choose a document type: ")
	for scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())
		if choice == "" {
			fmt.Fprint(w, "Choose a document type: ")
			continue
		}
		if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(supported) {
			return supported[n-1], nil
		}
		return choice, nil
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("read document type selection: %w", err)
	}
	return "", fmt.Errorf("no document type selected")
}
