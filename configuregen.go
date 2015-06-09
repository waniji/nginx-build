package main

import (
	"fmt"
	"runtime"
	"strings"
)

type StaticLibrary struct {
	Name    string
	Version string
	Option  string
}

func makeStaticLibrary(builder *Builder) StaticLibrary {
	return StaticLibrary{
		Name:    builder.name(),
		Version: builder.Version,
		Option:  builder.option()}
}

func configureGenModule3rd(modules3rd []Module3rd) string {
	result := ""
	for _, m := range modules3rd {
		if m.Form == "local" {
			result += fmt.Sprintf("--add-module=%s \\\n", m.Url)
		} else {
			result += fmt.Sprintf("--add-module=../%s \\\n", m.Name)
		}
	}
	return result
}

func configureGen(configure string, modules3rd []Module3rd, dependencies []StaticLibrary, options ConfigureOptions, rootDir string) string {
	openSSLStatic := false
	if len(configure) == 0 {
		configure = `#!/bin/sh

./configure \
`
		if runtime.GOOS == "darwin" {
			configure += "--with-cc-opt=\"-Wno-deprecated-declarations\" \\"
		}
	}

	for _, d := range dependencies {
		configure += fmt.Sprintf("%s=../%s-%s \\\n", d.Option, d.Name, d.Version)
		if d.Name == "openssl" {
			openSSLStatic = true
		}
	}

	if openSSLStatic && !strings.Contains(configure, "--with-http_ssl_module") {
		configure += "--with-http_ssl_module \\\n"
	}

	configure_modules3rd := configureGenModule3rd(modules3rd)
	configure += configure_modules3rd

	for _, option := range options.Values {
		if *option.Value != "" {
			if option.Name == "--add-module" {
				configure += normalizeAddModulePaths(*option.Value, rootDir)
			} else {
				if strings.Contains(*option.Value, " ") {
					configure += option.Name + "=" + "'" + *option.Value + "'" + " \\\n"
				} else {
					configure += option.Name + "=" + *option.Value + " \\\n"
				}
			}
		}
	}

	for _, option := range options.Bools {
		if *option.Enabled {
			configure += option.Name + " \\\n"
		}
	}

	return configure
}
