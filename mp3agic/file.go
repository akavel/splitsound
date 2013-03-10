package mp3agic

import (
	"io"
	"os"
	"mp3agic/id3v2"
)

type File struct {
	id3v1tag      *Id3v1Tag
	id3v2tag      *id3v2.Tag
	startOffset   int64
	endOffset     int64
	length        int64
	frameCount    int
	xingOffset    int64
	bitrates      map[int]int
	bitrate       float64
	xingBitrate   int
	channelMode   string
	emphasis      string
	layer         string
	modeExtension string
	sampleRate    uint32
	version       string
	copyrighted   bool
	original      bool
	customTag     []byte
}

const (
	DEFAULT_BUFFER_LENGTH = 65536
	MINIMUM_BUFFER_LENGTH = 40
)

// FIXME: use some minimal interface instead of os.File in the argument
// TODO: verify if errors handling (and ignoring) is ok
func ParseFile(raw *os.File, bufferLength int) (*File, os.Error) {
	if bufferLength == 0 {
		bufferLength = DEFAULT_BUFFER_LENGTH
	}
	mp3file := &File{
		startOffset: -1,
		endOffset:   -1,
		xingOffset:  -1,
		bitrates:    make(map[int]int)}
	mp3file.calcLength(raw)
	// TODO: try to move all Seek()s into this function

	mp3file.id3v1tag, _ = ExtractId3v1Tag(raw)

	offset := int64(0)
	id3v2tagHeader, _ := id3v2.ExtractTagHeader(raw)
	if id3v2tagHeader != nil {
		offset = int64(len(id3v2tagHeader)) + int64(id3v2tagHeader.DataLength())
	}

	err := mp3file.scanFile(raw, offset, bufferLength)
	if err != nil {
		return nil, err
	}
	if mp3file.startOffset < 0 {
		return nil, os.NewError("No mpegs frames found")
	}
	mp3file.id3v2tag, _ = id3v2.ExtractTag(raw)
	mp3file.extractCustomTag(raw)

	return mp3file, nil
}

// TODO: handle Seek() errors
func (f *File) calcLength(stream io.Seeker) {
	offset, _ := stream.Seek(0, 1) // remember current position
	length, _ := stream.Seek(0, 2)
	_, _ = stream.Seek(offset, 0)
	f.length = length
}

func (f *File) Length() int64 {
	return f.length
}

// FIXME: handle Seek() errors
func (f *File) scanFile(raw *os.File, offset int64, bufferLength int) os.Error {
	var err os.Error
	raw.Seek(offset, 0)

	buf := make([]byte, bufferLength)

	lastOffset := offset
	for lastBlock := false; !lastBlock; {
		readn, _ := raw.Read(buf[:]) // FIXME: handle error code
		if readn < len(buf) {
			lastBlock = true // FIXME: Java; might not be true in Go (rather check for: 0, os.EOF)
		}
		if readn < MINIMUM_BUFFER_LENGTH { // FIXME: data can be lost???
			continue
		}

		tmpOffset := 0
		if f.startOffset < 0 {
			tmpOffset = f.scanBlockForStart(buf, readn, offset, tmpOffset)
			lastOffset = f.startOffset
		}
		tmpOffset, err = f.scanBlock(buf, readn, offset, tmpOffset)

		if err == nil {
			offset += int64(tmpOffset)
			raw.Seek(offset, 0)
		}

		if err != nil { // in Java mp3agic was: "catch(InvalidDataException)"
			if f.frameCount < 2 {
				f.startOffset = -1
				f.xingOffset = -1
				f.frameCount = 0
				f.bitrates = make(map[int]int)
				lastBlock = false
				offset = lastOffset + 1
				if offset == 0 {
					return os.NewError("Valid start of mpeg frames not found:" + err.String())
				}
				raw.Seek(offset, 0)
			}
			return nil
		}
	}
	return nil
}

func (f *File) scanBlockForStart(buf []byte, readn int, offset int64, tmpOffset int) int {
	for tmpOffset < readn-MINIMUM_BUFFER_LENGTH {
		if buf[tmpOffset] != 0xff || buf[tmpOffset+1]&0xe0 != 0xe0 {
			tmpOffset++
			continue
		}

		frame, err := NewFrameHeader(buf[tmpOffset : tmpOffset+4])
		if err != nil { // could not decode
			tmpOffset++
			continue
		}

		if f.xingOffset < 0 && HasXingFrameTag(buf[tmpOffset:]) {
			f.xingOffset = offset + int64(tmpOffset)
			f.xingBitrate = frame.BitrateInKbps()
			tmpOffset += frame.LengthInBytes()
			continue
		}

		f.startOffset = offset + int64(tmpOffset)
		f.channelMode = frame.ChannelMode()
		f.emphasis = frame.Emphasis()
		f.layer = frame.Layer()
		f.modeExtension = frame.ModeExtension()
		f.sampleRate = frame.SampleRate()
		f.version = frame.Version()
		f.copyrighted = frame.Copyrighted()
		f.original = frame.Original()
		f.frameCount++
		f.addBitrate(frame.BitrateInKbps())
		tmpOffset += frame.LengthInBytes()
		break
	}
	return tmpOffset
}

func (f *File) scanBlock(buf []byte, readn int, offset int64, tmpOffset int) (int, os.Error) {
	for tmpOffset < readn-MINIMUM_BUFFER_LENGTH {
		frame, err := NewFrameHeader(buf[tmpOffset : tmpOffset+4])
		if err != nil {
			return 0, os.NewError("Could not decode MPEG frame: " + err.String())
		}
		err = f.sanityCheckFrame(frame, offset+int64(tmpOffset))
		if err != nil {
			return 0, err
		}
		newEndOffset := offset + int64(tmpOffset) + int64(frame.LengthInBytes()) - 1
		if newEndOffset >= f.maxEndOffset() {
			break
		}
		f.endOffset = newEndOffset
		f.frameCount++
		f.addBitrate(frame.BitrateInKbps())
		tmpOffset += frame.LengthInBytes()
	}
	return tmpOffset, nil
}

func (f *File) extractCustomTag(stream io.ReadSeeker) os.Error {
	bufferLength := int(f.Length() - (f.endOffset + 1))
	if f.HasId3v1Tag() {
		bufferLength -= Id3v1_length
	}
	if bufferLength <= 0 {
		f.customTag = nil
		return nil
	}

	customTag := make([]byte, bufferLength)
	stream.Seek(f.endOffset+1, 0)
	readn, err := stream.Read(customTag)
	if err != nil {
		return os.NewError("Reading custom tag: " + err.String())
	}
	if readn < len(customTag) {
		return os.NewError("Reading custom tag: Not enough bytes read")
	}
	f.customTag = customTag
	return nil
}

func (f *File) sanityCheckFrame(frame *FrameHeader, offset int64) os.Error {
	if f.SampleRate() != frame.SampleRate() {
		return os.NewError("Inconsistent frame header (sample rate)")
	}
	if f.Layer() != frame.Layer() {
		return os.NewError("Inconsistent frame header (layer)")
	}
	if f.Version() != frame.Version() {
		return os.NewError("Inconsistent frame header (version)")
	}
	if offset+int64(frame.LengthInBytes()) > f.Length() {
		return os.NewError("Frame would extend beyond end of file")
	}
	return nil
}

func (f *File) maxEndOffset() int64 {
	length := f.Length()
	if f.HasId3v1Tag() {
		length -= Id3v1_length
	}
	return length
}

func (f *File) addBitrate(bitrate int) {
	old, ok := f.bitrates[bitrate]
	if !ok {
		old = 0
	}
	f.bitrates[bitrate] = old + 1
	f.bitrate = ((f.bitrate * float64(f.frameCount-1)) + float64(bitrate)) / float64(f.frameCount)
}

func (f *File) Filename() string {
	return ""
}

func (f *File) LengthInSeconds() int64 {
	return (f.LengthInMilliseconds() + 500) / 1000
}

func (f *File) Version() string {
	return f.version
}

func (f *File) Layer() string {
	return f.layer
}

func (f *File) SampleRate() uint32 {
	return f.sampleRate
}

func (f *File) Bitrate() int {
	return int(f.bitrate + 0.5)
}

func (f *File) Vbr() bool {
	return len(f.bitrates) > 1
}

func (f *File) ChannelMode() string {
	return f.channelMode
}

func (f *File) XingOffset() int64 {
	return f.xingOffset
}

func (f *File) HasId3v1Tag() bool {
	return f.id3v1tag != nil
}

func (f *File) HasId3v2Tag() bool {
	return f.id3v2tag != nil
}

func (f *File) HasCustomTag() bool {
	return f.customTag != nil
}

func (f *File) StartOffset() int64 {
	return f.startOffset
}

func (f *File) EndOffset() int64 {
	return f.endOffset
}

func (f *File) HasXingFrame() bool {
	return f.xingOffset >= 0
}

func (f *File) FrameCount() int {
	return f.frameCount
}

func (f *File) Emphasis() string {
	return f.emphasis
}

func (f *File) Original() bool {
	return f.original
}

func (f *File) Copyrighted() bool {
	return f.copyrighted
}

func (f *File) XingBitrate() int {
	return f.xingBitrate
}

func (f *File) Bitrates() map[int]int {
	return f.bitrates
}

func (f *File) LengthInMilliseconds() int64 {
	d := 8 * float64(f.endOffset-f.startOffset)
	return int64(d/f.bitrate + 0.5)
}

func (f *File) CustomTag() []byte {
	return f.customTag
}

func (f *File) Id3v1Tag() *Id3v1Tag {
	return f.id3v1tag
}

func (f *File) Id3v2Tag() *id3v2.Tag {
	return f.id3v2tag
}

func probeXing(buf []byte, offset int) bool {
	if len(buf) < offset+3 {
		return false
	}
	probe := string(buf[offset : offset+4])
	return probe == "Xing" || probe == "Info"
}

func HasXingFrameTag(buf []byte) bool {
	return probeXing(buf, 13) || probeXing(buf, 21) || probeXing(buf, 36)
}
