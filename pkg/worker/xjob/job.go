package xjob

import "copy/pkg/flag"

func init() {
	flag.Register(
		&flag.StringFlag{
			Name:    "job",
			Usage:   "--job",
			Default: "",
		},
	)
}

type Runner interface {
	Run()
}
