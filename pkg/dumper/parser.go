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

const (
	Unknown = iota
	Ignore
	NoConvert
	Mesh
	WebP
)

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

func ScanOne() {
	var cacheAsset Cache
	if len(queue) == 0 {
		return
	}

	cacheAsset, queue = queue[0], queue[1:]
	parsed, err := ParseOne(cacheAsset)
	if err != nil {
		println("Error parsing cache:", err.Error())
		return
	}

	if parsed.Success {
		parsedCache = append(parsedCache, parsed)
	} else {
		println("Parsed cache was not successful for path:", cacheAsset.Path)
	}
}

func ScanAll() {
	for len(queue) > 0 {
		ScanOne()
	}
}

func ParseOne(cacheAsset Cache) (ParsedCache, error) {
	if cacheAsset.Data != nil {
		// parse from data
		return ParseData(cacheAsset.Data, cacheAsset.Path), nil
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

	return ParseData(data, cacheAsset.Path), nil
}

func ParseData(data []byte, path string) ParsedCache {
	if len(data) < 4 {
		return ParsedCache{Success: false}
	}

	// Check magic number
	magic := string(data[0:4])
	if magic != "RBXH" {
		println("Ignoring non-RBXH magic:", magic)
		return ParsedCache{Success: false}
	}

	// Skip header size (4 bytes)
	pos := 8

	// Read link length
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

	// Skip rogue byte
	pos++

	// Read status
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	status := binary.LittleEndian.Uint32(data[pos : pos+4])
	if status >= 300 {
		println("Ignoring non-successful cache:", status)
		return ParsedCache{Success: false}
	}
	pos += 4

	// Read header length
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	headerlen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4

	// Skip XXHash digest (4 bytes)
	pos += 4

	// Read content length
	if len(data) < pos+4 {
		return ParsedCache{Success: false}
	}
	contentlen := binary.LittleEndian.Uint32(data[pos : pos+4])
	pos += 4

	// Skip XXHash digest (4 bytes) and reserved bytes (4 bytes) and headers
	pos += 8 + int(headerlen)

	// Read content
	if len(data) < pos+int(contentlen) {
		return ParsedCache{Success: false}
	}
	content := data[pos : pos+int(contentlen)]

	println("Parsed cache with link:", link, "and content length:", len(content))

	// Return parsed cache
	return ParsedCache{Success: true, Link: link, Path: path, Data: content}
}

func DumpOne(parsed ParsedCache) {
	if !parsed.Success {
		println("Cannot dump unsuccessful parsed cache")
		return
	}

	assetType, ext, friendlyName, folderName := IdentifyOne(parsed.Data)
	if assetType == Ignore || assetType == Unknown {
		return
	}

	switch assetType {
	case NoConvert, WebP:
		// if link contains -AvatarHeadshot- or -Avatar-, skip
		if strings.Contains(parsed.Link, "-AvatarHeadshot-") || strings.Contains(parsed.Link, "-Avatar-") {
			println("Skipping avatar image")
			return
		}
		// create out directory if it doesn't exist
		outDir := path.Join(config.AppConfig.OutputDir, folderName)
		if !utils.DirectoryExists(outDir) {
			err := os.MkdirAll(outDir, os.ModePerm)
			if err != nil {
				println("Error creating output directory:", err.Error())
				return
			}
		}

		outPath := path.Join(outDir, path.Base(parsed.Path)+"."+ext)
		if utils.DirectoryExists(outPath) {
			//println("File already exists, skipping:", outPath)
			return
		}

		// write to file
		println("Saving asset of type", friendlyName)

		err := os.WriteFile(outPath, parsed.Data, 0644)
		if err != nil {
			println("Error writing file:", err.Error())
			return
		}

	default:
		//println("Unhandled asset type:", friendlyName)
	}
}

func DumpAll() {
	for _, parsed := range parsedCache {
		DumpOne(parsed)
	}
}

// IdentifyOne returns: (asset type, file extension, friendly name, folder name)
func IdentifyOne(data []byte) (int, string, string, string) {
	// no idea how this works but:
	/*
	   string begin = Encoding.UTF8.GetString(cnt[..Math.Min(48, cnt.Length - 1)]);
	   uint magic = BitConverter.ToUInt32(cnt, 0);
	*/
	begin := string(data[0:48])
	//magic := binary.LittleEndian.Uint32(data[0:4])

	/*
	   return begin switch
	   {
	       var s when s.Contains("<roblox!") => (AssetType.NoConvert, "rbxm", "RBXM", "RBXM"),
	       var s when s.Contains("<roblox xml") => (AssetType.Ignored, "", "unsupported XML", ""),
	       var s when !s.Contains("\"version") && s.StartsWith("version") => (AssetType.Mesh, "", "", ""),
	       var s when s.StartsWith("{\"translations") => (AssetType.Ignored, "", "translation list JSON", ""),
	       var s when s.Contains("{\"locale\":\"") => (AssetType.Translation, "", "", ""),
	       var s when s.Contains("PNG\r\n") => (AssetType.NoConvert, "png", "PNG", "Textures"),
	       var s when s.StartsWith("GIF87a") || s.StartsWith("GIF89a") => (AssetType.NoConvert, "gif", "GIF", "Textures"),
	       var s when s.Contains("JFIF") || s.Contains("Exif") => (AssetType.NoConvert, "jfif", "JFIF", "Textures"),
	       var s when s.StartsWith("RIFF") && s.Contains("WEBP") => (BlockAvatarImages ? (AssetType.WebP, "webp", "WebP", "Textures") : (AssetType.NoConvert, "webp", "WebP", "Textures")),
	       var s when s.StartsWith("OggS") => (AssetType.NoConvert, "ogg", "OGG", "Sounds"),
	       var s when s.StartsWith("ID3") || (cnt.Length > 2 && (cnt[0] & 0xFF) == 0xFF && (cnt[1] & 0xE0) == 0xE0) => (AssetType.NoConvert, "mp3", "MP3", "Sounds"),
	       var s when s.Contains("KTX 11") => (AssetType.Khronos, "", "", ""),
	       var s when s.StartsWith("#EXTM3U") => (AssetType.EXTM3U, "", "", ""),
	       var s when s.Contains("\"name\": \"") => (AssetType.FontList, "", "", ""),
	       var s when s.Contains("{\"applicationSettings") => (AssetType.Ignored, "", "FFlags JSON", ""),
	       var s when s.Contains("{\"version") => (AssetType.Ignored, "", "client version JSON", ""),
	       var s when s.Contains("GDEF") || s.Contains("GPOS") || s.Contains("GSUB") => (AssetType.Ignored, "", "OpenType/TrueType font", ""),
	       var s when magic == 0xFD2FB528 => (AssetType.Ignored, "", "Zstandard compressed data (likely FFlags)", ""),
	       var s when cnt.Length >= 4 && cnt[0] == 0x1A && cnt[1] == 0x45 && cnt[2] == 0xDF && cnt[3] == 0xA3 => (AssetType.Ignored, "", "VideoFrame segment", ""),
	       _ => (AssetType.Unknown, begin, "", "")
	   };
	*/
	switch {
	case strings.Contains(begin, "<roblox!"):
		return NoConvert, "rbxm", "RBXM", "RBXM"
	case strings.Contains(begin, "<roblox xml"):
		return Ignore, "", "unsupported XML", ""
	case !strings.Contains(begin, "\"version") && strings.HasPrefix(begin, "version"):
		return Mesh, "", "", ""
	case strings.HasPrefix(begin, "{\"translations"):
		return Ignore, "", "translation list JSON", ""
	case strings.Contains(begin, "{\"locale\":\""):
		return Ignore, "", "", ""
	case strings.Contains(begin, "PNG\r\n"):
		return NoConvert, "png", "PNG", "Textures"
	case strings.HasPrefix(begin, "GIF87a"), strings.HasPrefix(begin, "GIF89a"):
		return NoConvert, "gif", "GIF", "Textures"
	case strings.Contains(begin, "JFIF"), strings.Contains(begin, "Exif"):
		return NoConvert, "jfif", "JFIF", "Textures"
	case strings.HasPrefix(begin, "RIFF") && strings.Contains(begin, "WEBP"):
		return WebP, "webp", "WebP", "Textures"
	case strings.HasPrefix(begin, "OggS"):
		return NoConvert, "ogg", "OGG", "Sounds"
	case strings.HasPrefix(begin, "ID3"), len(data) > 2 && (data[0]&0xFF) == 0xFF && (data[1]&0xE0) == 0xE0:
		return NoConvert, "mp3", "MP3", "Sounds"
	// TODO: skipping Khronos, EXTM3U, FontList, FFlags JSON, client version JSON, OpenType/TrueType font, Zstandard, VideoFrame segment for now
	default:
		return Unknown, begin, "", ""
	}
}
