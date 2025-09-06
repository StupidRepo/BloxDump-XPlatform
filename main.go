package main

import (
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/config"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/dumper"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
)

func main() {
	err := config.LoadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			println("Config file not found. Creating default config file...")

			config.AppConfig = config.CreateConfig()
			err = config.SaveConfig()
		} else {
			panic(err)
		}
	}

	println("Config loaded successfully.")
	println("Roblox Cache Directory:", config.AppConfig.RobloxCacheDir)

	found, err := utils.GetFilesInDirectory(config.AppConfig.RobloxCacheDir)
	if err != nil {
		panic(err)
	}

	println("Found", len(found), "files in cache directory.")
	dumper.EnqueueAssets(found)

	println("Starting dump...")
	dumper.ScanAll()
	dumper.DumpAll()

	println("Dump complete.")
}
