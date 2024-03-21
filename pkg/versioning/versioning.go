package versioning

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type packageJSON struct {
	Version         string `json:"version"`
	PreviousVersion string `json:"previous_version"`
}

type VersionManager struct{}

func (v *VersionManager) BumpVersion(filename, updateType string) error {
	packageData, err := readFile(filename)
	if err != nil {
		return fmt.Errorf("error reading package.json: %w", err)
	}

	err = updateVersion(packageData, updateType)
	if err != nil {
		return fmt.Errorf("error bumping version: %w", err)
	}

	err = writeFile(filename, packageData)
	if err != nil {
		return fmt.Errorf("error writing to package.json: %w", err)
	}

	return nil
}

func readFile(filename string) (*packageJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var packageData packageJSON
	err = decoder.Decode(&packageData)
	if err != nil {
		return nil, err
	}

	return &packageData, nil
}

func updateVersion(packageData *packageJSON, updateType string) error {
	versionParts := strings.Split(packageData.Version, ".")
	if len(versionParts) != 3 {
		return fmt.Errorf("invalid version format in package.json")
	}

	major, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return err
	}
	minor, err := strconv.Atoi(versionParts[1])
	if err != nil {
		return err
	}
	patch, err := strconv.Atoi(versionParts[2])
	if err != nil {
		return err
	}

	switch updateType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	default:
		return fmt.Errorf("invalid update type: %s", updateType)
	}

	packageData.PreviousVersion = packageData.Version
	packageData.Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	return nil
}

func writeFile(filename string, packageData *packageJSON) error {
	existingData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var existingJSON map[string]interface{}
	err = json.Unmarshal(existingData, &existingJSON)

	existingJSON["version"] = packageData.Version
	existingJSON["previous_version"] = packageData.PreviousVersion

	newData, err := json.MarshalIndent(existingJSON, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, newData, 0644)
}
