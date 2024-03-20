package versioning

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type PackageJSON struct {
	Version         string `json:"version"`
	PreviousVersion string `json:"previous_version"`
}

func ReadFile(filename string) (*PackageJSON, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var packageData PackageJSON
	err = decoder.Decode(&packageData)
	if err != nil {
		return nil, err
	}

	return &packageData, nil
}

func UpdateVersion(packageData *PackageJSON, updateType string) error {
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

func WriteFile(filename string, packageData *PackageJSON) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(packageData)
	if err != nil {
		return err
	}
	return nil
}
