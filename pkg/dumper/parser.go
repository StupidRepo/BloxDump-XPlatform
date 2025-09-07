package dumper

import (
	"encoding/binary"
	"os"
)

type Asset struct {
	ID       int
	Location string
	Type     string
}

type ParsedCache struct {
	Success bool

	Link string
	Path string

	Data []byte
}

func ParseAll(paths []string) []ParsedCache {
	var parsedCache []ParsedCache

	for _, cacheAsset := range paths {
		parsed, err := parseCacheAsset(cacheAsset)
		if err != nil {
			println("Error parsing cache:", err.Error())
			continue
		}

		if parsed.Success {
			parsedCache = append(parsedCache, parsed)
		}
	}

	return parsedCache
}

func parseCacheAsset(path string) (ParsedCache, error) {
	// try to see if file exists
	file, err := os.Open(path)
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

	return parseCacheData(data, path), nil
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
