package id3v2_test
/*
import (
	"id3v2"
	"io"
	"os"
	. "testing"
)

const (
    tframe = `TP100"0ARTISTABCDEFGHIJKLMNOPQRSTUVWXYZ0`;
    longTframeA = `TP1\x00\x01\x01`
    longTframeB = `\x00Metamorphosis A a very long album B a very long album C a very long album D a very long album E a very long album F a very long album G a very long album H a very long album I a very long album J a very long album K a very long album L a very long album M\x00`

func TestReadValidLong32ObseleteTFrame(t *T) {


		byte[] bytes = BufferTools.stringToByteBuffer(LONG_T_FRAME, 0, LONG_T_FRAME.length());
		replaceNumbersWithBytes(bytes, 3);
		ID3v2ObseleteFrame frame = new ID3v2ObseleteFrame(bytes, 0);
		assertEquals(263, frame.getLength());
		assertEquals("TP1", frame.getId());
		String s = "0Metamorphosis A a very long album B a very long album C a very long album D a very long album E a very long album F a very long album G a very long album H a very long album I a very long album J a very long album K a very long album L a very long album M0";
		byte[] expectedBytes = BufferTools.stringToByteBuffer(s, 0, s.length());
		replaceNumbersWithBytes(expectedBytes, 0);
		assertTrue(Arrays.equals(expectedBytes, frame.getData()));
	}

	public void testShouldReadValid32ObseleteTFrame() throws Exception {
		byte[] bytes = BufferTools.stringToByteBuffer("xxxxx" + T_FRAME, 0, 5 + T_FRAME.length());
		replaceNumbersWithBytes(bytes, 8);
		ID3v2ObseleteFrame frame = new ID3v2ObseleteFrame(bytes, 5);
		assertEquals(40, frame.getLength());
		assertEquals("TP1", frame.getId());
		String s = "0ARTISTABCDEFGHIJKLMNOPQRSTUVWXYZ0";
		byte[] expectedBytes = BufferTools.stringToByteBuffer(s, 0, s.length());
		replaceNumbersWithBytes(expectedBytes, 0);
		assertTrue(Arrays.equals(expectedBytes, frame.getData()));
	}



func TestInitialiseFromHeaderBlockWithValidHeaders(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	buf[3] = 2
	buf[4] = 0
	tag, err := id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "2.0", "expected version 2.0, got", tag.Version())

	buf[3] = 3
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "3.0", "expected version 3.0, got", tag.Version())

	buf[3] = 4
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.Version() == "4.0", "expected version 4.0, got", tag.Version())
}

func TestCalculateCorrectDataLengthsFromHeaderBlock(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	tag, err := id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.DataLength() == 257, "data length expected 257, got", tag.DataLength())

	buf[8] = 0x09
	buf[9] = 0x41
	tag, err = id3v2.ExtractTag(reader)
	if err != nil {
		t.Error(err)
		return
	}
	assert(t, tag.DataLength() == 1217, "data length expected 1217, got", tag.DataLength())
}

func TestNonSupportedVersionInId3v2HeaderBlock(t *testing.T) {
	buf, reader := bufWrap(id3v2_header)
	buf[3] = 5
	buf[4] = 0
	_, err := id3v2.ExtractTag(reader)
	assert(t, err != nil, "expected error (wrong ID3v2 version), got nil")
}

func loadId3TagFile(fname string) (*id3v2.Tag, os.Error) {
	f, err := os.Open(RES_DIR+fname, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	tag, err := id3v2.ExtractTag(f)
	if err != nil {
		return nil, err
	}

	return tag, nil
}

func TestReadFramesFromMp3With32Tag(t *testing.T) {
	tag, err := loadId3TagFile("v1andv23tags.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}

	assert(t, tag.Version() == "3.0", "version expected 3.0, got", tag.Version())
	assert(t, tag.Length() == 0x44b, "length expected", 0x44b, "got", tag.Length())

	framesets := tag.FrameSets()
	if framesets == nil {
		t.Error("nil framesets")
		return
	}

	assert(t, len(framesets) == 12, "framesets length expected 12, got", len(framesets))
	assertFrameset := func(name string, count int) {
		fs, ok := framesets[name]
		if !ok {
			t.Error("absent frameset", name)
			return
		}
		n := len(fs)
		assert(t, n == count, "frameset", name, "elements expected", count, "got", n)
	}
	assertFrameset("TENC", 1)
	assertFrameset("WXXX", 1)
	assertFrameset("TCOP", 1)
	assertFrameset("TOPE", 1)
	assertFrameset("TCOM", 1)
	assertFrameset("COMM", 2)
	assertFrameset("TPE1", 1)
	assertFrameset("TALB", 1)
	assertFrameset("TRCK", 1)
	assertFrameset("TYER", 1)
	assertFrameset("TCON", 1)
	assertFrameset("TIT2", 1)
}

func TestReadId3v2WithFooter(t *testing.T) {
	tag, err := loadId3TagFile("v1andv24tags.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}
	assert(t, tag.Version() == "4.0", "version expected 4.0, got", tag.Version())
	assert(t, tag.Length() == 0x44b, "length expected", 0x44b, "got", tag.Length())
}

func TestReadTagFieldsFromMp3With32tag(t *testing.T) {
	tag, err := loadId3TagFile("v1andv23tagswithalbumimage.mp3")
	if err != nil {
		t.Error("error loading file:", err)
		return
	}
	assert(t, tag.Track() == "1", "track expected 1, got", tag.Track())
	assert(t, tag.Artist() == "ARTIST123456789012345678901234", "artist", tag.Artist())
	assert(t, tag.Title() == "TITLE1234567890123456789012345", "title", tag.Title())
	assert(t, tag.Album() == "ALBUM1234567890123456789012345", "album", tag.Album())
	assert(t, tag.Year() == "2001", "year", tag.Year())
	assert(t, tag.Genre() == 0x0d, "genre expected", 0x0d, "got", tag.Genre())
	assert(t, tag.GenreDescription() == "Pop", "genre description", tag.GenreDescription())
	assert(t, tag.Comment() == "COMMENT123456789012345678901", "comment", tag.Comment())
	assert(t, tag.Composer() == "COMPOSER23456789012345678901234", "composer", tag.Composer())
	assert(t, tag.OriginalArtist() == "ORIGARTIST234567890123456789012", "original artist", tag.OriginalArtist())
	assert(t, tag.Copyright() == "COPYRIGHT2345678901234567890123", "copyright", tag.Copyright())
	assert(t, tag.Url() == "URL2345678901234567890123456789", "url", tag.Url())
	assert(t, tag.Encoder() == "ENCODER234567890123456789012345", "encoder", tag.Encoder())
	assert(t, len(tag.AlbumImage()) == 1885, "len(album image)", len(tag.AlbumImage()))
	assert(t, tag.AlbumImageMimeType() == "image/png", "album image mime type", tag.AlbumImageMimeType())
}
*/
