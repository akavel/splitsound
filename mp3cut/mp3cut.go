package main

import (
	"flag"
	"fmt"
	"os"
)

func printferr(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, args...)
}

// Command-line arguments.
var (
	cueFilename string
	outScheme   string
	outDir      string
	srcFilename string
)

// Parse command-line.
func parseArgs() os.Error {
	flag.Usage = func() {
		printferr("USAGE: %s [OPTIONS] SOURCE-MP3-FILENAME\n", os.Args[0])
		printferr("  This tool is able to do sample granular cutting of MP3 streams via\n" +
			"  the LAME-Tag's delay/padding values. A player capable of properly\n" +
			"  interpreting the LAME-Tag is needed in order to enjoy this tool.\n" +
			"OPTIONS:\n")
		flag.PrintDefaults()
		// printferr("EXAMPLES:\n"+
		// //"  %s -cue something.cue --out \"%%n - %%t\"\n"+
		// "  %s -crop 1:0-8000,2:88.23s-3m10s largefile.mp3\n"+
		// "Originally developed by Sebastian Gesemann.\n"+
		// "Maintained by Chris Banes\n"+
		// "Go port by Mateusz Czaplinski\n",
		// //os.Args[0], 
		// os.Args[0])
		return
	}
	// flag.StringVar(&cueFilename, "cue", "", "split source mp3 via cue sheet;\n"+
	// "    mp3 source can be omitted if it's already referenced by the CUE sheet")
	flag.StringVar(&outScheme, "out", "%n. %p - %t", "specify custom naming scheme where:\n"+
		"    %s = source filename (without extension)\n"+
		"    %n = track number (leading zero)\n"+
		"    %t = track title (from CUE sheet)\n"+
		"    %p = track performer (from CUE sheet)\n"+
		"    %a = album name (from CUE sheet)")
	flag.StringVar(&outDir, "dir", ".", "specify destination directory")
	// System.out.println("  --album <albumname>      set album name (for ID3 tag)");
	// System.out.println("  --artist <artistname>    set artist name (for ID3 tag)");
	flag.Parse()

	if cueFilename == "" && len(flag.Args()) < 1 {
		// return os.NewError("file name argument or 'cue' option must be provided")
		return os.NewError("file name argument must be provided")
	}

	srcFilename = flag.Arg(0)
	return nil
}

type track struct {
	Performer   string
	Title       string
	TrackNumber int
	StartSector int64
	EndSector   int64
}

type cue struct {
	Performer   string
	Title       string
	Tracks      []track
	PathToMP3   string
	SampleCount int64
}

/*
func (c *cue) FillOutEndTrackSectors() {
    for (int i = 0; i < tracks.size() - 1; i++) {
        long endSector = tracks.get(i + 1).getStartSector();
        tracks.get(i).setEndSector(endSector);
    }
}
*/

func main() {

	exitcode := int(0)
	defer func() {
		os.Exit(exitcode)
	}()
	error := func(code int, err os.Error) {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		exitcode = code
	}

	err := parseArgs()
	if err != nil {
		error(1, err)
		flag.Usage()
		return
	}

	// TODO: buffer the file
	rawfile, err := os.Open(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		error(3, err)
		return
	}
	defer rawfile.Close()

	// // TODO: iterate args with wildcards expansion
	// mp3file, err = mp3agic.ParseFile(file, 0)
	// if err != nil {
	// error(3, err)
	// return
	// }
}
