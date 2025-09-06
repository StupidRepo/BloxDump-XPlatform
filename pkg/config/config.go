package config

import (
	"encoding/json"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
)

const (
	ExpectedVersion = 1
	Path            = "config.json"
)

type Config struct {
	Version int `json:"config_version"` // Config version for compatibility checks

	RobloxCacheDir string `json:"roblox_cache_dir"` // Directory for Roblox cache
	OutputDir      string `json:"output_dir"`       // Directory for output files
}

var AppConfig Config

func LoadConfig() error {
	file, err := os.Open(Path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&AppConfig); err != nil {
		return err
	}

	// check version
	if AppConfig.Version != ExpectedVersion {
		return utils.VersionMismatch(AppConfig.Version, ExpectedVersion)
	}

	// create output directory if it doesn't exist
	if !utils.DirectoryExists(AppConfig.OutputDir) {
		if err := os.MkdirAll(AppConfig.OutputDir, os.ModePerm); err != nil {
			return err
		}
	}

	// check directories
	return utils.DirectoriesExist([]string{AppConfig.RobloxCacheDir})
}

func SaveConfig() error {
	file, err := os.Create(Path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t") // Pretty-print with tabs
	if err := encoder.Encode(&AppConfig); err != nil {
		return err
	}

	println("Configuration saved successfully.")

	return nil
}

func CreateConfig() Config {
	return Config{
		Version: ExpectedVersion,

		OutputDir: "assets_output",
	}
}
