package dup_packages

import "github.com/nkovacs/counterfeiter/fixtures/dup_packages/a/v1"

type DupA interface {
	A() v1.S
}
