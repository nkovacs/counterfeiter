#!/bin/sh

go list -f '{{ join .Imports "\n"}}{{"\n"}}{{ join .TestImports "\n" }}{{"\n"}}{{ join .XTestImports "\n"}}' ./... | grep -v "github.com/nkovacs/counterfeiter" | xargs go get -v
