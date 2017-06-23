package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nkovacs/counterfeiter/arguments"
	"github.com/nkovacs/counterfeiter/generator"
	"github.com/nkovacs/counterfeiter/locator"
	"github.com/nkovacs/counterfeiter/model"
	"github.com/nkovacs/counterfeiter/terminal"
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fail("%s", usage)
		return
	}

	argumentParser := arguments.NewArgumentParser(
		fail,
		cwd,
		filepath.EvalSymlinks,
		os.Stat,
		terminal.NewUI(),
	)
	parsedArgs := argumentParser.ParseArguments(args...)

	outputPath := parsedArgs.OutputPath
	destinationPackage := parsedArgs.DestinationPackageName
	explicitPackageName := parsedArgs.ExplicitPackageName

	if parsedArgs.GenerateInterfaceAndShimFromPackageDirectory {
		generateInterfaceAndShim(parsedArgs.SourcePackageDir, outputPath, destinationPackage, parsedArgs.PrintToStdOut)
	} else {
		generateFake(parsedArgs.InterfaceName, parsedArgs.SourcePackageDir, parsedArgs.ImportPath, outputPath, destinationPackage, parsedArgs.FakeImplName, explicitPackageName, parsedArgs.PrintToStdOut)
	}
}

func generateFake(interfaceName, sourcePackageDir, importPath, outputPath, destinationPackage, fakeName string, explicitPackageName, printToStdOut bool) {
	var err error
	var iface *model.InterfaceToFake
	if sourcePackageDir == "" {
		var vendorPaths []string
		vendorPaths, err = locator.VendorPathsForDirPath(".")
		if err == nil {
			iface, sourcePackageDir, err = locator.GetInterfaceAndDirFromImportPath(interfaceName, importPath, vendorPaths...)
		}
	} else {
		iface, err = locator.GetInterfaceFromFilePath(interfaceName, sourcePackageDir)
	}
	if err != nil {
		fail("%v", err)
	}

	samePackage := filepath.Dir(outputPath) == sourcePackageDir
	if samePackage && !explicitPackageName {
		destinationPackage = iface.PackageName
	}

	var code string
	code, err = generator.CodeGenerator{
		Model:       *iface,
		StructName:  fakeName,
		PackageName: destinationPackage,
		SamePackage: samePackage,
	}.GenerateFake()

	if err != nil {
		fail("%v", err)
	}

	printCode(code, outputPath, printToStdOut)
	reportDone(outputPath, fakeName)
}

func generateInterfaceAndShim(sourceDir string, outputPath string, outPackage string, printToStdOut bool) {
	var code string
	functions, err := locator.GetFunctionsFromDirectory(path.Base(sourceDir), sourceDir)
	if err != nil {
		fail("%v", err)
	}

	interfaceName := strings.ToUpper(path.Base(sourceDir))[:1] + path.Base(sourceDir)[1:]

	code, err = generator.InterfaceGenerator{
		Model:                  functions,
		Package:                sourceDir,
		DestinationInterface:   interfaceName,
		DestinationPackageName: outPackage,
	}.GenerateInterface()

	if err != nil {
		fail("%v", err)
	}
	interfacePath := path.Join(outputPath, path.Base(sourceDir)+".go")
	interfaceDir := path.Dir(interfacePath)

	printCode(code, interfacePath, printToStdOut)

	reportDone(interfacePath, interfaceName)

	sourcePackage := path.Base(sourceDir)

	iface, err := locator.GetInterfaceFromFilePath(interfaceName, interfaceDir)
	if err != nil {
		fail("%v", err)
	}

	code, err = generator.ShimGenerator{
		Model:         *iface,
		StructName:    interfaceName + "Shim",
		PackageName:   outPackage,
		SourcePackage: sourcePackage,
	}.GenerateReal()

	if err != nil {
		fail("%v", err)
	}

	realPath := path.Join(interfaceDir, outPackage+".go")

	printCode(code, realPath, printToStdOut)
	reportDone(realPath, interfaceName+"Shim")
}

func printCode(code, outputPath string, printToStdOut bool) {
	newCode, err := format.Source([]byte(code))
	if err != nil {
		fail("%v", err)
	}

	code = string(newCode)

	if printToStdOut {
		fmt.Println(code)
	} else {
		os.MkdirAll(filepath.Dir(outputPath), 0777)
		file, err := os.Create(outputPath)
		if err != nil {
			fail("Couldn't create fake file - %v", err)
		}

		_, err = file.WriteString(code)
		if err != nil {
			fail("Couldn't write to fake file - %v", err)
		}
	}
}

func reportDone(outputPath, fakeName string) {
	rel, err := filepath.Rel(cwd(), outputPath)
	if err != nil {
		fail("%v", err)
	}

	fmt.Printf("Wrote `%s` to `%s`\n", fakeName, rel)
}

func cwd() string {
	dir, err := os.Getwd()
	if err != nil {
		fail("Error - couldn't determine current working directory")
	}
	return dir
}

func fail(s string, args ...interface{}) {
	fmt.Printf(s+"\n", args...)
	os.Exit(1)
}

var usage = `
USAGE
	counterfeiter
		[-o <output-path>] [-p] [--fake-name <fake-name>]
		[<source-path>] <interface> [-]

ARGUMENTS
	source-path
		Path to the file or directory containing the interface to fake.
		In package mode (-p), source-path should instead specify the path
		of the input package; alternatively you can use the package name
		(e.g. "os") and the path will be inferred from your GOROOT.

	interface
		If source-path is specified: Name of the interface to fake.
		If no source-path is specified: Fully qualified interface path of the interface to fake.
    If -p is specified, this will be the name of the interface to generate.

	example:
		# writes "FakeStdInterface" to ./packagefakes/fake_std_interface.go
		counterfeiter package/subpackage.StdInterface

	'-' argument
		Write code to standard out instead of to a file

OPTIONS
	-o
		Path to the file or directory for the generated fakes.
		This also determines the package name that will be used.
		By default, the generated fakes will be generated in
		the package "xyzfakes" which is nested in package "xyz",
		where "xyz" is the name of referenced package.

	example:
		# writes "FakeMyInterface" to ./mySpecialFakesDir/specialFake.go
		counterfeiter -o ./mySpecialFakesDir/specialFake.go ./mypackage MyInterface

	-p
		Package mode:  When invoked in package mode, counterfeiter
		will generate an interface and shim implementation from a
		package in your GOPATH.  Counterfeiter finds the public methods
		in the package <source-path> and adds those method signatures
		to the generated interface <interface-name>.

	example:
		# generates os.go (interface) and osshim.go (shim) in ${PWD}/osshim
		counterfeiter -p os
		# now generate fake in ${PWD}/osshim/os_fake (fake_os.go)
		go generate osshim/...

	--fake-name
		Name of the fake struct to generate. By default, 'Fake' will
		be prepended to the name of the original interface. (ignored in
		-p mode)

	example:
		# writes "CoolThing" to ./mypackagefakes/cool_thing.go
		counterfeiter --fake-name CoolThing ./mypackage MyInterface
`
