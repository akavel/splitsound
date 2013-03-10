package mp3agic

import (
	"fmt"
	"os"
)

const (
	MPEG_VERSION_1_0                    = "1.0"
	MPEG_VERSION_2_0                    = "2.0"
	MPEG_VERSION_2_5                    = "2.5"
	MPEG_LAYER_1                        = "I"
	MPEG_LAYER_2                        = "II"
	MPEG_LAYER_3                        = "III"
	CHANNEL_MODE_MONO                   = "Mono"
	CHANNEL_MODE_DUAL_MONO              = "Dual mono"
	CHANNEL_MODE_JOINT_STEREO           = "Joint stereo"
	CHANNEL_MODE_STEREO                 = "Stereo"
	MODE_EXTENSION_BANDS_4_31           = "Bands 4-31"
	MODE_EXTENSION_BANDS_8_31           = "Bands 8-31"
	MODE_EXTENSION_BANDS_12_31          = "Bands 12-31"
	MODE_EXTENSION_BANDS_16_31          = "Bands 16-31"
	MODE_EXTENSION_NONE                 = "None"
	MODE_EXTENSION_INTENSITY_STEREO     = "Intensity stereo"
	MODE_EXTENSION_M_S_STEREO           = "M/S stereo"
	MODE_EXTENSION_INTENSITY_M_S_STEREO = "Intensity & M/S stereo"
	MODE_EXTENSION_NA                   = "n/a"
	EMPHASIS_NONE                       = "None"
	EMPHASIS__50_15_MS                  = "50/15 ms"
	EMPHASIS_CCITT_J_17                 = "CCITT J.17"
	FRAME_SYNC                          = 0x7ff
)

type bitmask struct {
	mask  uint32
	shift uint
}

func newBitmask(mask uint32) bitmask {
	shift := uint(0)
	for ; shift <= 31; shift++ {
		if (mask>>shift)&1 == 1 {
			break
		}
	}
	return bitmask{mask: mask, shift: shift}
}

func (bm bitmask) Decode(data FrameHeader) int {
	return int((uint32(data) & bm.mask) >> bm.shift)
}

var (
	frameSyncMask     = newBitmask(0xffe00000)
	versionMask       = newBitmask(0x180000)
	layerMask         = newBitmask(0x60000)
	protectionMask    = newBitmask(0x10000)
	bitrateMask       = newBitmask(0xf000)
	sampleRateMask    = newBitmask(0xc00)
	paddingMask       = newBitmask(0x200)
	privateMask       = newBitmask(0x100)
	channelModeMask   = newBitmask(0xc0)
	modeExtensionMask = newBitmask(0x30)
	copyrightMask     = newBitmask(0x8)
	originalMask      = newBitmask(0x4)
	emphasisMask      = newBitmask(0x3)
)

type FrameHeader uint32

// TODO: unit tests for all methods!
func NewFrameHeader(buf []byte) (*FrameHeader, os.Error) {
	if len(buf) != 4 {
		return nil, os.NewError(fmt.Sprintf("decoding MPEG frame: expected %d bytes, got %d", 4, len(buf)))
	}
	f := FrameHeader(unpackInteger(buf))
	err := f.Verify()
	if err != nil {
		return nil, err
	}
	return &f, nil
}

func (f FrameHeader) Verify() (err os.Error) {
	sync := frameSyncMask.Decode(f)
	if sync != FRAME_SYNC {
		return os.NewError("Frame sync missing")
	}

	defer func() {
		if x := recover(); x != nil {
			err = os.NewError(fmt.Sprint(x))
		}
	}()

	f.Version()
	f.Layer()
	f.Protection()
	f.BitrateInKbps()
	f.SampleRate()
	f.Padding()
	f.Private()
	f.ChannelMode()
	f.ModeExtension()
	f.Copyrighted()
	f.Original()
	f.Emphasis()
	return
}

// TODO: test
func (f FrameHeader) Version() string {
	switch versionMask.Decode(f) {
	case 0:
		return MPEG_VERSION_2_5
	case 2:
		return MPEG_VERSION_2_0
	case 3:
		return MPEG_VERSION_1_0
	}
	panic("Invalid mpeg audio version in frame header")
}

func (f FrameHeader) layer() int {
	switch layerMask.Decode(f) {
	case 1:
		return 3
	case 2:
		return 2
	case 3:
		return 1
	}
	panic("Invalid mpeg layer description in frame header")
}

func (f FrameHeader) Layer() string {
	switch f.layer() {
	case 1:
		return MPEG_LAYER_1
	case 2:
		return MPEG_LAYER_2
	case 3:
		return MPEG_LAYER_3
	}
	panic("Invalid mpeg layer description in frame header")
}

// TODO: == CRC16 used?
func (f FrameHeader) Protection() bool {
	return protectionMask.Decode(f) == 1
}

func (f FrameHeader) BitrateInKbps() int {
	bitrate := bitrateMask.Decode(f)
	switch f.Version() {
	case MPEG_VERSION_1_0:
		switch f.layer() {
		case 1:
			switch bitrate {
			case 1:
				return 32
			case 2:
				return 64
			case 3:
				return 96
			case 4:
				return 128
			case 5:
				return 160
			case 6:
				return 192
			case 7:
				return 224
			case 8:
				return 256
			case 9:
				return 288
			case 10:
				return 320
			case 11:
				return 352
			case 12:
				return 384
			case 13:
				return 416
			case 14:
				return 448
			}
		case 2:
			switch bitrate {
			case 1:
				return 32
			case 2:
				return 48
			case 3:
				return 56
			case 4:
				return 64
			case 5:
				return 80
			case 6:
				return 96
			case 7:
				return 112
			case 8:
				return 128
			case 9:
				return 160
			case 10:
				return 192
			case 11:
				return 224
			case 12:
				return 256
			case 13:
				return 320
			case 14:
				return 384
			}
		case 3:
			switch bitrate {
			case 1:
				return 32
			case 2:
				return 40
			case 3:
				return 48
			case 4:
				return 56
			case 5:
				return 64
			case 6:
				return 80
			case 7:
				return 96
			case 8:
				return 112
			case 9:
				return 128
			case 10:
				return 160
			case 11:
				return 192
			case 12:
				return 224
			case 13:
				return 256
			case 14:
				return 320
			}
		}
	case MPEG_VERSION_2_0:
		fallthrough
	case MPEG_VERSION_2_5:
		switch f.layer() {
		case 1:
			switch bitrate {
			case 1:
				return 32
			case 2:
				return 48
			case 3:
				return 56
			case 4:
				return 64
			case 5:
				return 80
			case 6:
				return 96
			case 7:
				return 112
			case 8:
				return 128
			case 9:
				return 144
			case 10:
				return 160
			case 11:
				return 176
			case 12:
				return 192
			case 13:
				return 224
			case 14:
				return 256

			}
		case 2:
			fallthrough
		case 3:
			switch bitrate {
			case 1:
				return 8
			case 2:
				return 16
			case 3:
				return 24
			case 4:
				return 32
			case 5:
				return 40
			case 6:
				return 48
			case 7:
				return 56
			case 8:
				return 64
			case 9:
				return 80
			case 10:
				return 96
			case 11:
				return 112
			case 12:
				return 128
			case 13:
				return 144
			case 14:
				return 160
			}
		}
	}
	panic("Invalid bitrate in frame header")
}

func (f FrameHeader) SampleRate() uint32 {
	sampleRate := sampleRateMask.Decode(f)
	switch f.Version() {
	case MPEG_VERSION_1_0:
		switch sampleRate {
		case 0:
			return 44100
		case 1:
			return 48000
		case 2:
			return 32000
		}
	case MPEG_VERSION_2_0:
		switch sampleRate {
		case 0:
			return 22050
		case 1:
			return 24000
		case 2:
			return 16000
		}
	case MPEG_VERSION_2_5:
		switch sampleRate {
		case 0:
			return 11025
		case 1:
			return 12000
		case 2:
			return 8000
		}
	}

	panic("Invalid sample rate in frame header")
}

func (f FrameHeader) SamplesPerFrame() int {
	if f.layer() == 1 {
		return 384
	}

	// layer 2: always 1152 samples/frame
	// layer 3: MPEG1: 1152 samples/frame, MPEG2/2.5: 576
	// samples/frame
	if f.layer() == 2 || f.Version() == MPEG_VERSION_1_0 {
		return 1152
	}
	return 576
}

func (f FrameHeader) Padding() bool {
	return paddingMask.Decode(f) == 1
}

func (f FrameHeader) Private() bool {
	return privateMask.Decode(f) == 1
}

func (f FrameHeader) ChannelMode() string {
	switch channelModeMask.Decode(f) {
	case 0:
		return CHANNEL_MODE_STEREO
	case 1:
		return CHANNEL_MODE_JOINT_STEREO
	case 2:
		return CHANNEL_MODE_DUAL_MONO
	case 3:
		return CHANNEL_MODE_MONO
	}
	panic("Invalid channel mode in frame header")
}

func (f FrameHeader) Channels() int {
	if f.ChannelMode() == CHANNEL_MODE_MONO {
		return 1
	}
	return 2
}

func (f FrameHeader) ModeExtension() string {
	if f.ChannelMode() != CHANNEL_MODE_JOINT_STEREO {
		return MODE_EXTENSION_NA
	}
	modeExtension := modeExtensionMask.Decode(f)
	switch f.layer() {
	case 1:
		fallthrough
	case 2:
		switch modeExtension {
		case 0:
			return MODE_EXTENSION_BANDS_4_31
		case 1:
			return MODE_EXTENSION_BANDS_8_31
		case 2:
			return MODE_EXTENSION_BANDS_12_31
		case 3:
			return MODE_EXTENSION_BANDS_16_31
		}
	case 3:
		switch modeExtension {
		case 0:
			return MODE_EXTENSION_NONE
		case 1:
			return MODE_EXTENSION_INTENSITY_STEREO
		case 2:
			return MODE_EXTENSION_M_S_STEREO
		case 3:
			return MODE_EXTENSION_INTENSITY_M_S_STEREO
		}
	}
	panic("Invalid mode extension in frame header")
}

func (f FrameHeader) Copyrighted() bool {
	return copyrightMask.Decode(f) == 1
}

func (f FrameHeader) Original() bool {
	return originalMask.Decode(f) == 1
}

func (f FrameHeader) Emphasis() string {
	switch emphasisMask.Decode(f) {
	case 0:
		return EMPHASIS_NONE
	case 1:
		return EMPHASIS__50_15_MS
	case 3:
		return EMPHASIS_CCITT_J_17
	}
	panic("Invalid emphasis in frame header")
}

// TODO: verify if the calculations are ok, or should we use int64
func (f FrameHeader) LengthInBytes() int {
	pad := uint32(0)
	if f.Padding() {
		pad = 1
	}
	var length int
	if f.layer() == 1 {
		length = int(48000*uint32(f.BitrateInKbps())/f.SampleRate() + pad*4)
	} else {
		length = int(144000*uint32(f.BitrateInKbps())/f.SampleRate() + pad)
	}
	return length
}

func (f FrameHeader) SideInfoStart() int {
	if f.Protection() {
		return 6
	}
	return 4
}

func (f FrameHeader) SideInfoSize() int {
	if f.Version() == MPEG_VERSION_1_0 {
		if f.Channels() == 2 {
			return 32
		}
		return 17
	}

	if f.Channels() == 2 {
		return 17
	}
	return 9
}

func (f FrameHeader) SideInfoEnd() int {
	return f.SideInfoStart() + f.SideInfoSize()
}

func unpackInteger(b4 []byte) int32 {
	return int32(b4[0])<<24 + int32(b4[1])<<16 + int32(b4[2])<<8 + int32(b4[3])
}
