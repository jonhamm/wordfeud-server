package main

import (
	"flag"
	"fmt"
	"net/http"
	"strconv"
	. "wordfeud/context"

	"golang.org/x/text/language"
)

type Server struct {
	options        *GameOptions
	serviceOptions *GameOptions
}

type serverEndpoint func(server *Server, w http.ResponseWriter, req *http.Request)
type endpointFunc func(w http.ResponseWriter, req *http.Request)

func endpointWrapper(server *Server, f serverEndpoint) endpointFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		server.serviceOptions = server.options.Copy()
		server.serviceOptions.Out = w

		query := req.URL.Query()
		if s, ok := query["v"]; ok {
			v, err := strconv.Atoi(s[0])
			if err == nil {
				server.serviceOptions.Verbose = v > 0
			}
		}
		if s, ok := query["d"]; ok {
			d, err := strconv.Atoi(s[0])
			if err == nil {
				server.serviceOptions.Debug = uint(d)
			}
		}
		if s, ok := query["l"]; ok {
			tag, err := language.Default.Parse(s[0])
			if err == nil {
				server.serviceOptions.Language = tag
			}
		}
		if s, ok := query["r"]; ok {
			r, err := strconv.ParseUint(s[0], 10, 64)
			if err == nil {
				server.serviceOptions.RandSeed = r
			}
		}
		if s, ok := query["n"]; ok {
			server.serviceOptions.Name = s[0]
		}
		f(server, w, req)
	}
}

func _hello(server *Server, w http.ResponseWriter, req *http.Request) {

	fmt.Fprintf(w, "hello\n")
	if server.serviceOptions.Verbose {
		fmt.Fprintf(w, "options: %+v\n", server.serviceOptions)
	}
}

func _headers(server *Server, w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
	if server.serviceOptions.Verbose {
		fmt.Fprintf(w, "options: %+v\n", server.serviceOptions)
	}
}

func serveCmd(options *GameOptions, args []string) {
	flag := flag.NewFlagSet("serve", flag.ExitOnError)
	flag.Usage = func() { fmt.Fprint(options.Out, httpUsage) }

	var port int
	registerGlobalFlags(flag)
	IntVarFlag(flag, &port, []string{"port", "p"}, 6789, "the port on which the http server listens")
	options.Help = false
	flag.Parse(args)
	//args = flag.Args()
	if options.Debug > 0 {
		options.Verbose = true
		fmt.Fprintf(options.Out, "options: %+v\n", options)
	}
	if options.Help {
		flag.Usage()
		return
	}

	server := &Server{options, nil}

	http.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("./www"))))
	http.HandleFunc("/hello/", endpointWrapper(server, _hello))
	http.HandleFunc("/headers/", endpointWrapper(server, _headers))
	http.HandleFunc("/scrabble/", endpointWrapper(server, scrabbleWWW))

	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
