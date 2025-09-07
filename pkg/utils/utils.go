package utils

import (
	"bytes"
	"os"
	"regexp"
	"strings"
)

const (
	Unknown = iota
	Ignore

	NoConvert

	//EXTM3U

	Mesh
	Kronos
)

var versionRegex = regexp.MustCompile(`^version \d+\.\d+`)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DirectoriesExist(paths []string) error {
	for _, path := range paths {
		if path == "" {
			return EmptyDirectory
		}
		if !Exists(path) {
			return InvalidDirectory(path)
		}
	}

	return nil
}

func GetFilesInDirectory(path string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, path+"/"+entry.Name())
		}
	}

	return files, nil
}

// IdentifyAssetType returns: (asset type, file extension, friendly name, folder name)
func IdentifyAssetType(data []byte) (int, string, string, string) {
	begin := string(data[0:48])

	switch {
	// RoBloX Models
	case strings.Contains(begin, "<roblox!"):
		return NoConvert, "rbxm", "RBXM", "RBXM"

	// images (2d)
	case strings.Contains(begin, "PNG\r\n"):
		return NoConvert, "png", "PNG", "Images"
	case strings.HasPrefix(begin, "GIF87a"), strings.HasPrefix(begin, "GIF89a"):
		return NoConvert, "gif", "GIF", "Images"
	case strings.Contains(begin, "JFIF"), strings.Contains(begin, "Exif"):
		return NoConvert, "jfif", "JFIF", "Images"
	case strings.HasPrefix(begin, "RIFF") && strings.Contains(begin, "WEBP"):
		return NoConvert, "webp", "WebP", "Images"

	// textures (3d)
	// kronos prefix: 0xAB, 0x4B, 0x54, 0x58, 0x20, 0x31, 0x31, 0xBB, 0x0D, 0x0A, 0x1A, 0x0A
	// a.k.a "KTX 11" with some extra bytes (but we're cool so we check the full prefix)
	// btw, "png" because we're going to convert it to png later.
	case bytes.HasPrefix(data, []byte{0xAB, 0x4B, 0x54, 0x58, 0x20, 0x31, 0x31, 0xBB, 0x0D, 0x0A, 0x1A, 0x0A}):
		return Kronos, "png", "KTX (Kronos TeXture)", "Textures"

	// meshes
	// TODO: convert to FBX instead of OBJ, and have texture support (if possible)
	case strings.HasPrefix(begin, "version ") && versionRegex.MatchString(begin):
		return Mesh, "obj", "Mesh", "Meshes"

	// audio
	case strings.HasPrefix(begin, "OggS"):
		return NoConvert, "ogg", "OGG", "Audio"
	case strings.HasPrefix(begin, "ID3"), len(data) > 2 && (data[0]&0xFF) == 0xFF && (data[1]&0xE0) == 0xE0:
		return NoConvert, "mp3", "MP3", "Audio"

	// TODO: Khronos, FontList, OpenType/TrueType font, Zstandard
	default:
		return Unknown, begin, "", ""
	}
}
