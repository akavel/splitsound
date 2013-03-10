package mp3agic_test

import (
	asrt "assert"
	"mp3agic"
	"os"
	. "testing"
)

var assertEq = asrt.Eq

func TestLoadMp3WithNoTags(t *T) {
	length := int64(2869)
	loadAndCheckTestMp3WithNoTags(t, length, 0)
	loadAndCheckTestMp3WithNoTags(t, length, 41)
	loadAndCheckTestMp3WithNoTags(t, length, 256)
	loadAndCheckTestMp3WithNoTags(t, length, 1024)
	loadAndCheckTestMp3WithNoTags(t, length, 5000)
}

func TestLoadMp3WithTags(t *T) {
	filename := "v1andv23tags.mp3"
	length := int64(4096)
	loadAndCheckTestMp3WithTags(t, filename, length, 0)
	loadAndCheckTestMp3WithTags(t, filename, length, 41)
	loadAndCheckTestMp3WithTags(t, filename, length, 256)
	loadAndCheckTestMp3WithTags(t, filename, length, 1024)
	loadAndCheckTestMp3WithTags(t, filename, length, 5000)
}

func TestLoadMp3WithFakeStartAndEndFrames(t *T) {
	filename := "dummyframes.mp3"
	length := int64(4096)
	loadAndCheckTestMp3WithTags(t, filename, length, 0)
	loadAndCheckTestMp3WithTags(t, filename, length, 41)
	loadAndCheckTestMp3WithTags(t, filename, length, 256)
	loadAndCheckTestMp3WithTags(t, filename, length, 1024)
	loadAndCheckTestMp3WithTags(t, filename, length, 5000)
}

func TestLoadMp3WithCustomTag(t *T) {
	filename := "v1andv23andcustomtags.mp3"
	length := int64(4129)
	loadAndCheckTestMp3WithCustomTag(t, filename, length, 0)
	loadAndCheckTestMp3WithCustomTag(t, filename, length, 41)
	loadAndCheckTestMp3WithCustomTag(t, filename, length, 256)
	loadAndCheckTestMp3WithCustomTag(t, filename, length, 1024)
	loadAndCheckTestMp3WithCustomTag(t, filename, length, 5000)
}

func TestErrorForFileThatIsNotAnMp3(t *T) {
	_, err := loadMp3(t, "notanmp3.mp3", 0)
	assert(t, err != nil, "err should be non nil")
	assertEq(t, "No mpegs frames found", err.String(), "error message")
}

func TestIgnoreIncompleteMpegFrame(t *T) {
	file, err := loadMp3(t, "incompletempegframe.mp3", 256)
	if err != nil {
		t.Fatal(err)
		return
	}
	assertEq(t, int64(0x44b), file.XingOffset(), "xing offset")
	assertEq(t, int64(0x5ec), file.StartOffset(), "start offset")
	assertEq(t, int64(0xf17), file.EndOffset(), "end offset")
	assert(t, file.HasId3v1Tag(), "has id3v1 tag")
	assert(t, file.HasId3v2Tag(), "has id3v2 tag")
	assertEq(t, 5, file.FrameCount(), "frame count")
}

func loadAndCheckTestMp3WithNoTags(t *T, length int64, bufferLength int) {
	file := loadAndCheckTestMp3(t, "notags.mp3", length, bufferLength)
	assertEq(t, int64(0x000), file.XingOffset(), "xing offset")
	assertEq(t, int64(0x1a1), file.StartOffset(), "start offset")
	assertEq(t, int64(0xb34), file.EndOffset(), "end offset")
	assert(t, !file.HasId3v1Tag(), "has id3v1 tag")
	assert(t, !file.HasId3v2Tag(), "has id3v2 tag")
	assert(t, !file.HasCustomTag(), "has custom tag")
}

func loadAndCheckTestMp3WithTags(t *T, filename string, length int64, bufferLength int) {
	file := loadAndCheckTestMp3(t, filename, length, bufferLength)
	assertEq(t, int64(0x44b), file.XingOffset(), "xing offset")
	assertEq(t, int64(0x5ec), file.StartOffset(), "start offset")
	assertEq(t, int64(0xf7f), file.EndOffset(), "end offset")
	assert(t, file.HasId3v1Tag(), "has id3v1 tag")
	assert(t, file.HasId3v2Tag(), "has id3v2 tag")
	assert(t, !file.HasCustomTag(), "has custom tag")
}

func loadAndCheckTestMp3WithCustomTag(t *T, filename string, length int64, bufferLength int) {
	file := loadAndCheckTestMp3(t, filename, length, bufferLength)
	assertEq(t, int64(0x44b), file.XingOffset(), "xing offset")
	assertEq(t, int64(0x5ec), file.StartOffset(), "start offset")
	assertEq(t, int64(0xf7f), file.EndOffset(), "end offset")
	assert(t, file.HasId3v1Tag(), "has id3v1 tag")
	assert(t, file.HasId3v2Tag(), "has id3v2 tag")
	assert(t, file.HasCustomTag(), "has custom tag")
}

func loadMp3(t *T, filename string, bufferLength int) (*mp3agic.File, os.Error) {
	// TODO: buffer the file
	rawfile, err := os.Open(RES_DIR+filename, os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
		return nil, err
	}
	defer rawfile.Close()

	file, err := mp3agic.ParseFile(rawfile, bufferLength)
	return file, err
}

func loadAndCheckTestMp3(t *T, filename string, length int64, bufferLength int) *mp3agic.File {
	file, err := loadMp3(t, filename, bufferLength)
	if err != nil {
		t.Fatal(err)
		return nil
	}

	assert(t, file.HasXingFrame(), "has xing frame")
	assertEq(t, 6, file.FrameCount(), "frame count")
	assertEq(t, mp3agic.MPEG_VERSION_1_0, file.Version(), "version")
	assertEq(t, mp3agic.MPEG_LAYER_3, file.Layer(), "layer")
	assertEq(t, 44100, file.SampleRate(), "sample rate")
	assertEq(t, mp3agic.CHANNEL_MODE_JOINT_STEREO, file.ChannelMode(), "channel mode")
	assertEq(t, mp3agic.EMPHASIS_NONE, file.Emphasis(), "emphasis")
	assert(t, file.Original(), "original")
	assert(t, !file.Copyrighted(), "copyrighted")
	assertEq(t, 128, file.XingBitrate(), "xing bitrate")
	assertEq(t, 125, file.Bitrate(), "file bitrate")
	assertEq(t, 1, file.Bitrates()[224], "bitrates[224]")
	assertEq(t, 1, file.Bitrates()[112], "bitrates[112]")
	assertEq(t, 2, file.Bitrates()[96], "bitrates[96]")
	assertEq(t, 1, file.Bitrates()[192], "bitrates[192]")
	assertEq(t, 1, file.Bitrates()[32], "bitrates[32]")
	assertEq(t, length, file.Length(), "length")
	assertEq(t, int64(156), file.LengthInMilliseconds(), "length in ms")

	return file
}
