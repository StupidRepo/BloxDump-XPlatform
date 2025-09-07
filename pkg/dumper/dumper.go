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

	outDir := path.Join(*config.AppConfig.OutputDir, folderName)
	outPath := path.Join(outDir, path.Base(parsed.Path)+"."+ext)

	if utils.Exists(outPath) {
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

	if finalData == nil || len(finalData) == 0 {
		return
	}

	println("Saving asset...")

	if !utils.Exists(outDir) {
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			println("Error creating output directory:", err.Error())
			return
		}
	}

	err := os.WriteFile(outPath, finalData, 0644)
	if err != nil {
		println("Error writing file:", err.Error())
		return
	}
}
