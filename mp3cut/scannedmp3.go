package main

import (
	"io"
	"mp3agic"
	"os"
)

const (
	UNKNOWN_START_SAMPLE = -(int64(1) << 42)

	// FRame REcord Size (in bytes)
	FRRES = 4 + 2 + 2 + 2
	// Frame Record Count Per Page
	FRCPP = 0x2000 / FRRES
	// Byte Buffer Size Per Page
	BBSPP = FRCPP * FRRES

	minOverlapSamplesStart = 576
	minOverlapSamplesEnd   = 1152

	MASK_ATH_KILL_NO_GAP_START = 0x7F
	MASK_ATH_KILL_NO_GAP_END   = 0xBF
)

type MyCountingJunkHandler int

func (offset *MyCountingJunkHandler) Write(bite byte) {
	(*offset)++
}
func (offset *MyCountingJunkHandler) Inc(i int) {
	(*offset) += MyCountingJunkHandler(i)
}
func (offset *MyCountingJunkHandler) EndOfJunkBlock() {}

type scannedMp3 struct {
	firstFrameHeader *mp3agic.FrameHeader
	xiltFrame        XingInfoLameTagFrame
	isVBR            bool
	avgBitrate       float32
	maxRes           int
	encDelay         int
	encPadding       int
	byteBuffers      [][]byte
	currBB           []byte
	currBBofs        int
	musicFrameCount  int
	samplesPerFrame  int
	startSample      int64
}

func newScannedMp3() *scannedMp3 {
	return &scannedMp3{encDelay: 567, encPadding: 567 * 3, startSample: UNKNOWN_START_SAMPLE}
}

func (m *scannedMp3) scan(ips io.ReadSeeker) os.Error {
	temp := make([]byte, MAX_MPAFRAME_SIZE)

	jh := new(MyCountingJunkHandler)
	mpafp := &mpaFrameParser{ips: ips, junkh: jh}

	filter := FILTER_LAYER3
	var fh *mp3agic.FrameHeader
	frameCounter := 0
	firstFrameFound := false
	isMPEG1 := false
	firstkbps := 0
	sumMusicFrameSize := 0
	//try {
	for {
		fh, err := mpafp.getNextFrame(filter, temp, fh)
		if err != nil && err != os.EOF {
			return err
		}
		frameSize := fh.LengthInBytes()
		//int frameSize = fh.getFrameSize();
		if !firstFrameFound {
			firstFrameFound = true
			firstkbps = fh.BitrateInKbps()
			m.samplesPerFrame = fh.SamplesPerFrame()
			isMPEG1 = (fh.Version() == mp3agic.MPEG_VERSION_1_0)
			if isMPEG1 {
				m.maxRes = 511
			} else {
				m.maxRes = 255
			}
			filter = getFilterFor(fh)
			m.firstFrameHeader = fh
			if m.xiltFrame.parse(temp, 0) {
				if m.xiltFrame.hasXingTag {
					m.isVBR = true
				}
				frameCounter--
				if m.xiltFrame.hasLameTag {
					m.encDelay = m.xiltFrame.encDelay
					m.encPadding = m.xiltFrame.encPadding
				}
			}
		} else {
			checkBitRate := true
			if frameCounter == 0 {
				checkBitRate = false
				// first music frame. might be a PCUT-tag
				// reservoir-filler frame
				sie := fh.SideInfoEnd()
				// a pcut frame contains its tag in the first 10
				// bytes of the
				// main data section
				pcutFrame := (sie+10 <= frameSize) && (temp[sie] == 0x50) && // P
					(temp[sie+1] == 0x43) && // C
					(temp[sie+2] == 0x55) && // U
					(temp[sie+3] == 0x54) // T
				if pcutFrame {
					// temp[sie+4] tag revision (always 0 for now)
					t := int64(temp[sie+5]) // fetch 40 bit start sample
					t = (t << 8) | int64(temp[sie+6])
					t = (t << 8) | int64(temp[sie+7])
					t = (t << 8) | int64(temp[sie+8])
					t = (t << 8) | int64(temp[sie+9])
					m.startSample = t
				} else {
					for b := range temp[fh.SideInfoStart():fh.SideInfoEnd()] {
						if b != 0 {
							checkBitRate = true
							break
						}
					}
				}
			}
			// we don't want the first "music frame" to be checked
			// if it's
			// possibly a PCUT generated reservoir frame
			if checkBitRate && fh.BitrateInKbps() != firstkbps {
				m.isVBR = true
			}
		}
		if frameCounter >= 0 {
			sumMusicFrameSize += frameSize
			accessFrameRecord(frameCounter)
			setFrameFileOfs(int(*jh))
			setFrameSize(frameSize)
			sis := fh.SideInfoStart()
			ofs := sis
			brPointer := int(temp[ofs])
			if isMPEG1 {
				brPointer = (brPointer << 1) | int((temp[ofs+1]&0x80)>>7)
			}
			setBitResPtr(brPointer)
			setMainDataSectionSize(frameSize - sis - fh.SideInfoSize())
		}
		jh.Inc(frameSize)
		frameCounter++
	}
	//}
	//catch (EOFException x) {
	//}

	m.musicFrameCount = frameCounter
	if !firstFrameFound {
		return os.NewError("no mp3 data found")
	}

	var framerate float32 = float32(m.firstFrameHeader.SampleRate()) /
		float32(m.firstFrameHeader.SamplesPerFrame())
	m.avgBitrate = (float32(sumMusicFrameSize) / float32(m.musicFrameCount)) * framerate / 125
	/*    
		} finally {
		    try {
		        ips.close();
		    }
		    catch (IOException x) {
		    }
		}
	*/
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// FIXME: make sure this has same semantics as Java Math.round()
func round(f float32) int {
	return int(f)
}

func (m *scannedMp3) crop(startSample, endSample int64, ips io.Reader, ops io.Writer) os.Error {
	//defer ips.Close() // FIXME: leave this to caller

	startSample = maxInt64(startSample, int64(-m.encDelay))
	endSample = minInt64(endSample, m.SampleCount()+int64(m.encPadding))

	maskAth := 0xff
	if startSample != 0 {
		maskAth &= MASK_ATH_KILL_NO_GAP_START
	}
	if endSample != m.SampleCount() {
		maskAth &= MASK_ATH_KILL_NO_GAP_END
	}

	firstFrameInclusive := max(0, int((startSample+int64(m.encDelay-minOverlapSamplesStart))/int64(m.samplesPerFrame)))
	lastFrameExclusive := min(m.musicFrameCount, int(
		(endSample+int64(m.encDelay+minOverlapSamplesEnd+m.samplesPerFrame-1))/int64(m.samplesPerFrame)))
	newEncDelay := m.encDelay + int(startSample-int64(firstFrameInclusive)*int64(m.samplesPerFrame))
	newEncPadding := int(int64(lastFrameExclusive-firstFrameInclusive)*
		int64(m.samplesPerFrame) - int64(newEncDelay) - (endSample - startSample))
	accessFrameRecord(firstFrameInclusive)

	needBytesFromReservoir := getBitResPtr()
	gotBytesFromReservoir := 0
	needPreFrames := 0
	for firstFrameInclusive-needPreFrames > 0 &&
		needBytesFromReservoir > gotBytesFromReservoir &&
		newEncDelay+1152 <= 4095 {
		needPreFrames++
		accessFrameRecord(firstFrameInclusive - needPreFrames)
		gotBytesFromReservoir += getMainDataSectionSize()
	}

	var resFrame []byte
	resFrameSize := 0
	firstFrameNum := firstFrameInclusive
	if needPreFrames == 0 {
		// force writing of PCUT tag frame
		needPreFrames = 1
	}
	if needPreFrames > 0 {
		firstFrameNum--
		newEncDelay += m.samplesPerFrame
		resFrame = make([]byte, MAX_MPAFRAME_SIZE)
		var newAbsStartSample int64 = startSample
		if m.startSample != UNKNOWN_START_SAMPLE {
			newAbsStartSample += m.startSample
		}
		resFrameSize = constructReservoirFrame(resFrame, m.firstFrameHeader,
			needBytesFromReservoir, newAbsStartSample)
	}

	seekTable := make([]byte, 100)
	var avgBytesPerFrame, avgBytesPerSecnd, avgkbps float32
	musiLen := 0
	{ // calculate seek table
		accessFrameRecord(firstFrameInclusive)
		ofs00 := getFrameFileOfs() - resFrameSize
		accessFrameRecord(max(0, lastFrameExclusive-1))
		ofsXX := getFrameFileOfs() + getFrameSize()
		musiLen = ofsXX - ofs00
		avgBytesPerFrame = float32(ofsXX-ofs00) /
			float32(lastFrameExclusive-firstFrameInclusive)
		avgBytesPerSecnd = avgBytesPerFrame * float32(m.firstFrameHeader.SampleRate()) /
			float32(m.firstFrameHeader.SamplesPerFrame())
		avgkbps = avgBytesPerSecnd / float32(125)
		for i := 0; i < 100; i++ {
			fidx := round(float32(firstFrameInclusive+i+1) / 101.0 *
				float32(lastFrameExclusive-firstFrameInclusive))
			accessFrameRecord(max(0, fidx))
			seekTable[i] = byte(round(float32(getFrameFileOfs()-ofs00) * float32(255) /
				float32(ofsXX-ofs00)))
		}
	}

	frameBuff := make([]byte, MAX_MPAFRAME_SIZE)
	fl := XingInfoLameTagFrame.createHeaderFrame(firstFrameHeader, isVBR, avgkbps,
		lastFrameExclusive-firstFrameNum, musiLen, 50, seektable, newEncDelay,
		newEncPadding, xiltFrame, frameBuff, 0, maskATH)
	ops.write(frameBuff, 0, fl)
	filepos := 0
	sideInfoSize := firstFrameHeader.getSideInfoSize()
	bitRes := 0

	if needPreFrames > 0 {
		reservoir = make([]byte, 511)
		if needBytesFromReservoir > 0 {
			for fi := firstFrameInclusive - needPreFrames; fi < firstFrameInclusive; fi++ {
				accessFrameRecord(fi)
				tmp := getFrameFileOfs()
				ips.skip(tmp - filepos)
				filepos = tmp
				fl = getFrameSize()
				readFully(ips, frameBuff, 0, fl)
				filepos += fl
				mdss := getMainDataSectionSize()
				if mdss >= 511 {
					copy(reservoir, frameBuff[fl-511:])
					// System.arraycopy(frameBuff, fl - 511, reservoir, 0, 511);
				} else {
					move := 511 - mdss
					copy(reservoir[:move], reservoir[511-move:])
					copy(reservoir[move:], frameBuff[fl-mdss:])
					// System.arraycopy(reservoir, 511 - move, reservoir, 0, move);
					// System.arraycopy(frameBuff, fl - mdss, reservoir, move, mdss);
				}
			}
			copy(resFrame[resFrameSize-needBytesFromReservoir:resFrameSize], reservoir[511-needBytesFromReservoir:])
			// System.arraycopy(reservoir, 511 - needBytesFromReservoir, resFrame,
			//         resFrameSize - needBytesFromReservoir, needBytesFromReservoir);
		}
		ops.write(resFrame, 0, resFrameSize)
		bitRes = needBytesFromReservoir
	}

	for fi := firstFrameInclusive; fi < lastFrameExclusive; fi++ {
		accessFrameRecord(fi)
		tmp := getFrameFileOfs()
		ips.skip(tmp - filepos)
		filepos = tmp
		fl = getFrameSize()
		readFully(ips, frameBuff, 0, fl)
		filepos += fl
		tmp = getBitResPtr()
		if tmp > bitRes {
			silenceFrame(frameBuff, 0, sideInfoSize)
		}
		ops.Write(frameBuff[:fl])
		tmp = getMainDataSectionSize()
		bitRes = min(bitRes+tmp, maxRes)
	}
}

func (m *scannedMp3) SampleCount() int64 {
	return m.musicFrameCount*m.samplesPerFrame - m.encDelay - m.encPadding
}

func accessFrameRecord(idx int) {
	page := idx / FRCPP
	if underflow := page - len(byteBuffers); underflow > 0 {
		byteBuffers = append(byteBuffers, make([][]byte, underflow)...)
	}
	// for page >= len(byteBuffers) {
	// byteBuffers.add(null);
	// }
	bb := byteBuffers[page]
	if bb == nil {
		bb = make([]byte, BBSPP)
		byteBuffers[page] = bb
	}
	currBB = bb
	currBBofs = (idx % FRCPP) * FRRES
}

func constructReservoirFrame(dest []byte, header *mp3agic.FrameHeader, minResSize int, absStartSample int64) int {
	// increase for 10-byte-header inclusion
	minResSize += 10
	h32 := header.getHeader32() | 0x00010000 // switch off CRC usage
	fh2 := &FrameHeader{}
	for bri := 1; bri <= 14; bri++ {
		h32 = (h32 & 0xFFFF0FFF) + (bri << 12)
		fh2.setHeader32(h32)
		frameSize := fh2.getFrameSize()
		sideInfoEnd := fh2.getSideInfoEnd()
		mainDataBlockSize := frameSize - sideInfoEnd
		if mainDataBlockSize >= minResSize {
			dest[0] = byte(h32 >> 24)
			dest[1] = byte(h32 >> 16)
			dest[2] = byte(h32 >> 8)
			dest[3] = byte(h32)
			fill(dest[4:sideInfoEnd], 0)
			fill(dest[sideInfoEnd:frameSize], 0x78)
			// Arrays.fill(dest, 4, sideInfoEnd, (byte) 0);
			// Arrays.fill(dest, sideInfoEnd, frameSize, (byte) 0x78);
			copy(dest[sideInfoEnd:], []byte("PCUT"))
			// dest[sideInfoEnd] = 0x50; // P
			// dest[sideInfoEnd + 1] = 0x43; // C
			// dest[sideInfoEnd + 2] = 0x55; // U
			// dest[sideInfoEnd + 3] = 0x54; // T
			copy(dest[sideInfoEnd+4:], []byte{
				// revision 0
				0,
				// absolute sample start pos
				byte(uint64(absStartSample) >> 32),
				byte(uint64(absStartSample) >> 24),
				byte(uint64(absStartSample) >> 16),
				byte(uint64(absStartSample) >> 8),
				byte(uint64(absStartSample))})
			return frameSize
		}
	}
	return -1
}

func silenceFrame(data []byte, ofs, sisize int) {
	siend := 4 + sisize
	crcProtection := ((data[ofs+1] & 1) == 0)
	if crcProtection {
		siend += 2
	}
	fill(data[ofs+4:ofs+siend], 0)
	//Arrays.fill(data, ofs + 4, ofs + siend, (byte) 0);
	if crcProtection {
		crc16 := uint16(0xFFFF)
		crc16 = CRC16.updateMPEG(crc16, data[ofs+2])
		crc16 = CRC16.updateMPEG(crc16, data[ofs+3])
		for o2 := 6; o2 < siend; o2++ { // FIXME: o2++ added by me; was a bug?
			crc16 = CRC16.updateMPEG(crc16, data[ofs+o2])
		}
		data[ofs+4] = byte(crc16 >> 8)
		data[ofs+5] = byte(crc16 & 0xff)
		//data[ofs + 4] = (byte) (crc16 >>> 8);
		//data[ofs + 5] = (byte) crc16;
	}
}

func mpegCrcUpdate(crc uint16, value uint16) uint16 {
	valueExt := uint32(value) << 8
	crcExt := uint32(crc)
	for i := 0; i < 8; i++ {
		valueExt <<= 1
		crcExt <<= 1
		if (crc^value)&0x10000 != 0 {
			crc ^= 0x8005
		}
	}
	return uint16(crc & 0xffff)
}

/*
    int i;
    value <<= 8;
    for (i = 0; i < 8; i++) {
        value <<= 1;
        crc <<= 1;
        if (((crc ^ value) & 0x10000) != 0)
            crc ^= 0x8005;
    }
    return crc & 0xFFFF;
}
*/

// FIXME: implement properly
func readFully(ips io.Reader, dest []byte, dofs int, length int) os.Error {
	for len > 0 {
		res := ips.read(dest, dofs, length)
		if res < 0 {
			return os.EOF
		}
		dofs += res
		length -= res
	}
	return nil
}

func fill(buf []byte, val byte) {
	for i := range buf {
		buf[i] = val
	}
}

func getBitResPtr() int {
	return getInt16(currBB, currBBofs+4)
}

func getMainDataSectionSize() int {
	return getInt16(currBB, currBBofs+8)
}

func getFrameFileOfs() int {
	return getInt32(currBB, currBBofs)
}

func getFrameSize() int {
	return getInt16(currBB, currBBofs+6)
}

// FIXME: make sure this handles signedness OK
func getInt16(bb []byte, ofs int) int {
	return (int(bb[ofs]) << 8) | (int(bb[ofs+1]) & 0xFF)
}

func setFrameFileOfs(offset int) {
	setInt32(currBB, currBBofs, offset)
}

func setFrameSize(fs int) {
	setInt16(currBB, currBBofs+6, fs)
}

func setBitResPtr(brptr int) {
	setInt16(currBB, currBBofs+4, brptr)
}

func setMainDataSectionSize(mdss int) {
	setInt16(currBB, currBBofs+8, mdss)
}

type XingInfoLameTagFrame struct {
	frameSize  int
	bb         []byte //= new byte[MPAFrameParser.MAX_MPAFRAME_SIZE];
	hasXingTag bool
	hasInfoTag bool
	hasLameTag bool
	lameTagOfs int // starting at VBR scale of XingTag !
	encDelay   int
	encPadding int
}

func (f *XingInfoLameTagFrame) parse(byte []data, int ofs) bool {
	// TODO: verify if 'data' is long enough to contain all the offsets
	origOfs := ofs
	fh, _ = mp3agic.NewFrameHeader(data[ofs : ofs+4])

	ofs += 4 + fh.SideInfoSize()
	hasXingTag := HasXingFrameTag(data[ofs:])
	hasInfoTag := len(data) >= ofs+4 && string(data[ofs:ofs+4]) == "Info"
	if !hasXingTag && !hasInfoTag {
		return false
	}

	ofs += 4
	f.hasXingTag = hasXingTag
	f.hasInfoTag = hasInfoTag
	f.frameSize = fh.LengthInBytes()
	f.bb = make([]byte, f.frameSize)
	copy(f.bb, data)

	flags := data[ofs+3]
	ofs += 4
	if flags&0x01 != 0 {
		ofs += 4 // skip frame count
	}
	if flags&0x02 != 0 {
		ofs += 4 // skip byte count
	}
	if flags&0x04 != 0 {
		ofs += 100 // skip seek table
	}
	if flags&0x08 != 0 {
		ofs += 4 // skip VBR scale
	}
	tagEndOfs := ofs + 0x24

	crc := uint16(0)
	for i := origOfs; i < tagEndOfs-2; i++ {
		crc = CRC16.updateLAME(crc, data[i])
	}

	f.hasLameTag = (string(data[ofs:ofs+4]) == "LAME")
	if !f.hasLameTag {
		f.hasLameTag = (string(data[ofs:ofs+4]) == "GOGO")
	}
	f.hasLameTag = f.hasLameTag || ((uint16(data[tagEndOfs-2])<<8)|uint16(data[tagEndOfs-1]))^crc == 0
	//this.lameTag |= ((((data[tagEndOfs - 2] << 8) | (data[tagEndOfs - 1] & 0xFF)) ^ crc) & 0xFFFF) == 0;
	if f.hasLameTag {
		lameTagOfs = ofs - origOfs - 4
	}

	ofs += 0x15
	t := data[ofs+1]
	encDelay = int(data[ofs]<<4) | int(t>>4)
	encPadding = (int(t&0x0F) << 8) | int(data[ofs+2])
	if !f.hasLameTag {
		if encDelay > 2880 || encPadding > 2304 {
			encDelay = 576
			encPadding = 0
		}
	}
	return true
}

func createHeaderFrame(toBeSimilar mp3agic.FrameHeader, vbr bool, kbps float32,
	frameCount, musicBytes, vbrScale int, seektable []byte,
	encDelay, encPadding int, srcTag *XingInfoLameTagFrame, dest []byte, dofs int,
	maskATH int) int {
	fh32 := toBeSimilar | 0x00010000 // disable CRC if any
	frameSize := 0
	tagOffset := 0
	{ // calculate optimal header frame size
		tmp := mp3agic.FrameHeader()
		minDist := float32(9999)
		for i := 1; i < 15; i++ {
			var th32 int = (fh32 & 0xFFFF0FFF) | (i << 12)
			tmp.setHeader32(th32)
			if tmp.getFrameSize() >= 0xC0 {
				var ikbps int = tmp.BitrateInKbps()
				var dist float32 = Math.abs(kbps - ikbps)
				if dist < minDist {
					minDist = dist
					fh32 = th32
					frameSize = tmp.getFrameSize()
					tagOffset = tmp.getSideInfoSize() + 4
				}
			}
		}
	}
	tagOffset += dofs
	fill(dest[dofs:dofs+frameSize], 0)
	//Arrays.fill(dest, dofs, dofs + frameSize, (byte) 0);
	dest[dofs] = byte(fh32 >> 24)
	dest[dofs+1] = byte(fh32 >> 16)
	dest[dofs+2] = byte(fh32 >> 8)
	dest[dofs+3] = byte(fh32)
	if vbr {
		copy(dest, "Xing")
	} else {
		copy(dest, "Info")
	}
	tagOffset += 4
	copy(dest, []byte{0, 0, 0, 0x0f})
	tagOffset += 4

	add := func(b byte) {
		dest[tagOffset] = b
		tagOffset++
	}
	add32 := func(i uint32) {
		add(byte(i >> 24))
		add(byte(i >> 16))
		add(byte(i >> 8))
		add(byte(i))
	}

	add32(uint32(frameCount))
	add32(uint32(frameSize + musicBytes))
	copy(dest[tagOffset:tagOffset+100], seektable)
	//System.arraycopy(seektable, 0, dest, tagOffset, 100);
	tagOffset += 100
	add32(uint32(vbrScale))
	if srcTag != nil && srcTag.Verify() == nil && srcTag.hasLameTag {
		copy(dest[tagOffset-4:], srcTag.bb[srcTag.lameTagOfs:srcTag.lameTagOfs+40])
		//System.arraycopy(srcTag.bb, srcTag.lameTagOfs, dest, tagOffset - 4, 40);
		tagOffset += 4
		// delete LAME's replaygain tag
		for i := 0; i < 8; i++ {
			dest[tagOffset+0x07+i] = 0
		}
		// deleting no-gap flags ...
		dest[tagOffset+0x0F] &= maskATH
	} else {
		copy(dest, "LAME")
		tagOffset += 4
	}

	encDelay = max(0, min(encDelay, 4095))
	encPadding = max(0, min(encPadding, 4095))
	tagOffset += 0x11
	// write encDelay / encPadding ...
	add(byte(uint32(encDelay) >> 4))
	add(byte(((encDelay & 0xF) << 4) | (uint32(encPadding) >> 8)))
	add(byte(encPadding))

	tagOffset += 4
	add32(uint32(frameSize + musicBytes))

	crc = uint16(0)
	for i := 0; i < 190; i++ {
		crc = CRC16.updateLAME(crc, dest[dofs+i])
	}
	tagOffset += 6
	add(byte(crc >> 8))
	add(byte(crc))
	return frameSize
}
