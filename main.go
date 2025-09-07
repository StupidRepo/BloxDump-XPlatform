package main

import (
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/config"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/dumper"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
	"sync"
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
	println("Roblox Cache Directory:", *config.AppConfig.RobloxCacheDir)

	found, err := utils.GetFilesInDirectory(*config.AppConfig.RobloxCacheDir)
	if err != nil {
		panic(err)
	}

	println("Found", len(found), "files in cache directory.")

	parsedList := dumper.ParseAll(found)
	println("Dumping", len(parsedList), "parsed cache files...")

	var wg sync.WaitGroup
	jobs := make(chan dumper.ParsedCache, len(parsedList))

	for w := 0; w < *config.AppConfig.Threads; w++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for parsed := range jobs {
				dumper.DumpOne(parsed)
			}
		}()
	}

	for _, parsed := range parsedList {
		jobs <- parsed
	}
	close(jobs)

	wg.Wait()
	println("Dump complete.")
}
