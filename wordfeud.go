package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
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
	help     bool
	verbose  bool
	debug    int
	randSeed uint64
	out      io.Writer
	language language.Tag
	rand     *rand.Rand
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

	wordfeud {options} game 
    	play game automatically 

	options:	
		-help 				show this usage info
		-verbose			increase output from execution
		-debug=dd			show debug output when dd > 0 (the larger dd is the more output)
		-rand=nn			seed random number generator with nn 
							0 or default will seed with timestamp

	abbreviated options:
		-h		-help
		-v		-verbose
		-d		-debug
`

const httpUsage = `
	options:	
		?h=1 				show this usage info
		?v=1				increase output from execution
		?d=dd				show debug output when dd > 0 (the larger dd is the more output)
		?r=nn			    seed random number generator with nn (!= 0)
							0 or default will seed with timestamp
`

func main() {
	var options GameOptions
	var languageSpec string
	options.out = os.Stdout
	options.language = language.Danish
	flag.Usage = func() { fmt.Fprint(options.out, usage) }
	BoolVarFlag(flag.CommandLine, &options.verbose, []string{"verbose", "v"}, false, "show more output")
	BoolVarFlag(flag.CommandLine, &options.help, []string{"help", "h"}, false, "print usage information")
	IntVarFlag(flag.CommandLine, &options.debug, []string{"debug", "d"}, 0, "increase above 0 to get debug info - more than verbose")
	Uint64VarFlag(flag.CommandLine, &options.randSeed, []string{"rand", "r"}, 0, "seed for random number generator - 0 will seed with timestamp")
	StringVarFlag(flag.CommandLine, &languageSpec, []string{"language", "l"}, "", "the requested corpus language")

	flag.Parse()
	args := flag.Args()
	if options.help {
		flag.Usage()
	}
	if len(args) == 0 {
		if !options.help {
			fmt.Fprintln(options.out, "Please specify a subcommand. (-help for more info)")
		}
		return
	}
	if len(languageSpec) > 0 {
		tag, err := language.Default.Parse(languageSpec)
		if err != nil {
			fmt.Fprintf(options.out, "unknown language \"%s\"\n", languageSpec)
			return
		}
		options.language = tag
	}

	cmd, args := args[0], args[1:]
	if options.debug > 0 {
		options.verbose = true
		fmt.Fprintf(options.out, "cmd: %v\noptions: %+v\n", cmd, options)
	}
	if options.debug > 2 {
		DAWG_TRACE = true
	}

	if options.randSeed == 0 {
		options.randSeed = uint64(time.Now().UnixNano())
	}
	options.rand = rand.New(rand.NewSource(int64(options.randSeed)))

	switch cmd {
	case "serve":
		//serveCmd(&options, args)
	case "corpus":
		result := corpusCmd(&options, args)
		fmt.Fprint(options.out, strings.Join(result.Log, "\n"))
	case "dawg":
		result := dawgCmd(&options, args)
		fmt.Fprint(options.out, strings.Join(result.Log, "\n"))
	case "game":
		result := gameCmd(&options, args)
		fmt.Fprint(options.out, strings.Join(result.Log, "\n"))
	case "autoplay":
		result := autoplayCmd(&options, args)
		fmt.Fprint(options.out, strings.Join(result.Log, "\n"))
	default:
		fmt.Fprintf(options.out, "unknown subcommand '%q'.  (-help for more info)\n", cmd)
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

// IntVarFlag defines an int flag with specified names, default value, and usage string.
// The argument p points to an int variable in which to store the value of the flag.
func Uint64VarFlag(f *flag.FlagSet, p *uint64, names []string, value uint64, usage string) {
	for _, name := range names {
		f.Uint64Var(p, name, value, usage)
	}
}
