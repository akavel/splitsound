/*
 * Created on 29.09.2005
 */

package de.zebee.mpa;

import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.Arrays;

/**
 * ID3v1 and ID3v1.1 implementation. If you set the track number to a value greater than zero
 * it'll be an ID3v1.1 tag with 28 comment characters at max. If you set the track number to zero
 * then it'll be an ID3v1 tag and you can use 30 characters for the comment.
 *
 * @author Sebastian Gesemann
 */
public class ID3v1 {

	public static final int TAG_SIZE = 128;

	public static final int OFS_TITLE = 3;
	public static final int OFS_ARTIST = 33;
	public static final int OFS_ALBUM = 63;
	public static final int OFS_YEAR = 93;
	public static final int OFS_COMMENT = 97;
	public static final int OFS_GENRE = 127;

	public static final int LEN_TITLE = 30;
	public static final int LEN_ARTIST = 30;
	public static final int LEN_ALBUM = 30;
	public static final int LEN_YEAR = 4;
	public static final int LEN_COMMENT = 30;
	public static final int LEN_GENRE = 1;

	//      0123456789ABCDEF
	// ---+-----------------
	//	4x |  ABCDEFGHIJKLMNO
	// 5x | PQRSTUVWXYZ

	private byte[] data = null;
	private String titleCache = null;
	private String artistCache = null;
	private String albumCache = null;
	private String yearCache = null;
	private String commentCache = null;
	
	public ID3v1() {
	}

	public boolean isValid() {
		return data!=null;
	}

	private static boolean tagPresent(byte[] src, int ofs) {
		return src[ofs]==0x54 && src[ofs+1]==0x41 && src[ofs+2]==0x47;
	}

	public void readFrom(byte[] src, int ofs) {
		titleCache = null;
		artistCache = null;
		albumCache = null;
		yearCache = null;
		commentCache = null;
		if (tagPresent(src,ofs)) {
			data = new byte[TAG_SIZE];
			System.arraycopy(src,ofs,data,0,TAG_SIZE);
		} else {
			data = null;
		}
	}

	public void readFrom(InputStream ips) throws IOException {
		titleCache = null;
		artistCache = null;
		albumCache = null;
		yearCache = null;
		commentCache = null;
		alloc();
		int l=0;
		while (l<TAG_SIZE) {
			int rr = ips.read(data,l,TAG_SIZE-l);
			if (rr<0) throw new EOFException();
			l += rr;
		}
		if (!tagPresent(data,0)) {
			data = null;
		}
	}

	public void writeTo(byte[] dst, int ofs) {
		if (data==null) throw new NullPointerException();
		System.arraycopy(data,0,dst,ofs,TAG_SIZE);
	}

	public void writeTo(OutputStream ops) throws IOException {
		if (data==null) throw new NullPointerException();
		ops.write(data);
	}

	private String getString(int ofs, int maxlen) {
		int l = 0;
		while (l<maxlen && data[ofs+l]!=0) l++;
		return (new String(data,ofs,l)).trim();
	}

	public String getTitle() {
		if (titleCache==null) {
			if (data==null) return null;
			titleCache = getString(OFS_TITLE,LEN_TITLE);
		}
		return titleCache;
	}

	public String getArtist() {
		if (artistCache==null) {
			if (data==null) return null;
			artistCache = getString(OFS_ARTIST,LEN_ARTIST);
		}
		return artistCache;
	}

	public String getAlbum() {
		if (albumCache==null) {
			if (data==null) return null;
			albumCache = getString(OFS_ALBUM,LEN_ALBUM);
		}
		return albumCache;
	}

	public String getYear() {
		if (yearCache==null) {
			if (data==null) return null;
			yearCache = getString(OFS_YEAR,LEN_YEAR);
		}
		return yearCache;
	}

	public String getComment() {
		if (commentCache==null) {
			if (data==null) return null;
			commentCache = getString(OFS_COMMENT,LEN_COMMENT);
		}
		return commentCache;
	}

	public int getTrackNum() {
		int o1 = OFS_COMMENT+LEN_COMMENT-2;
		int o2 = OFS_COMMENT+LEN_COMMENT-2;
		if (data[o1]==0) {
			return data[o2] & 0xFF;
		} else {
			return 0;
		}
	}

	public int getGenre() {
		return (data[OFS_GENRE] & 0xFF);
	}

	private void alloc() {
		if (data==null) {
			data = new byte[128];
			data[0] = 0x54;
			data[1] = 0x41;
			data[2] = 0x47;
		}
	}

	private boolean putString(String s, int ofs, int maxlen) {
		alloc();
		byte[] ba = s.getBytes();
		Arrays.fill(data,ofs,ofs+maxlen,(byte)0);
		int toCopy = Math.min(maxlen,ba.length);
		for (int i=0; i<toCopy; i++) {
			int bite = ba[i];
			if (bite==0) bite=0x20;
			data[ofs+i] = (byte) bite;
		}
		return toCopy<=maxlen;
	}

	public boolean setTitle(String title) {
		titleCache = title;
		return putString(title,OFS_TITLE,LEN_TITLE);
	}

	public boolean setArtist(String artist) {
		artistCache = artist;
		return putString(artist,OFS_ARTIST,LEN_ARTIST);
	}

	public boolean setAlbum(String album) {
		albumCache = album;
		return putString(album,OFS_ALBUM,LEN_ALBUM);
	}

	public boolean setYear(String year) {
		yearCache = year;
		return putString(year,OFS_YEAR,LEN_YEAR);
	}

	public boolean setComment(String comment) {
		commentCache = comment;
		final int actualLen = getTrackNum()==0 ? LEN_COMMENT : LEN_COMMENT-2;
		return putString(comment,OFS_COMMENT,actualLen);
	}

	public void setTrackNum(int trackNo) {
		alloc();
		if (getTrackNum()>0 && trackNo==0) {
			data[OFS_COMMENT+LEN_COMMENT-1] = 0;
			// rewrite comment and possibly restore 2 lost characters
			setComment(commentCache);
		} else {
			if (trackNo>0 && getTrackNum()==0) {
				data[OFS_COMMENT+LEN_COMMENT-2] = 0;
			}
			data[OFS_COMMENT+LEN_COMMENT-1] = (byte)trackNo;
		}
	}

	public void setGenre(int bite) {
		data[OFS_GENRE] = (byte) bite;
	}

}
