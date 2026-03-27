package fax

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CheckCallFileExists(spoolPath, callFileName string) bool {
	_, err := os.Stat(filepath.Join(spoolPath, callFileName))
	return err == nil
}

// FaxResult represents what we parsed from the Asterisk log for a fax job.
type FaxResult struct {
	Status string // "sent" or "failed"
	Error  string
}

// maxTailBytes is how far back from the end of the log to read (2MB).
// Fax sessions are short, so the result will be in recent log entries.
const maxTailBytes = 2 * 1024 * 1024

// CheckFaxResult reads the tail of the Asterisk log for the outcome of a fax job
// by searching for the TIFF file path. Returns nil if no result found yet.
func CheckFaxResult(logPath, tiffPath string) *FaxResult {
	if logPath == "" || tiffPath == "" {
		return nil
	}

	f, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	// Seek to tail of file to avoid scanning gigabytes
	info, err := f.Stat()
	if err != nil {
		return nil
	}
	offset := info.Size() - maxTailBytes
	if offset < 0 {
		offset = 0
	}
	f.Seek(offset, io.SeekStart)

	data, err := io.ReadAll(f)
	if err != nil {
		return nil
	}
	tail := string(data)

	// Find the TIFF path in the log to confirm this fax was attempted
	if !strings.Contains(tail, tiffPath) {
		return nil
	}

	// Split into lines and scan for result
	lines := strings.Split(tail, "\n")

	var foundOurFax bool
	var lastStatus string
	var callFinished bool

	for _, line := range lines {
		if strings.Contains(line, tiffPath) {
			foundOurFax = true

			if strings.Contains(line, "Cannot open source TIFF") {
				return &FaxResult{Status: "failed", Error: "TIFF file could not be opened"}
			}
		}

		if foundOurFax {
			if strings.Contains(line, "Status changing to") {
				if idx := strings.Index(line, "Status changing to '"); idx != -1 {
					rest := line[idx+len("Status changing to '"):]
					if end := strings.Index(rest, "'"); end != -1 {
						lastStatus = rest[:end]
					}
				}
			}

			if strings.Contains(line, "T30_PHASE_CALL_FINISHED") {
				callFinished = true
			}

			if callFinished {
				if lastStatus == "" || strings.Contains(strings.ToLower(lastStatus), "call completed") {
					return &FaxResult{Status: "sent"}
				}
				return &FaxResult{Status: "failed", Error: lastStatus}
			}
		}
	}

	return nil
}
