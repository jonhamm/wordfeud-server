package context

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strings"

	"golang.org/x/text/language"
)

type Options struct {
	Verbose bool
	Debug   uint
}

type GameOptions struct {
	Options
	Help       bool
	Move       uint
	MoveDebug  uint
	RandSeed   uint64
	Count      int
	Name       string
	Out        io.Writer
	Language   language.Tag
	Rand       *rand.Rand
	WriteFile  bool
	File       string
	Directory  string
	FileFormat FileFormat
	Cmd        string
	Args       []string
}

type FileFormat byte

const (
	FILE_FORMAT_NONE  = FileFormat(0)
	FILE_FORMAT_TEXT  = FileFormat(1)
	FILE_FORMAT_JSON  = FileFormat(2)
	FILE_FORMAT_HTML  = FileFormat(3)
	FILE_FORMAT_WWW   = FileFormat(4)
	FILE_FORMAT_DEBUG = FileFormat(5)
)

func (format FileFormat) Extension() string {
	switch format {
	case FILE_FORMAT_NONE:
		return ""
	case FILE_FORMAT_TEXT, FILE_FORMAT_DEBUG:
		return ".txt"
	case FILE_FORMAT_JSON:
		return ".json"
	case FILE_FORMAT_HTML:
	case FILE_FORMAT_WWW:
		return ""
	}
	panic(fmt.Sprintf("illegal FileFormat %d (FileFormat.Extension)", format))
}

func ParseFileFormat(formatSpec string) FileFormat {
	switch strings.ToLower(formatSpec) {
	case "txt", "text":
		return FILE_FORMAT_TEXT
	case "json", "jsn":
		return FILE_FORMAT_JSON
	case "html":
		return FILE_FORMAT_HTML
	case "www":
		return FILE_FORMAT_WWW
	case "dbg", "debug":
		return FILE_FORMAT_DEBUG
	}
	return FILE_FORMAT_NONE
}

func (format FileFormat) String() string {

	switch format {
	case FILE_FORMAT_NONE:
		return "none"
	case FILE_FORMAT_TEXT:
		return "text"
	case FILE_FORMAT_JSON:
		return "json"
	case FILE_FORMAT_HTML:
		return "html"
	case FILE_FORMAT_WWW:
		return "www"
	case FILE_FORMAT_DEBUG:
		return "debug"
	}
	panic(fmt.Sprintf("illegal FileFormat %d (FileFormat.String)", format))

}

func (options *GameOptions) Print(args ...string) {
	options.Fprint(os.Stdout, args...)
}

func (options *GameOptions) Copy() *GameOptions {
	args := slices.Clone(options.Args)
	opt := Options{Verbose: options.Verbose, Debug: options.Debug}
	copy := &GameOptions{
		Options:    opt,
		Help:       options.Help,
		Move:       options.Move,
		MoveDebug:  options.MoveDebug,
		RandSeed:   options.RandSeed,
		Count:      options.Count,
		Name:       options.Name,
		Out:        options.Out,
		Language:   options.Language,
		Rand:       options.Rand,
		WriteFile:  options.WriteFile,
		File:       options.File,
		Directory:  options.Directory,
		FileFormat: options.FileFormat,
		Cmd:        options.Cmd,
		Args:       args,
	}
	return copy
}

func (options *GameOptions) Fprint(f io.Writer, args ...string) {
	indent := ""
	if len(args) > 0 {
		indent = args[0]
	}
	fmt.Fprintf(f, "%sGameOptions:\n", indent)
	fmt.Fprintf(f, "%s   cmd:         %s\n", indent, options.Cmd)
	fmt.Fprintf(f, "%s   args:        %v\n", indent, options.Args)
	fmt.Fprintf(f, "%s   Help:        %v\n", indent, options.Help)
	fmt.Fprintf(f, "%s   Verbose:     %v\n", indent, options.Verbose)
	fmt.Fprintf(f, "%s   Debug:       %v\n", indent, options.Debug)
	fmt.Fprintf(f, "%s   move:        %v\n", indent, options.Move)
	fmt.Fprintf(f, "%s   MoveDebug:   %v\n", indent, options.MoveDebug)
	fmt.Fprintf(f, "%s   ranSeed:     %v\n", indent, options.RandSeed)
	fmt.Fprintf(f, "%s   count:       %v\n", indent, options.Count)
	fmt.Fprintf(f, "%s   name:        %s\n", indent, options.Name)
	fmt.Fprintf(f, "%s   language:    %s\n", indent, options.Language.String())
	fmt.Fprintf(f, "%s   writeFile:   %v\n", indent, options.WriteFile)
	fmt.Fprintf(f, "%s   directory:   %s\n", indent, options.Directory)
	fmt.Fprintf(f, "%s   file:        %s\n", indent, options.File)
	fmt.Fprintf(f, "%s   fileFormat:  %s\n", indent, options.FileFormat.String())
}
