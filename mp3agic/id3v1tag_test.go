package mp3agic_test

import (
	"io"
	"mp3agic"
	"os"
	"strings"
	"testing"
)

const RES_DIR = "../test-res/"

func assert(t *testing.T, condition bool, errmsg ...interface{}) {
	if !condition {
		t.Error(errmsg...)
	}
}

func TestValid(t *testing.T) {
	var tag mp3agic.Id3v1Tag
	settag := func(s string) *mp3agic.Id3v1Tag {
		copy(tag[0:3], s)
		return &tag
	}
	assert(t, !tag.Valid(), "expected zero tag to be invalid")
	assert(t, settag("TAG").Valid(), "expected tag to be valid: TAG")

	assertInvalid := func(s string) {
		assert(t, !settag(s).Valid(), "expected tag to be invalid:", s)
	}
	assertInvalid("tag")
	assertInvalid("ID3")
}

func TestReadTagFieldsFromMp3(t *testing.T) {
	f, err := os.Open(RES_DIR+"v1andv23tags.mp3", os.O_RDONLY, 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	tag, err := mp3agic.ExtractId3v1Tag(f)
	if err != nil {
		t.Error(err)
		return
	}

	assert(t, tag.Track() == "1", "track", tag.Track())
	assert(t, tag.Artist() == "ARTIST123456789012345678901234", "artist", tag.Artist())
	assert(t, tag.Title() == "TITLE1234567890123456789012345", "title", tag.Title())
	assert(t, tag.Album() == "ALBUM1234567890123456789012345", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Genre() == 0x0d, "genre", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
	assert(t, tag.Comment() == "COMMENT123456789012345678901", "comment", tag.Comment())
}

type BufReaderAt []byte

func (buf BufReaderAt) ReadAt(p []byte, off int64) (n int, err os.Error) {
	if off >= int64(len(buf)) {
		return 0, os.EOF
	}
	return copy(p, (buf)[off:]), nil
}

func bufWrap(pattern string) ([]byte, *io.SectionReader) {
	buf := make([]byte, len(pattern))
	copy(buf, pattern)
	return buf, io.NewSectionReader(BufReaderAt(buf), 0, int64(len(buf)))
}

const (
	VALID_TAG                 = "TAGTITLE1234567890123456789012345ARTIST123456789012345678901234ALBUM12345678901234567890123452001COMMENT123456789012345678901234"
	VALID_TAG_WITH_WHITESPACE = "TAGTITLE                         ARTIST                        ALBUM                         2001COMMENT                        "
)

func TestExtractMaximumLengthFieldsFromValid10Tag(t *testing.T) {
	buf, reader := bufWrap(VALID_TAG)
	buf[len(buf)-1] = 0x8D
	tag, err := mp3agic.ExtractId3v1Tag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Title() == "TITLE1234567890123456789012345", "title", tag.Title())
	assert(t, tag.Artist() == "ARTIST123456789012345678901234", "artist", tag.Artist())
	assert(t, tag.Album() == "ALBUM1234567890123456789012345", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Comment() == "COMMENT12345678901234567890123", "comment", tag.Comment())
	assert(t, tag.Track() == "", "track", tag.Track())
	assert(t, tag.Genre() == 0x8D, "genre", tag.Genre())
	assert(t, tag.GenreDescription() == "Synthpop", "genre description", tag.GenreDescription())
}

func TestExtractMaximumLengthFieldsFromValid11Tag(t *testing.T) {
	buf, r := bufWrap(VALID_TAG)
	buf[len(buf)-3] = 0x00
	buf[len(buf)-2] = 0x01
	buf[len(buf)-1] = 0x0d
	tag, err := mp3agic.ExtractId3v1Tag(r)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Comment() == "COMMENT123456789012345678901", "comment", tag.Comment())
	assert(t, tag.Track() == "1", "track", tag.Track())
	assert(t, tag.Genre() == 0x0D, "genre", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
}

func TestExtractTrimmedFieldsFromValid11TagWithWhitespace(t *testing.T) {
	buf, r := bufWrap(VALID_TAG_WITH_WHITESPACE)
	buf[len(buf)-3] = 0x00
	buf[len(buf)-2] = 0x01
	buf[len(buf)-1] = 0x0d
	tag, err := mp3agic.ExtractId3v1Tag(r)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Title() == "TITLE", "title", tag.Title())
	assert(t, tag.Artist() == "ARTIST", "artist", tag.Artist())
	assert(t, tag.Album() == "ALBUM", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Comment() == "COMMENT", "comment", tag.Comment())
	assert(t, tag.Track() == "1", "track", tag.Track())
	assert(t, tag.Genre() == 0x0D, "genre", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
}

func TestExtractTrimmedFieldsFromValid11TagWithNullspace(t *testing.T) {
	buf, r := bufWrap(strings.Replace(VALID_TAG_WITH_WHITESPACE, " ", "\x00", -1))
	buf[len(buf)-3] = 0x00
	buf[len(buf)-2] = 0x01
	buf[len(buf)-1] = 0x0d
	tag, err := mp3agic.ExtractId3v1Tag(r)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Title() == "TITLE", "title", tag.Title())
	assert(t, tag.Artist() == "ARTIST", "artist", tag.Artist())
	assert(t, tag.Album() == "ALBUM", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Comment() == "COMMENT", "comment", tag.Comment())
	assert(t, tag.Track() == "1", "track", tag.Track())
	assert(t, tag.Genre() == 0x0D, "genre", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
}
