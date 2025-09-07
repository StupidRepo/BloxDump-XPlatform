package dumper

import (
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/config"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
	"path"
	"strings"
)

func DumpOne(parsed ParsedCache) {
	if !parsed.Success {
		println("Cannot dump unsuccessful parsed cache")
		return
	}

	// if link contains -AvatarHeadshot- or -Avatar-, skip
	if strings.Contains(parsed.Link, "-AvatarHeadshot-") || strings.Contains(parsed.Link, "-Avatar-") {
		println("Skipping avatar image")
		return
	}

	assetType, ext, friendlyName, folderName := utils.IdentifyAssetType(parsed.Data)
	if assetType == utils.Ignore || assetType == utils.Unknown {
		return
	}

	println("Preparing to save asset of type", friendlyName)

	outDir := path.Join(config.AppConfig.OutputDir, folderName)
	outPath := path.Join(outDir, path.Base(parsed.Path)+"."+ext)

	if !utils.DirectoryExists(outDir) {
		err := os.MkdirAll(outDir, os.ModePerm)
		if err != nil {
			println("Error creating output directory:", err.Error())
			return
		}
	}

	if utils.DirectoryExists(outPath) {
		//println("File already exists, skipping:", outPath)
		return
	}

	var finalData []byte // final data to write to file

	switch assetType {
	case utils.NoConvert:
		finalData = parsed.Data
	// TODO: Khronos, EXTM3U, FontList, FFlags JSON, client version JSON, OpenType/TrueType font, Zstandard, VideoFrame segment
	default:
		println("Unhandled asset type:", friendlyName)
	}

	println("Saving asset...")

	err := os.WriteFile(outPath, finalData, 0644)
	if err != nil {
		println("Error writing file:", err.Error())
		return
	}
}
func DumpAll() {
	// drain the parsed cache
	println("Dumping", len(parsedCache), "parsed cache files...")
	for len(parsedCache) > 0 {
		var parsed ParsedCache

		parsed, parsedCache = parsedCache[0], parsedCache[1:]
		DumpOne(parsed)
	}
}
