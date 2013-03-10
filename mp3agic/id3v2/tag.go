package id3v2

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	genrePattern = regexp.MustCompile(`^\(([^\)]+)\)(.*)`)
)

type Tag struct {
	header         *TagHeader
	extendedHeader []byte
	frameSets      map[string][]*Frame
}

func ExtractTag(mp3stream io.ReadSeeker) (*Tag, os.Error) {
	tag := Tag{}
	var err os.Error

	tag.header, err = ExtractTagHeader(mp3stream)
	if err != nil {
		return nil, err
	}

	if tag.header.ExtendedHeader() {
		err = tag.extractExtendedHeader(mp3stream)
		if err != nil {
			return nil, err
		}
	}

	err = tag.extractFrameSets(mp3stream)
	if err != nil {
		return nil, err
	}

	// TODO: extract footer

	return &tag, nil
}

func (tag *Tag) Version() string {
	return strconv.Itoa(tag.header.MajorVersion()) + "." + strconv.Itoa(tag.header.MinorVersion())
}

func (tag *Tag) DataLength() int {
	return tag.header.DataLength()
}

func (tag *Tag) Length() int {
	return tag.DataLength() + len(tag.header)
}

func (tag *Tag) Track() string {
	return tag.textFrameData("TRCK")
}

func (tag *Tag) Artist() string {
	return tag.textFrameData("TPE1")
}

func (tag *Tag) Title() string {
	return tag.textFrameData("TIT2")
}

func (tag *Tag) Album() string {
	return tag.textFrameData("TALB")
}

func (tag *Tag) Year() string {
	return tag.textFrameData("TYER")
}

func (tag *Tag) Genre() int {
	raw := strings.TrimSpace(tag.textFrameData("TCON"))
	if raw == "" {
		return -1
	}
	m := genrePattern.FindStringSubmatch(raw)
	if len(m) >= 2 {
		genre, err := strconv.Atoi(m[1])
		if err == nil {
			return genre
		}
		// TODO: search genre number by name in standard table/map
	}
	genre, err := strconv.Atoi(raw)
	if err == nil {
		return genre
	}
	return -1
}

func (tag *Tag) GenreDescription() string {
	raw := strings.TrimSpace(tag.textFrameData("TCON"))
	m := genrePattern.FindStringSubmatch(raw)
	if len(m) >= 3 {
		return m[2]
	}
	return raw
	// TODO: search genre name by number in standard table
}

func (tag *Tag) Comment() string {
	d := commentUnpack(tag.frameData("COMM"))
	if d == nil {
		return ""
	}
	return d.Comment
}

func (tag *Tag) Composer() string {
	return tag.textFrameData("TCOM")
}

func (tag *Tag) OriginalArtist() string {
	return tag.textFrameData("TOPE")
}

func (tag *Tag) Copyright() string {
	return tag.textFrameData("TCOP")
}

func (tag *Tag) Url() string {
	url := urlUnpack(tag.frameData("WXXX"))
	if url == nil {
		return ""
	}
	return url.Url
}

func (tag *Tag) Encoder() string {
	return tag.textFrameData("TENC")
}

func (tag *Tag) AlbumImage() []byte {
	pict := pictureUnpack(tag.frameData("APIC"))
	if pict == nil {
		return make([]byte, 0)
	}
	return pict.ImageData
}

func (tag *Tag) AlbumImageMimeType() string {
	pict := pictureUnpack(tag.frameData("APIC"))
	if pict == nil {
		return ""
	}
	return pict.MimeType
}

func (tag *Tag) textFrameData(id string) string {
	data := tag.frameData(id)
	if data == nil {
		return ""
	}
	text, err := textDecode(data[0], data[1:], false)
	if err != nil {
		return ""
	}
	return string(text)
}

func (tag *Tag) frameData(id string) []byte {
	fs, ok := tag.frameSets[id]
	if !ok {
		return nil
	}
	return fs[0].Data
}

func (tag *Tag) extractExtendedHeader(mp3stream io.ReadSeeker) os.Error {
	var lengthBuf [4]byte
	err := readStream(mp3stream, lengthBuf[:])
	if err != nil {
		return err
	}
	length := int(unpackSynchsafeInteger(lengthBuf[:]))

	data := make([]byte, length)
	err = readStream(mp3stream, data)
	if err != nil {
		return err
	}
	tag.extendedHeader = data
	return nil
}

// TODO: v2.2, v2.4
func (tag *Tag) extractFrameSets(mp3stream io.ReadSeeker) os.Error {
	//startOffset, err := mp3stream.Seek(0, 1) // remember current offset
	//if err != nil {
	//	return err
	//}

	framesLen := int64(tag.DataLength())
	if tag.header.Footer() {
		framesLen -= 10
	}

	tag.frameSets = make(map[string][]*Frame)
	fss := tag.frameSets
	for readn := int64(0); readn < framesLen; {
		frame, err := extractFrame(mp3stream)
		if err != nil {
			break
		}

		readn += int64(frame.Length())
		frameset, ok := fss[frame.Id()]
		if !ok {
			frameset = make([]*Frame, 0)
		}
		fss[frame.Id()] = append(frameset, frame)
	}

	return nil
}

func (tag *Tag) FrameSets() map[string][]*Frame {
	return tag.frameSets
}

type TagHeader [10]byte

const (
	id3v2tag_magic = "ID3"
)

func ExtractTagHeader(mp3stream io.ReadSeeker) (*TagHeader, os.Error) {
	_, err := mp3stream.Seek(0, 0)
	if err != nil {
		return nil, err
	}

	var header TagHeader
	err = readStream(mp3stream, header[:])
	if err != nil {
		return nil, err
	}

	err = header.Validate()
	if err != nil {
		return nil, err
	}

	return &header, nil
}

func (header *TagHeader) Validate() os.Error {
	if string(header[:len(id3v2tag_magic)]) != id3v2tag_magic {
		return os.NewError("stream does not start with ID3v2 magic code")
	}

	vmajor := header.MajorVersion()
	if vmajor != 2 && vmajor != 3 && vmajor != 4 {
		return os.NewError(fmt.Sprintf("unsupported ID3 version 2.%d.%d", vmajor, header.MinorVersion()))
	}

	flags := header.flags()
	if flags&0x0f != 0 {
		return os.NewError("unrecognized flags")
	}

	if header.DataLength() < 1 {
		return os.NewError("zero size tag")
	}

	return nil // ok
}

func (header *TagHeader) MajorVersion() int {
	return int(header[3])
}

func (header *TagHeader) MinorVersion() int {
	return int(header[4])
}

func (header *TagHeader) version() int16 {
	return int16(header[3])<<8 + int16(header[4])
}

func (header *TagHeader) DataLength() int {
	return int(unpackSynchsafeInteger(header[6:10]))
}

func (header *TagHeader) ExtendedHeader() bool {
	switch header.version() {
	case 0x30:
	case 0x40:
		return header.flags()&(1<<6) != 0
	}
	return false
}

func (header *TagHeader) Footer() bool {
	switch header.version() {
	case 0x40:
		return header.flags()&(1<<4) != 0
	}
	return false
}

func (header *TagHeader) flags() byte {
	return header[5]
}

func unpackSynchsafeInteger(b4 []byte) int32 {
	return (int32(b4[0]&0x7f) << 21) +
		(int32(b4[1]&0x7f) << 14) +
		(int32(b4[2]&0x7f) << 7) +
		int32(b4[3]&0x7f)
}

func readStream(stream io.Reader, buf []byte) os.Error {
	readn, err := stream.Read(buf)
	if err != nil {
		return err
	}
	if readn != len(buf) {
		return os.EOF
	}
	return nil
}
