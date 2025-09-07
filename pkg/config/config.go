package config

import (
	"encoding/json"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
)

const (
	ExpectedVersion = 3
	Path            = "config.json"
)

type Config struct {
	Version *int `json:"config_version"`

	RobloxCacheDir *string `json:"roblox_cache_dir"`
	OutputDir      *string `json:"output_dir"`

	Threads *int `json:"threads"`

	Migrations *[]int `json:"migrations"`
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
	if *AppConfig.Version != ExpectedVersion {
		// migrate config
		println("Config version mismatch, migrating config...")
		MigrateConfig(&AppConfig)

		if err := SaveConfig(); err != nil {
			return err
		}
	}

	// create output directory if it doesn't exist
	if !utils.Exists(*AppConfig.OutputDir) {
		if err := os.MkdirAll(*AppConfig.OutputDir, os.ModePerm); err != nil {
			return err
		}
	}

	// check directories
	return utils.DirectoriesExist([]string{*AppConfig.RobloxCacheDir})
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

	return nil
}

func CreateConfig() Config {
	version := ExpectedVersion
	robloxCacheDir := "change-me"
	outputDir := "assets_output"
	threads := 4

	var migrations []int

	return Config{
		Version:        &version,
		RobloxCacheDir: &robloxCacheDir,
		OutputDir:      &outputDir,
		Threads:        &threads,
		Migrations:     &migrations,
	}
}
