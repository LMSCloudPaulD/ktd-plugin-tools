package archiving

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Archiver struct{}

type packageJSON struct {
	Version string `json:"version"`
}

func (a *Archiver) ArchiveFiles(pattern, archiveDir string) error {
	packageData, err := os.ReadFile("package.json")
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var packageJSON packageJSON
	err = json.Unmarshal(packageData, &packageJSON)
	if err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	err = moveFilesToArchive(pattern, archiveDir, packageJSON.Version)
	if err != nil {
		return fmt.Errorf("failed to archive files: %w", err)
	}

	return nil
}

func extractVersionFromFilename(regexp *regexp.Regexp, fileName string) string {
	match := regexp.FindStringSubmatch(fileName)

	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func compareVersions(versionA, versionB string) int {
	partsA := strings.Split(versionA, ".")
	partsB := strings.Split(versionB, ".")

	for i := range partsA {
		if i >= len(partsB) {
			return 1
		}

		numberA, errA := strconv.Atoi(partsA[i])
		numberB, errB := strconv.Atoi(partsB[i])
		if errA != nil || errB != nil {
			return -1
		}

		if numberA > numberB {
			return 1
		} else if numberA < numberB {
			return -1
		}
	}

	return 0
}

func ensureArchiveDir(archiveDir string) error {
	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		return fmt.Errorf("failed to create archive dir: %w", err)
	}
	return nil
}

func moveFilesToArchive(pattern, archiveDir, version string) error {
	timestamp := time.Now().Format("20060102150405")

	if err := ensureArchiveDir(archiveDir); err != nil {
		return err
	}

	files, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	fileRegex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile pattern: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !fileRegex.MatchString(file.Name()) {
			continue
		}

		fileVersion := extractVersionFromFilename(fileRegex, file.Name())
		if fileVersion == "" || compareVersions(fileVersion, version) >= 0 {
			continue
		}

		sourcePath := filepath.Join(".", file.Name())
		destinationPath := filepath.Join(archiveDir, fmt.Sprintf("%s_%s", timestamp, file.Name()))

		if err := os.Rename(sourcePath, destinationPath); err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", file.Name(), destinationPath, err)
		}

		fmt.Printf("moved %s to %s\n", file.Name(), destinationPath)
	}

	return nil
}
