package dumper

import (
	"encoding/binary"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/config"
	"github.com/StupidRepo/BloxDump-XPlatform/pkg/utils"
	"os"
	"path"
	"strings"
)

type Asset struct {
	ID       int
	Location string
	Type     string
}

type Cache struct {
	Path string
	Data []byte
}

type ParsedCache struct {
	Success bool

	Link string
	Path string

	Data []byte
}

var queue []Cache
var parsedCache []ParsedCache

func EnqueueAsset(path string) {
	queue = append(queue, Cache{Path: path})
}

func EnqueueAssets(paths []string) {
	for _, p := range paths {
		EnqueueAsset(p)
	}
}

func ScanAll() {
	// drain the queue
	println("Scanning", len(queue), "cache files...")
	for len(queue) > 0 {
		var cacheAsset Cache

		cacheAsset, queue = queue[0], queue[1:]
		parsed, err := parseCacheAsset(cacheAsset)
		if err != nil {
			println("Error parsing cache:", err.Error())
			continue
		}

		if parsed.Success {
			parsedCache = append(parsedCache, parsed)
		}
	}
}

func parseCacheAsset(cacheAsset Cache) (ParsedCache, error) {
	if cacheAsset.Data != nil {
		// parse from data
		return parseCacheData(cacheAsset.Data, cacheAsset.Path), nil
	}

	// try to see if file exists
	file, err := os.Open(cacheAsset.Path)
	if err != nil {
		return ParsedCache{Success: false}, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// parse from file
	fileInfo, err := file.Stat()
	if err != nil {
		return ParsedCache{Success: false}, err
	}

	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	if err != nil {
		return ParsedCache{Success: false}, err
	}

	return parseCacheData(data, cacheAsset.Path), nil
}

func parseCacheData(data []byte, path string) ParsedCache {
	if len(data) < 4 {
		return ParsedCache{Success: false}
	}

	// get magic number
	magic := string(data[0:4])
	if magic != "RBXH" {
		println("Ignoring non-RBXH magic:", magic)
		return ParsedCache{Success: false}
	}

	// skip header size (4 bytes)
	pos := 8

	// read link length
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	linklen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4

	var link string
	if linklen > 0 {
		if len(data) < pos+int(linklen) {
			return ParsedCache{Success: false}
		}
		link = string(data[pos : pos+int(linklen)])
		pos += int(linklen)
	}

	// skip byte
	pos++

	// read http status
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	status := binary.LittleEndian.Uint32(data[pos : pos+4])
	if status >= 300 {
		return ParsedCache{Success: false}
	}
	pos += 4

	// read header length
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	headerlen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4

	// skip xxhash digest (4 bytes)
	pos += 4

	// read content length
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	contentlen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4

	// Skip XXHash digest (4 bytes) and reserved bytes (4 bytes) and headers
	pos += 8 + int(headerlen)

	// read content
	if len(data) < pos+int(contentlen) {
		return ParsedCache{Success: false}
	}
	content := data[pos : pos+int(contentlen)]

	return ParsedCache{Success: true, Link: link, Path: path, Data: content}
}

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
