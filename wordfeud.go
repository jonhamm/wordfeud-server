package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/language"
)

/*
*

	wordle -all
		solve all words in the built-in word list
	wordle -secret=

*
*/

type GameOptions struct {
	help       bool
	verbose    bool
	debug      uint
	move       uint
	moveDebug  uint
	randSeed   uint64
	count      int
	name       string
	out        io.Writer
	language   language.Tag
	rand       *rand.Rand
	writeFile  bool
	file       string
	directory  string
	fileFormat FileFormat
	cmd        string
	args       []string
}

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
		-help 				show this usage info
		-verbose			increase output from execution
		-debug=dd			show debug output when dd > 0 (the larger dd is the more output)
		-move=mm			only set debug as specified by -debug after move mm has completed
		-rand=nn			seed random number generator with nn 
							0 or default will seed with timestamp
		-count=nn	        repeat count for autoplay - default is 1
		-name=xxxxx			autoplay game files will be named "xxxxx-nn" where nn is 1..count
							xxxxx default is "scrabble"
		-out=file-or-dir	the name of the file or directory to hold game result
							if -count is specified > 1 the file will be "file-or-dir-nn" where nn is 1..count
							if not specified no file will be produced
							if file-or-dir is the name of a directory the 
							html file will be "file-or-dir/xxxxx-nn" where xxxxx and nn 
							are as explained in the -name option
		-format=zzzz		the format of output file if one is produced (see -out)
							valid formats are:
								"txt": simple text file
								"json": json file
						

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
		?d=dd				show debug output when dd > 0 (the larger dd is the more output)
		?m=dd				only set debug as specified by -debug after move mm has completed
		?r=nn			    seed random number generator with nn (!= 0)
							0 or default will seed with timestamp
		?c=nn	        	repeat count for autoplay - default is 1
		?n=xxxxx			autoplay game files will be named xxxxx-nn where nn is 1..count
							xxxxx default is "scrabble"
`

func main() {
	var options GameOptions
	var languageSpec string
	var fileFormatSpec string
	var ranSeedSpec string
	options.out = os.Stdout
	options.language = language.Danish
	flag.Usage = func() { fmt.Print(usage) }
	BoolVarFlag(flag.CommandLine, &options.verbose, []string{"verbose", "v"}, false, "show more output")
	BoolVarFlag(flag.CommandLine, &options.help, []string{"help", "h"}, false, "print usage information")
	UintVarFlag(flag.CommandLine, &options.debug, []string{"debug", "d"}, 0, "increase above 0 to get debug info - more than verbose")
	UintVarFlag(flag.CommandLine, &options.move, []string{"move", "m"}, 0, "increase above 0 to get debug info - more than verbose")
	IntVarFlag(flag.CommandLine, &options.count, []string{"count", "c"}, 0, "increase above 0 to get debug info - more than verbose")
	StringVarFlag(flag.CommandLine, &ranSeedSpec, []string{"rand", "r"}, "", "seed for random number generator - 0 will seed with timestamp")
	StringVarFlag(flag.CommandLine, &languageSpec, []string{"language", "l"}, "", "the requested corpus language")
	StringVarFlag(flag.CommandLine, &options.name, []string{"name", "n"}, "", "name of game files ")
	StringVarFlag(flag.CommandLine, &options.file, []string{"out", "o"}, "", "the name of the file or directory to hold game result")
	StringVarFlag(flag.CommandLine, &fileFormatSpec, []string{"format", "f"}, "", "the format of output file")

	flag.Parse()
	args := flag.Args()
	if options.help {
		flag.Usage()
	}
	if len(args) == 0 {
		if !options.help {
			fmt.Fprintln(os.Stderr, "Please specify a subcommand. (-help for more info)")
		}
		return
	}
	if len(languageSpec) > 0 {
		tag, err := language.Default.Parse(languageSpec)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unknown language \"%s\"\n", languageSpec)
			return
		}
		options.language = tag
	}

	if len(ranSeedSpec) > 0 {
		ranSeedSpec = strings.ReplaceAll(ranSeedSpec, ",", "")
		ranSeedSpec = strings.ReplaceAll(ranSeedSpec, ".", "")
		s, err := strconv.ParseUint(ranSeedSpec, 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid random seed \"%s\" : %s\n", ranSeedSpec, err.Error())
			return
		}
		options.randSeed = s
	}
	if options.randSeed == 0 {
		options.randSeed = uint64(time.Now().UnixNano())
	}
	if options.count < 1 {
		options.count = 1
	}

	if len(options.name) == 0 {
		options.name = "scrabble"
	}

	if len(options.file) > 0 {
		info, err := os.Stat(options.file)
		if err == nil {
			if info.IsDir() {
				options.directory = options.file
				options.file = options.name
			}
		} else {
			options.directory, options.file = path.Split(options.file)
		}
		options.fileFormat = ParseFileFormat(fileFormatSpec)
		if options.fileFormat == FILE_FORMAT_NONE {
			options.fileFormat = FILE_FORMAT_TEXT
		}
		options.writeFile = true
	}

	cmd, args := args[0], args[1:]
	options.cmd = cmd
	options.args = args
	if options.debug > 0 {
		options.verbose = true
		PrintOptions(&options)
	}
	if options.debug > 2 {
		DAWG_TRACE = true
	}

	options.rand = rand.New(rand.NewSource(int64(options.randSeed)))

	switch cmd {
	case "serve":
		//serveCmd(&options, args)
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
	case "keepdebugfunction":
		debugState(nil)
		debugPlayers(nil, PlayerStates{})
		debugPlayer(nil, nil)
		debugPartialMove(nil)
		debugPartialMoves(nil)
		debugMove(nil)
		debugDawgState(nil, DawgState{})
	case "nil":

	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand '%q'.  (-help for more info)\n", cmd)
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
