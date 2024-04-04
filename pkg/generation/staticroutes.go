package staticroutes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"

	pluginfinder "github.com/LMSCloudPaulD/ktd-plugin-tools/pkg/meta"
)

type StaticRoutesGenerator struct{}

func (s *StaticRoutesGenerator) Generate() error {
	return generateStaticRoutes()
}

func getPathSpecTemplate() (*template.Template, error) {
	tmpl, err := template.New("static").Parse(`
{
  "get": {
    "x-mojo-to": "Static#get",
    "tags": ["pluginStatic"],
    "responses": {
      "200": {
        "description": "File found",
        "schema": {
          "type": "file"
        }
      },
      "400": {
        "description": "Bad request",
        "schema": {
          "type": "object",
          "properties": {
            "error": {
              "description": "An explanation for the error",
              "type": "string"
            }
          }
        }
      },
      "404": {
        "description": "File not found",
        "schema": {
          "type": "object",
          "properties": {
            "error": {
              "description": "An explanation for the error",
              "type": "string"
            }
          }
        }
      },
      "500": {
        "description": "Internal server error",
        "schema": {
          "type": "object",
          "properties": {
            "error": {
              "description": "An explanation for the error",
              "type": "string"
            }
          }
        }
      }
    }
  }
}
`)

	if err != nil {
		return nil, err

	}

	return tmpl, nil
}

func directoryExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return info.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func hasValidSuffix(filename string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(filename, suffix) {
			return true
		}
	}
	return false
}

func generateStaticRoutes() error {
	pluginFinder := &pluginfinder.PluginFinder{}
	pluginMetadata, err := pluginFinder.Find(".")
	if err != nil {
		return err
	}

	staticApiFile := pluginMetadata.PluginModulePath + "/staticapi.json"

	localesPath := pluginMetadata.PluginModulePath + "/locales"
	assetsPath := pluginMetadata.PluginModulePath + "/assets"
	distPath := pluginMetadata.PluginModulePath + "/dist"

	pathSpecsMap := make(map[string]interface{})

	tmpl, err := getPathSpecTemplate()
	if err != nil {
		return err
	}

	var templateBuffer bytes.Buffer
	err = tmpl.Execute(&templateBuffer, nil)
	if err != nil {
		return err
	}

	var unmarshalled map[string]interface{}
	json.Unmarshal(templateBuffer.Bytes(), &unmarshalled)

	localesPathExists, _ := directoryExists(localesPath)
	if localesPathExists {
		fmt.Printf("Processing %v directory...", localesPath)
		callback := func(dirPath, itemName string) (interface{}, error) {
			if !strings.HasSuffix(itemName, ".json") {
				return nil, nil
			}

			return []string{dirPath, itemName}, nil

		}

		localeFiles, err := pluginfinder.RecursiveDirSearch(localesPath, callback, nil, true)
		if err != nil {
			return err
		}

		for _, localeFile := range localeFiles {
			pathSpecsMap[localeFile.(string)] = unmarshalled
		}
	}

	assetsPathExists, _ := directoryExists(assetsPath)
	if assetsPathExists {
		fmt.Printf("Processing %v directory...", assetsPath)

		assetFiles, err := os.ReadDir(assetsPath)
		if err != nil {
			return err
		}

		for _, assetFile := range assetFiles {
			pathSpecsMap[assetFile.Name()] = unmarshalled
		}
	}

	distPathExists, _ := directoryExists(distPath)
	if distPathExists {
		fmt.Printf("Processing %v directory...", distPath)

		distFiles, err := os.ReadDir(distPath)
		if err != nil {
			return err
		}

		validSuffixes := []string{".js", ".js.map", ".css", ".css.map"}
		for _, distFile := range distFiles {
			if hasValidSuffix(distFile.Name(), validSuffixes) {
				pathSpecsMap[distFile.Name()] = unmarshalled
			}
		}
	}

	pathSpecs, err := json.MarshalIndent(pathSpecsMap, "", "  ")
	if err != nil {
		return err
	}

	os.WriteFile(staticApiFile, pathSpecs, 0644)

	return nil
}
