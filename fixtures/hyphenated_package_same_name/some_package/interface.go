package some_package

import "github.com/nkovacs/counterfeiter/fixtures/hyphenated_package_same_name/hyphen-ated/some_package"

type SomeInterface interface {
	CreateThing() some_package.Thing
}
