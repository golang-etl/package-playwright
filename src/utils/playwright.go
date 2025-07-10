package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-etl/package-general/src/utils"
	"github.com/playwright-community/playwright-go"
)

func ToSameSite(value string) *playwright.SameSiteAttribute {
	switch value {
	case "Lax":
		return playwright.SameSiteAttributeLax
	case "Strict":
		return playwright.SameSiteAttributeStrict
	}

	return playwright.SameSiteAttributeNone
}

func GetStorage(page playwright.Page) (localStorage map[string]string, sessionStorage map[string]string, err error) {
	localStorageRaw, err := page.Evaluate(`Object.fromEntries(Object.entries(localStorage))`)

	if err != nil {
		return nil, nil, fmt.Errorf("error getting localStorage: %w", err)
	}

	sessionStorageRaw, err := page.Evaluate(`Object.fromEntries(Object.entries(sessionStorage))`)

	if err != nil {
		return nil, nil, fmt.Errorf("error getting sessionStorage: %w", err)
	}

	localStorage = MapToStringMap(localStorageRaw)
	sessionStorage = MapToStringMap(sessionStorageRaw)

	return localStorage, sessionStorage, nil
}

func MapToStringMap(input interface{}) map[string]string {
	result := make(map[string]string)

	if inputMap, ok := input.(map[string]interface{}); ok {
		for k, v := range inputMap {
			result[k] = fmt.Sprintf("%v", v)
		}
	}

	return result
}

func JSONStringToOptionalStorageState(input string) (*playwright.OptionalStorageState, error) {
	var state playwright.OptionalStorageState
	err := json.Unmarshal([]byte(input), &state)

	if err != nil {
		return nil, err
	}

	return &state, nil
}

func GenerateOptionalStorageState(input string) *playwright.OptionalStorageState {
	state, err := JSONStringToOptionalStorageState(input)

	if err != nil {
		panic(fmt.Errorf("error al convertir el estado de la sesion: %w", err))
	}

	return state
}

func GenerateSessionStorageInitScript(sessionStorage map[string]string) string {
	script := "(() => {"

	for key, value := range sessionStorage {
		script += fmt.Sprintf("sessionStorage.setItem(%q, %q);", key, value)
	}

	script += "})();"

	return script
}

func GetTracingFolderPath() string {
	if utils.IsRuntimeEnvironmentLocal() {
		return "tracings"
	}

	if utils.IsRuntimeEnvironmentGCPCloudRun() {
		return os.TempDir()
	}

	panic("can't get tracing folder path in this runtime environment")
}

func StartTracing(context playwright.BrowserContext) {
	err := context.Tracing().Start(playwright.TracingStartOptions{
		Screenshots: playwright.Bool(true),
		Snapshots:   playwright.Bool(true),
		Sources:     playwright.Bool(true),
	})

	if err != nil {
		panic(fmt.Errorf("error starting tracing: %w", err))
	}
}

func StopTracing(context playwright.BrowserContext) (traceToken string, fileName string, fileAbsolutePath string) {
	folderPath := GetTracingFolderPath()
	traceToken = utils.GenerateHexToken(16)
	fileName = traceToken + ".zip"
	fileAbsolutePath = filepath.Join(folderPath, fileName)
	err := context.Tracing().Stop(fileAbsolutePath)

	if err != nil {
		panic(fmt.Errorf("error stopping tracing: %w", err))
	}

	return traceToken, fileName, fileAbsolutePath
}

func CookiesToHeader(cookies []playwright.Cookie) string {
	var parts []string

	for _, c := range cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}

	return strings.Join(parts, "; ")
}
