package main

import (
	"flag"
	"strings"
	"time"
)

type Config struct {
	Policy    string
	Addresses StringSet
	Includes  StringSet
	Timeout   time.Duration
	Chunk     int
}

type StringSet []string

func (s *StringSet) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSet) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var config Config

func init() {
	flag.StringVar(&config.Policy, "p", "policy.yaml", "policy file to upload")
	flag.Var(&config.Addresses, "s", "server(s) to upload policy to")
	flag.Var(&config.Includes, "i", "included content to upload")
	flag.DurationVar(&config.Timeout, "t", 5*time.Second, "connection timeout")
	flag.IntVar(&config.Chunk, "c", 64*1024, "size of chunk for splitting uploads")

	flag.Parse()
}
