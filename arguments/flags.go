package arguments

import (
	"flag"
)

var (
	fakeNameFlag = flag.String(
		"fake-name",
		"",
		"The name of the fake struct",
	)

	outputPathFlag = flag.String(
		"o",
		"",
		"The file or directory to which the generated fake will be written",
	)

	packageNameFlag = flag.String(
		"pkg",
		"",
		"The package name of the generated file",
	)

	packageFlag = flag.Bool(
		"p",
		false,
		"whether or not to generate a package shim",
	)
)
