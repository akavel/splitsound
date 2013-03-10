package main

import (
	"io"
	"mp3agic"
	"os"
)

const (
	MAX_MPAFRAME_SIZE = 2048

	FILTER_MPEG1   = uint32(0x0001)
	FILTER_MPEG2   = uint32(0x0002)
	FILTER_MPEG25  = uint32(0x0004)
	FILTER_LAYER1  = uint32(0x0008)
	FILTER_LAYER2  = uint32(0x0010)
	FILTER_LAYER3  = uint32(0x0020)
	FILTER_32000HZ = uint32(0x0040)
	FILTER_44100HZ = uint32(0x0080)
	FILTER_48000HZ = uint32(0x0100)
	FILTER_MONO    = uint32(0x0200)
	FILTER_STEREO  = uint32(0x0400)
)

type mpaFrameParser struct {
	ips            io.Reader
	junkh          *MyCountingJunkHandler
	masker, masked uint32
	headBuff       [4]byte
}

func getMpegFilter(fh *mp3agic.FrameHeader) uint32 {
	if fh.Verify() != nil {
		return 0
	}
	switch fh.Version() {
	case mp3agic.MPEG_VERSION_1_0:
		return FILTER_MPEG1
	case mp3agic.MPEG_VERSION_2_0:
		return FILTER_MPEG2
	case mp3agic.MPEG_VERSION_2_5:
		return FILTER_MPEG25
	}
	return 0
}

func getModeFilter(fh *mp3agic.FrameHeader) uint32 {
	if fh.Verify() != nil {
		return 0
	}
	if fh.Channels() == 1 {
		return FILTER_MONO
	}
	return FILTER_STEREO
}

func getSamplingrateFilter(fh *mp3agic.FrameHeader) uint32 {
	if fh.Verify() != nil {
		return 0
	}
	switch fh.SampleRate() {
	case 32000, 16000, 8000:
		return FILTER_32000HZ
	case 44100, 22050, 11025:
		return FILTER_44100HZ
	case 48000, 24000, 12000:
		return FILTER_48000HZ
	}
	return 0
}

func getLayerFilter(fh *mp3agic.FrameHeader) uint32 {
	if fh.Verify() != nil {
		return 0
	}
	switch fh.Layer() {
	case mp3agic.MPEG_LAYER_1:
		return FILTER_LAYER1
	case mp3agic.MPEG_LAYER_2:
		return FILTER_LAYER2
	case mp3agic.MPEG_LAYER_3:
		return FILTER_LAYER3
	}
	return 0
}

func getFilterFor(fh *mp3agic.FrameHeader) uint32 {
	return getMpegFilter(fh) | getModeFilter(fh) | getSamplingrateFilter(fh) | getLayerFilter(fh)
}

// TODO: update the comment
/**
 * tries to find the next MPEG Audio Frame, loads it into the destination
 * buffer (including 32bit header) and returns a FrameHeader object. (If
 * destFH is non-null, that object will be used to store the header infos)
 * will block until data is available. will throw EOFException of any other
 * IOException created by the InputStream object. set filter to 0 or to any
 * other value using the FILTER_xxx flags to force a specific frame type.
 */
func (p *mpaFrameParser) getNextFrame(filter uint32, destBuffer []byte, destFH *mp3agic.FrameHeader) (*mp3agic.FrameHeader, os.Error) {
	p.setupFilter(filter)
	fill(p.headBuff[:], 0)

	hbPos := 0
	fh := destFH
	if fh == nil {
		fh = new(mp3agic.FrameHeader)
	}

	var tmp [1]byte
	skipped := -4
	for {
		readn, err := p.ips.Read(tmp[:])
		if readn == 0 && err == os.EOF { // EOF ?
			if p.junkh != nil {
				for i := 0; i < 4; i++ { // flush headBuff
					if skipped >= 0 {
						p.junkh.Write(p.headBuff[(hbPos+i)&3])
					}
					skipped++
				}
				p.junkh.EndOfJunkBlock()
			}
			return nil, os.EOF
		} else if err != os.EOF {
			return nil, err
		}
		if p.junkh != nil && skipped >= 0 {
			p.junkh.Write(p.headBuff[hbPos])
		}

		p.headBuff[hbPos] = tmp[0]
		skipped++
		hbPos = (hbPos + 1) & 3

		if p.headBuff[hbPos] != 0xFF {
			continue // not the beginning of a sync-word
		}

		header32 := uint32(p.headBuff[hbPos])
		for z := 1; z < 4; z++ {
			header32 <<= 8
			header32 |= uint32(p.headBuff[(hbPos+z)&3])
		}

		if header32&p.masker != p.masked {
			continue // not a frame header
		}

		*fh = mp3agic.FrameHeader(header32)
		if fh.Verify() != nil { // doesn't look like a proper header
			continue
		}

		if filter&FILTER_STEREO != 0 && fh.Channels() != 2 {
			continue
		}

		offs := 0
		for ; offs < 4; offs++ {
			destBuffer[offs] = p.headBuff[(hbPos+offs)&3]
		}

		tmp2 := fh.LengthInBytes() - offs
		//tmp2 := fh.getFrameSize() - offs;

		// FIXME: behaviour when not enough data read (different from Java)
		readn, err = p.ips.Read(destBuffer[offs : offs+tmp2])

		if readn != tmp2 {
			if err != os.EOF {
				panic(err)
			}
			if p.junkh != nil {
				readn += 4 // inklusive header
				for z := 0; z < readn; z++ {
					p.junkh.Write(destBuffer[z] & 0xFF)
				}
				p.junkh.EndOfJunkBlock()
			}
			return nil, err
		}

		if p.junkh != nil {
			p.junkh.EndOfJunkBlock()
		}

		break
	}
	return fh, nil
}

func (p *mpaFrameParser) setupFilter(filter uint32) {
	p.masker = 0xFFE00000
	p.masked = 0xFFE00000
	switch {
	case filter&FILTER_MPEG1 != 0:
		p.masker |= 0x00180000
		p.masked |= 0x00180000
	case filter&FILTER_MPEG2 != 0:
		p.masker |= 0x00180000
		p.masked |= 0x00100000
	}

	switch {
	case filter&FILTER_LAYER1 != 0:
		p.masker |= 0x00060000
		p.masked |= 0x00060000
	case filter&FILTER_LAYER2 != 0:
		p.masker |= 0x00060000
		p.masked |= 0x00040000
	case filter&FILTER_LAYER3 != 0:
		p.masker |= 0x00060000
		p.masked |= 0x00020000
	}

	switch {
	case filter&FILTER_32000HZ != 0:
		p.masker |= 0x00000C00
		p.masked |= 0x00000800
	case filter&FILTER_44100HZ != 0:
		p.masker |= 0x00000C00
		p.masked |= 0x00000000
	case filter&FILTER_48000HZ != 0:
		p.masker |= 0x00000C00
		p.masked |= 0x00000400
	}

	if filter&FILTER_MONO != 0 {
		p.masker |= 0x000000C0
		p.masked |= 0x000000C0
	}
}
