package pluginfinder

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type PluginInfo struct {
	FileName string `json:"pm_file_name"`
	FilePath string `json:"pm_file_path"`
}

type KohaPluginMetadata struct {
	PluginModulePath string `json:"pmFilePath"`
	PluginModuleFile string `json:"pmFileName"`
	KohaDirectory    string `json:"kohaDirectory"`
}

type PluginFinder struct{}

func (p *PluginFinder) Find(searchPath string) (*KohaPluginMetadata, error) {
	kohaPluginMetadata, err := getPluginObject(searchPath)
	if err != nil {
		return nil, err
	}

	return kohaPluginMetadata, nil
}

func getExcludeDirs(dirPath string) ([]string, error) {
	ignoreFilePath := filepath.Join(dirPath, ".kohaignore")

	ignoreFile, err := os.Open(ignoreFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	defer ignoreFile.Close()

	fileContents, err := io.ReadAll(ignoreFile)
	if err != nil {
		return nil, err
	}

	var excludedDirs []string
	for _, line := range strings.Split(string(fileContents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		excludedDirs = append(excludedDirs, line)
	}

	return excludedDirs, nil
}

func RecursiveDirSearch(dirPath string, callback func(dirPath, itemName string) (interface{}, error), excludedDirs []string, findAll bool) ([]interface{}, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var results []interface{}

	for _, file := range files {
		if file.IsDir() && isDirExcluded(file.Name(), excludedDirs) {
			continue
		}

		result, err := callback(dirPath, file.Name())
		if err != nil {
			return nil, err
		}

		if result != nil {
			if findAll {
				results = append(results, result)
			} else {
				return []interface{}{result}, nil
			}
		}

		if file.IsDir() {
			subdirResults, err := RecursiveDirSearch(filepath.Join(dirPath, file.Name()), callback, excludedDirs, findAll)
			if err != nil {
				return nil, err
			}

			if subdirResults != nil {
				if findAll {
					results = append(results, subdirResults...)
				} else {
					return subdirResults, nil
				}

			}
		}
	}

	return results, nil
}

func findPluginPath(dirPath string) (*PluginInfo, error) {
	callback := func(dirPath, itemName string) (interface{}, error) {
		if strings.HasSuffix(itemName, ".pm") {
			baseName := strings.TrimSuffix(itemName, ".pm")
			potentialPluginDir := filepath.Join(dirPath, baseName)

			if info, err := os.Stat(potentialPluginDir); err == nil && info.IsDir() {
				return &PluginInfo{
					FileName: itemName,
					FilePath: potentialPluginDir,
				}, nil
			}
		}

		return nil, nil // Continue searching
	}

	results, err := RecursiveDirSearch(dirPath, callback, nil, false) // Find the first match
	if err != nil {
		return nil, fmt.Errorf("no .pm file matching plugin convention found")
	}

	return results[0].(*PluginInfo), nil
}

func findKohaDir(rootPath string, excludedDirs []string) (string, error) {
	callback := func(dirPath, itemName string) (interface{}, error) {
		if itemName == "Koha" {
			return filepath.Join(rootPath, dirPath), nil
		}
		return nil, nil
	}

	results, err := RecursiveDirSearch(rootPath, callback, excludedDirs, false)
	if err != nil {
		return "", err
	}

	if len(results) > 0 {
		return results[0].(string), nil
	}

	return "", fmt.Errorf("koha directory not found")
}

func getPluginObject(searchPath string) (*KohaPluginMetadata, error) {
	startTime := time.Now()

	excludedDirs, err := getExcludeDirs(searchPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get excluded directories: %v", err)
	}

	kohaPath, err := findKohaDir(searchPath, excludedDirs)
	if err != nil {
		return nil, fmt.Errorf("koha directory not found: %v", err)
	}

	pluginInfo, err := findPluginPath(kohaPath)
	if err != nil {
		return nil, fmt.Errorf("no .pm file found: %v", err)
	}

	metadata := &KohaPluginMetadata{
		PluginModuleFile: pluginInfo.FileName,
		PluginModulePath: pluginInfo.FilePath,
		KohaDirectory:    kohaPath,
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime).Seconds() * 1000

	fmt.Printf("Execution time: %.2f ms\n", duration)

	return metadata, nil
}

func isDirExcluded(dirName string, excludedDirs []string) bool {
	for _, excludedDir := range excludedDirs {
		if dirName == excludedDir {
			return true
		}
	}
	return false
}
