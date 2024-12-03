package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	. "wordfeud/context"
	. "wordfeud/corpus"
	. "wordfeud/dawg"
	. "wordfeud/game"

	"golang.org/x/text/language"
)

/*
*

	wordle -all
		solve all words in the built-in word list
	wordle -secret=

*
*/

const usage = `
	wordfeud {options} serve {-port=pppp}
		start http server on port pppp (default is 6789)
 
	wordfeud {options} corpus 
		return corpus information

	wordfeud {options} dawg 
    	return dawg information

	wordfeud {options} game 
    	return game information

	wordfeud {options} autoplay 
    	play game automatically 

	options:	
		-Help 				show this usage info
		-Verbose			increase output from execution
		-Debug=dd			show Debug output when dd > 0 (the larger dd is the more output)
		-move=mm			only set Debug as specified by -Debug after move mm has completed
		-rand=nn			seed random number generator with nn 
							0 or default will seed with timestamp
		-count=nn	        repeat count for autoplay - default is 1
		-name=xxxxx			autoplay game files will be named "xxxxx-nn" where nn is 1..Count
							xxxxx default is "scrabble"
		-out=dir	        the name of the directory to hold game result
							if -count is specified > 1 the file will be "dir/xxxxx-nn" where xxxxx and nn 
							are as explained in the -name option
							if format is html "dir/xxxxx-nn" will be a directory holding .html files
							"index.html" and "move-ii.html" where ii is 0..number of moves in game
		-format=zzzz		the format of output file if one is produced (see -out)
							valid formats are:
								"txt": simple text file
								"debug": text file with debug info
								"json": json file
								"html": json file
						

	abbreviated options:
		-h		-help
		-v		-verbose
		-d		-debug
		-m		-move
		-r 		-rand
		-c		-count
		-n		-name
		-o		-out
		-f		-format
`

const httpUsage = `
	options:	
		?h=1 				show this usage info
		?v=1				increase output from execution
		?d=dd				show Debug output when dd > 0 (the larger dd is the more output)
		?r=nn			    seed random number generator with nn (!= 0)
							0 or default will seed with timestamp
		?n=xxxxx			autoplay game files will be named xxxxx-nn where nn is 1..Count
							xxxxx default is "scrabble"
`

func main() {
	var options GameOptions
	var languageSpec string
	var fileFormatSpec string
	var ranSeedSpec string
	options.Out = os.Stdout
	options.Language = language.Danish
	flag.Usage = func() { fmt.Print(usage) }
	BoolVarFlag(flag.CommandLine, &options.Verbose, []string{"Verbose", "v"}, false, "show more output")
	BoolVarFlag(flag.CommandLine, &options.Help, []string{"Help", "h"}, false, "print usage information")
	UintVarFlag(flag.CommandLine, &options.Debug, []string{"debug", "d"}, 0, "increase above 0 to get Debug info - more than Verbose")
	UintVarFlag(flag.CommandLine, &options.Move, []string{"move", "m"}, 0, "increase above 0 to get Debug info - more than Verbose")
	IntVarFlag(flag.CommandLine, &options.Count, []string{"count", "c"}, 0, "increase above 0 to get Debug info - more than Verbose")
	StringVarFlag(flag.CommandLine, &ranSeedSpec, []string{"rand", "r"}, "", "seed for random number generator - 0 will seed with timestamp")
	StringVarFlag(flag.CommandLine, &languageSpec, []string{"language", "l"}, "", "the requested corpus language")
	StringVarFlag(flag.CommandLine, &options.Name, []string{"name", "n"}, "", "name of game files ")
	StringVarFlag(flag.CommandLine, &options.Directory, []string{"out", "o"}, "", "the name of the file or directory to hold game result")
	StringVarFlag(flag.CommandLine, &fileFormatSpec, []string{"format", "f"}, "", "the format of output file")

	flag.Parse()
	args := flag.Args()
	if options.Help {
		flag.Usage()
	}
	if len(args) == 0 {
		if !options.Help {
			fmt.Fprintln(os.Stderr, "Please specify a subcommand. (-Help for more info)")
		}
		return
	}
	if len(languageSpec) > 0 {
		tag, err := language.Default.Parse(languageSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unknown language \"%s\"\n", languageSpec)
			return
		}
		if !SupportedLanguage(tag) {
			fmt.Fprintf(os.Stderr, "unsupported language \"%s\"\n", languageSpec)
			return
		}
		options.Language = tag
	}

	if len(ranSeedSpec) > 0 {
		ranSeedSpec = strings.ReplaceAll(ranSeedSpec, ",", "")
		ranSeedSpec = strings.ReplaceAll(ranSeedSpec, ".", "")
		s, err := strconv.ParseUint(ranSeedSpec, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid random seed \"%s\" : %s\n", ranSeedSpec, err.Error())
			return
		}
		options.RandSeed = s
	}
	if options.RandSeed == 0 {
		options.RandSeed = uint64(time.Now().UnixNano())
	}
	if options.Count < 1 {
		options.Count = 1
	}

	if len(options.Name) == 0 {
		options.Name = "scrabble"
	}

	if len(options.Directory) > 0 {
		options.File = path.Join(options.Directory, options.Name)
		options.FileFormat = ParseFileFormat(fileFormatSpec)
		if options.FileFormat == FILE_FORMAT_NONE {
			options.FileFormat = FILE_FORMAT_TEXT
		}
		options.WriteFile = true
	}

	cmd, args := args[0], args[1:]
	options.Cmd = cmd
	options.Args = args
	if options.Debug > 0 {
		options.Verbose = true
	}
	if options.Debug > 2 {
		DAWG_TRACE = true
	}

	options.Rand = rand.New(rand.NewSource(int64(options.RandSeed)))

	if options.Verbose {
		options.Print()
	}

	switch cmd {
	case "serve":
		serveCmd(&options, args)
	case "corpus":
		result := corpusCmd(&options, args)
		fmt.Print(strings.Join(result.Log, "\n"))
	case "dawg":
		result := dawgCmd(&options, args)
		fmt.Print(strings.Join(result.Log, "\n"))
	case "game":
		result := gameCmd(&options, args)
		fmt.Print(strings.Join(result.Log, "\n"))
	case "autoplay":
		result := autoplayCmd(&options, args)
		fmt.Print(strings.Join(result.Log, "\n"))
	case "keepDebugfunction":
		DebugState(nil)
		DebugPlayers(nil, PlayerStates{})
		DebugPlayer(nil, nil)
		DebugPartialMove(nil)
		DebugPartialMoves(nil)
		DebugMove(nil)
		DebugDawgState(nil, nil)
		DebugStateBoard(nil)
	case "nil":

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand '%q'.  (-Help for more info)\n", cmd)
	}
}

func registerGlobalFlags(fset *flag.FlagSet) {
	flag.VisitAll(
		func(f *flag.Flag) {
			fset.Var(f.Value, f.Name, f.Usage)
		})
}

// IntVarFlag defines an int flag with specified names, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func IntVarFlag(f *flag.FlagSet, p *int, names []string, value int, usage string) {
	for _, name := range names {
		f.IntVar(p, name, value, usage)
	}
}

// UintVarFlag defines an uint flag with specified names, default value, and usage string.
// The argument p points to an uint variable in which to store the value of the flag.
func UintVarFlag(f *flag.FlagSet, p *uint, names []string, value uint, usage string) {
	for _, name := range names {
		f.UintVar(p, name, value, usage)
	}
}

// BoolVarFlag defines a bool flag with specified names, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func BoolVarFlag(f *flag.FlagSet, p *bool, names []string, value bool, usage string) {
	for _, name := range names {
		f.BoolVar(p, name, value, usage)
	}
}

// StringVarFlag defines a string flag with specified names, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func StringVarFlag(f *flag.FlagSet, p *string, names []string, value string, usage string) {
	for _, name := range names {
		f.StringVar(p, name, value, usage)
	}
}

// Uint64VarFlag defines an uint64 flag with specified names, default value, and usage string.
// The argument p points to an uint64 variable in which to store the value of the flag.
func Uint64VarFlag(f *flag.FlagSet, p *uint64, names []string, value uint64, usage string) {
	for _, name := range names {
		f.Uint64Var(p, name, value, usage)
	}
}
