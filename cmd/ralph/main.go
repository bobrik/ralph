package main

import (
	"flag"
	"log"
	"strings"

	"github.com/bobrik/marathoner"
	"github.com/bobrik/ralph"
)

func main() {
	u := flag.String("u", "127.0.0.1:7676", "updater location")
	b := flag.String("b", "127.0.0.1", "ip address to bind")
	c := flag.String("c", "", "nutcracker config path")
	a := flag.String("a", "", "extra args for nutcracker")
	flag.Parse()

	if *c == "" {
		flag.PrintDefaults()
		return
	}

	log.Printf("%#v\n", strings.Split(*a, " "))

	t := ralph.NewTwemproxyConfigurator(*c, *b, strings.Split(*a, " "))
	l := marathoner.NewListener(strings.Split(*u, ","), t)
	l.Start()
}
