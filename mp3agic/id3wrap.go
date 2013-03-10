package mp3agic

import (
	"mp3agic/id3v2"
)

type Id3Wrapper struct {
	Tagv1 *Id3v1Tag
	Tagv2 *id3v2.Tag
}

func (w *Id3Wrapper) Track() string {
	if w.Tagv2 != nil && w.Tagv2.Track() != "" {
		return w.Tagv2.Track()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Track()
	}
	return ""
}

func (w *Id3Wrapper) Artist() string {
	if w.Tagv2 != nil && w.Tagv2.Artist() != "" {
		return w.Tagv2.Artist()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Artist()
	}
	return ""
}

func (w *Id3Wrapper) Title() string {
	if w.Tagv2 != nil && w.Tagv2.Title() != "" {
		return w.Tagv2.Title()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Title()
	}
	return ""
}

func (w *Id3Wrapper) Album() string {
	if w.Tagv2 != nil && w.Tagv2.Album() != "" {
		return w.Tagv2.Album()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Album()
	}
	return ""
}

func (w *Id3Wrapper) Year() string {
	if w.Tagv2 != nil && w.Tagv2.Year() != "" {
		return w.Tagv2.Year()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Year()
	}
	return ""
}

func (w *Id3Wrapper) Genre() int {
	if w.Tagv1 != nil && w.Tagv1.Genre() != -1 {
		return w.Tagv1.Genre()
	}
	if w.Tagv2 != nil {
		return w.Tagv2.Genre()
	}
	return -1
}

func (w *Id3Wrapper) GenreDescription() string {
	if w.Tagv1 != nil {
		return w.Tagv1.GenreDescription()
	}
	if w.Tagv2 != nil {
		return w.Tagv2.GenreDescription()
	}
	return ""
}

func (w *Id3Wrapper) Comment() string {
	if w.Tagv2 != nil && w.Tagv2.Comment() != "" {
		return w.Tagv2.Comment()
	}
	if w.Tagv1 != nil {
		return w.Tagv1.Comment()
	}
	return ""
}

func (w *Id3Wrapper) Composer() string {
	if w.Tagv2 != nil {
		return w.Tagv2.Composer()
	}
	return ""
}

func (w *Id3Wrapper) OriginalArtist() string {
	if w.Tagv2 != nil {
		return w.Tagv2.OriginalArtist()
	}
	return ""
}

func (w *Id3Wrapper) Copyright() string {
	if w.Tagv2 != nil {
		return w.Tagv2.Copyright()
	}
	return ""
}

func (w *Id3Wrapper) Url() string {
	if w.Tagv2 != nil {
		return w.Tagv2.Url()
	}
	return ""
}

func (w *Id3Wrapper) Encoder() string {
	if w.Tagv2 != nil {
		return w.Tagv2.Encoder()
	}
	return ""
}

func (w *Id3Wrapper) AlbumImage() []byte {
	if w.Tagv2 != nil {
		return w.Tagv2.AlbumImage()
	}
	return nil
}

func (w *Id3Wrapper) AlbumImageMimeType() string {
	if w.Tagv2 != nil {
		return w.Tagv2.AlbumImageMimeType()
	}
	return ""
}
