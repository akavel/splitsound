package mp3agic

import (
	"io"
	"os"
	"strconv"
)

type Id3v1Tag [128]byte

const (
	Id3v1_length = 128
	id3v1_magic  = "TAG"
)

func ExtractId3v1Tag(mp3stream io.ReadSeeker) (*Id3v1Tag, os.Error) {
	var tag Id3v1Tag

	_, err := mp3stream.Seek(int64(-len(tag)), 2) // at end of file
	if err != nil {
		return nil, err
	}

	readn, err := mp3stream.Read(tag[:])
	if err != nil {
		return nil, err
	}
	if readn != len(tag) {
		return nil, os.NewError("stream too short for ID3v1 tag")
	}

	if !tag.Valid() {
		return nil, os.NewError("stream does not contain ID3v1 magic code")
	}
	return &tag, nil
}

func (tag *Id3v1Tag) Valid() bool {
	return string(tag[:len(id3v1_magic)]) == id3v1_magic
}

func (tag *Id3v1Tag) Track() string {
	if !tag.hasTrack() {
		return ""
	}
	track := tag[126]
	if track == 0 {
		return ""
	}
	return strconv.Itoa(int(track))
}

func (tag *Id3v1Tag) Artist() string {
	return tag.substring(33, 30)
}

func (tag *Id3v1Tag) Title() string {
	return tag.substring(3, 30)
}

func (tag *Id3v1Tag) Album() string {
	return tag.substring(63, 30)
}

func (tag *Id3v1Tag) Year() string {
	return tag.substring(93, 4)
}

func (tag *Id3v1Tag) Genre() int {
	genre := tag[127] & 0xff
	if genre == 0xff {
		return -1
	}
	return int(genre)
}

func (tag *Id3v1Tag) GenreDescription() string {
	genre := tag.Genre()
	if genre < 0 || genre >= len(id3v1_genres) {
		return "Unknown"
	}
	return id3v1_genres[genre]
}

func (tag *Id3v1Tag) Comment() string {
	if tag.hasTrack() {
		return tag.substring(97, 28)
	}
	return tag.substring(97, 30)
}

func (tag *Id3v1Tag) Version() string {
	if tag.hasTrack() {
		return "1"
	}
	return "0"
}

func (tag *Id3v1Tag) substring(offset, length int) string {
	pos := offset + length - 1
	for ; pos >= offset; pos-- {
		if tag[pos] > 32 {
			break
		}
	}
	if pos < offset {
		return ""
	}
	return string(tag[offset : pos+1])
}

func (tag *Id3v1Tag) hasTrack() bool {
	return tag[125] == 0
}

var id3v1_genres = [...]string{
	"Blues",
	"Classic Rock",
	"Country",
	"Dance",
	"Disco",
	"Funk",
	"Grunge",
	"Hip-Hop",
	"Jazz",
	"Metal",
	"New Age",
	"Oldies",
	"Other",
	"Pop",
	"R&B",
	"Rap",
	"Reggae",
	"Rock",
	"Techno",
	"Industrial",
	"Alternative",
	"Ska",
	"Death Metal",
	"Pranks",
	"Soundtrack",
	"Euro-Techno",
	"Ambient",
	"Trip-Hop",
	"Vocal",
	"Jazz+Funk",
	"Fusion",
	"Trance",
	"Classical",
	"Instrumental",
	"Acid",
	"House",
	"Game",
	"Sound Clip",
	"Gospel",
	"Noise",
	"Alt Rock",
	"Bass",
	"Soul",
	"Punk",
	"Space",
	"Meditative",
	"Instrumental Pop",
	"Instrumental Rock",
	"Ethnic",
	"Gothic",
	"Darkwave",
	"Techno-Industrial",
	"Electronic",
	"Pop-Folk",
	"Eurodance",
	"Dream",
	"Southern Rock",
	"Comedy",
	"Cult",
	"Gangsta",
	"Top 40",
	"Christian Rap",
	"Pop/Funk",
	"Jungle",
	"Native American",
	"Cabaret",
	"New Wave",
	"Psychedelic",
	"Rave",
	"Showtunes",
	"Trailer",
	"Lo-Fi",
	"Tribal",
	"Acid Punk",
	"Acid Jazz",
	"Polka",
	"Retro",
	"Musical",
	"Rock & Roll",
	"Hard Rock",
	"Folk",
	"Folk/Rock",
	"National Folk",
	"Swing",
	"Fast Fusion",
	"Bebob",
	"Latin",
	"Revival",
	"Celtic",
	"Bluegrass",
	"Avantgarde",
	"Gothic Rock",
	"Progressive Rock",
	"Psychedelic Rock",
	"Symphonic Rock",
	"Slow Rock",
	"Big Band",
	"Chorus",
	"Easy Listening",
	"Acoustic",
	"Humour",
	"Speech",
	"Chanson",
	"Opera",
	"Chamber Music",
	"Sonata",
	"Symphony",
	"Booty Bass",
	"Primus",
	"Porn Groove",
	"Satire/Parody",
	"Slow Jam",
	"Club",
	"Tango",
	"Samba",
	"Folklore",
	"Ballad",
	"Power Ballad",
	"Rhythmic Soul",
	"Freestyle",
	"Duet",
	"Punk Rock",
	"Drum Solo",
	"Acapella",
	"Euro-House",
	"Dance Hall",
	"Goa",
	"Drum & Bass",
	"Club-House",
	"Hardcore",
	"Terror",
	"Indie",
	"BritPop",
	"Negerpunk",
	"Polsk Punk",
	"Beat",
	"Christian Gangsta",
	"Heavy Metal",
	"Thrash Metal",
	"Anime",
	"JPop",
	"Synthpop"}
