package id3v2

import (
	"fmt"
	"io"
	"os"
)

type Frame struct {
	Header [10]byte
	Data   []byte
}

// TODO: obsolete frame format (ID3v2ObseleteFrame.java)
func extractFrame(mp3stream io.ReadSeeker) (*Frame, os.Error) {
	frame := Frame{}

	err := readStream(mp3stream, frame.Header[:])
	if err != nil {
		return nil, err
	}

	err = frame.ValidateHeader()
	if err != nil {
		return nil, err
	}

	frame.Data = make([]byte, frame.DataLength())
	err = readStream(mp3stream, frame.Data)
	if err != nil {
		return nil, err
	}

	return &frame, nil
}

func (frame *Frame) ValidateHeader() os.Error {
	id := frame.Id()
	for i := 0; i < len(id); i++ {
		c := int(id[i])
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return os.NewError("invalid ID3v2 frame tag: " + id)
		}
	}
	return nil
}

func (frame *Frame) Id() string {
	return string(frame.Header[0:4])
}

func (frame *Frame) DataLength() int {
	return int(unpackInteger(frame.Header[4:8]))
}

func (frame *Frame) Length() int {
	return frame.DataLength() + len(frame.Header)
}

func unpackInteger(b4 []byte) int32 {
	return int32(b4[0])<<24 + int32(b4[1])<<16 + int32(b4[2])<<8 + int32(b4[3])
}

// TODO: unsynch=true, encoding!=0, stop reading on '\0'
func textDecode(encoding byte, data []byte, unsynch bool) (string, os.Error) {
	switch encoding {
	case 0:
		return string(data), nil
	}
	return "", os.NewError(fmt.Sprintf("unknown ID3v2 encoding %v", encoding))
}

type pictureData struct {
	MimeType    string
	PictureType byte
	Description string
	ImageData   []byte
}

//TODO: check boundaries [? or does Go do it?]
//TODO: handle error from textDecode
func pictureUnpack(buf []byte) *pictureData {
	if buf == nil {
		return nil
	}
	data := new(pictureData)

	enc := buf[0]

	x, buf := splitOnZero(buf[1:])
	data.MimeType = string(x)
	data.PictureType = buf[0]

	x, buf = splitOnZero(buf[1:])
	data.Description, _ = textDecode(enc, x, false)

	x = buf
	data.ImageData = make([]byte, len(x))
	copy(data.ImageData, x)

	return data
}

type urlData struct {
	Description string
	Url         string
}

//TODO: check boundaries [? or does Go do it?]
//TODO: handle error from textDecode
func urlUnpack(buf []byte) *urlData {
	if buf == nil {
		return nil
	}
	data := new(urlData)

	enc := buf[0]

	x, buf := splitOnZero(buf[1:])
	data.Description, _ = textDecode(enc, x, false)

	x, _ = splitOnZero(buf)
	data.Url = string(x)

	return data
}

type commentData struct {
	Language    string
	Description string
	Comment     string
}

func commentUnpack(buf []byte) *commentData {
	if buf == nil {
		return nil
	}
	d := new(commentData)

	enc := buf[0]
	d.Language = string(buf[1:4])

	x, buf := splitOnZero(buf[4:])
	d.Description, _ = textDecode(enc, x, false)

	x, _ = splitOnZero(buf)
	d.Comment, _ = textDecode(enc, x, false)

	return d
}

func splitOnZero(buf []byte) (head, tail []byte) {
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			return buf[:i], buf[i+1:]
		}
	}
	return buf, nil
}
