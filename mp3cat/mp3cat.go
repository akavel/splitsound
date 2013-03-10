package main

import (
	"fmt"
	"mp3agic"
	"os"
	"strings"
)

const (
	MAX_CUSTOM_TAG_BYTES_TO_SHOW = 64
)

var (
	mp3file   *mp3agic.File
	vbrString = map[bool]string{true: "VBR", false: "CBR"}
)

func main() {
	if len(os.Args) < 2 {
		//TODO: port original usage()
		fmt.Printf("Usage: %v <FILE.mp3>\n", os.Args[0])
		return
	}

	exitcode := int(0)
	defer func() {
		os.Exit(exitcode)
	}()
	error := func(code int, err os.Error) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitcode = code
	}

	// TODO: buffer the file
	file, err := os.Open(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		error(1, err)
		return
		//fmt.Fprintf(os.Stderr, "error: %v\n", err)
		//os.Exit(1)
	}
	defer file.Close()

	// TODO: iterate args with wildcards expansion
	mp3file, err = mp3agic.ParseFile(file, 0)
	if err != nil {
		error(2, err)
		return
	}
	dumpMp3Fields()
	dumpId3Fields()
	dumpCustomTag()
}

func dumpMp3Fields() {
	dumpCsvRecord(false,
		mp3file.Filename(),
		mp3file.Length(),
		mp3file.LengthInSeconds(),
		mp3file.Version(),
		mp3file.Layer(),
		mp3file.SampleRate(),
		mp3file.Bitrate(),
		vbrString[mp3file.Vbr()],
		mp3file.ChannelMode())
}

func dumpId3Fields() {
	wrapper := &mp3agic.Id3Wrapper{
		Tagv1: mp3file.Id3v1Tag(),
		Tagv2: mp3file.Id3v2Tag()}
	var v1 string
	if wrapper.Tagv1 != nil {
		v1 = "1." + wrapper.Tagv1.Version()
	}
	var v2 string
	if wrapper.Tagv2 != nil {
		v2 = "2." + wrapper.Tagv2.Version()
	}
	dumpCsvRecord(false,
		v1,
		v2,
		wrapper.Track(),
		wrapper.Artist(),
		wrapper.Album(),
		wrapper.Title(),
		wrapper.Year(),
		wrapper.GenreDescription(),
		wrapper.Comment(),
		wrapper.Composer(),
		wrapper.OriginalArtist(),
		wrapper.Copyright(),
		wrapper.Url(),
		wrapper.Encoder(),
		wrapper.AlbumImageMimeType())
}

func dumpCustomTag() {
	var txt string
	if raw := mp3file.CustomTag(); raw != nil {
		if len(raw) > MAX_CUSTOM_TAG_BYTES_TO_SHOW {
			raw = raw[0:MAX_CUSTOM_TAG_BYTES_TO_SHOW]
		}
		asciiOnly := make([]byte, len(raw))
		for i := 0; i < len(raw); i++ {
			c := raw[i]
			if c >= 32 && c <= 126 {
				asciiOnly[i] = c
			} else {
				asciiOnly[i] = '?'
			}
		}
		txt = string(asciiOnly)
	}
	dumpCsvRecord(true,
		txt)
}

func dumpCsvRecord(endline bool, args ...interface{}) {
	for i := 0; i < len(args); i++ {
		s := fmt.Sprintf("%v", args[i])
		fmt.Printf(`"%s"`, strings.Replace(s, `"`, `""`, -1))

		if endline && i == len(args)-1 {
			fmt.Printf("\n")
		} else {
			fmt.Printf(",")
		}
	}
}
