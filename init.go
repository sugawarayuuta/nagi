package main

import (
	"encoding/json"
	"errors"
	"nagi/semver"
	"nagi/stack"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func Init(arg string) *stack.Stacktrace {

	if arg != "" {
		name, version := semver.Split(arg)
		scope, base := path.Split(name)
		arg := path.Join(scope, "create-"+base+"@"+version)
		return Exec(arg)
	}

	pkg := map[string]interface{}{}
	opn, _ := os.Open(filepath.Join(work, "package.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(&pkg)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()
	jsconfig := map[string]interface{}{}
	opn, _ = os.Open(filepath.Join(work, "jsconfig.json"))
	if opn != nil {
		err := json.NewDecoder(opn).Decode(&jsconfig)
		if err != nil {
			return stack.Stack(err, 0)
		}
	}
	defer opn.Close()

	fields := []field{
		{pkg, "name", filepath.Base(work)},
		{pkg, "version", "0.1.0"},
		{pkg, "main", "index.js"},
		{pkg, "description", ""},
		{pkg, "scripts", map[string]interface{}{}},
		{jsconfig, "compilerOptions", map[string]interface{}{}},
		{jsconfig, "exclude", []string{"dist", "node_modules", "build", ".vscode", "tmp", "temp"}},
	}
	for index := range fields {
		if _, ok := fields[index].Map[fields[index].Key]; !ok {
			fields[index].Map[fields[index].Key] = fields[index].Val
		}
	}
	scripts := pkg["scripts"].(map[string]interface{})
	compilerOptions := jsconfig["compilerOptions"].(map[string]interface{})
	extras := []field{
		{scripts, "start", "node index.js"},
		{compilerOptions, "module", "esnext"},
		{compilerOptions, "target", "esnext"},
		{compilerOptions, "moduleResolution", "node"},
		{compilerOptions, "allowJs", true},
		{compilerOptions, "resolveJsonModule", true},
	}
	for index := range extras {
		if _, ok := extras[index].Map[extras[index].Key]; !ok {
			extras[index].Map[extras[index].Key] = extras[index].Val
		}
	}
	crt, err := os.Create(filepath.Join(work, "package.json"))
	if err != nil {
		return stack.Stack(err, 0)
	}
	defer crt.Close()
	enc := json.NewEncoder(crt)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(&pkg)
	if err != nil {
		return stack.Stack(err, 0)
	}
	crt, err = os.Create(filepath.Join(work, "jsconfig.json"))
	if err != nil {
		return stack.Stack(err, 0)
	}
	defer crt.Close()
	enc = json.NewEncoder(crt)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(&jsconfig)
	if err != nil {
		return stack.Stack(err, 0)
	}

	names := []string{
		"index.js",
		"README.md",
		".gitignore",
	}
	contents := [][]string{
		{"console.log(\"hello nodejs!\")"},
		{filepath.Base(work), "to run,", "```bash", "node index.js", "```"},
		{"# https://github.com/github/gitignore/blob/main/Node.gitignore","","# Logs","logs","*.log","npm-debug.log*","yarn-debug.log*","yarn-error.log*","lerna-debug.log*",".pnpm-debug.log*","","# Diagnostic reports (https://nodejs.org/api/report.html)","report.[0-9]*.[0-9]*.[0-9]*.[0-9]*.json","","# Runtime data","pids","*.pid","*.seed","*.pid.lock","","# Directory for instrumented libs generated by jscoverage/JSCover","lib-cov","","# Coverage directory used by tools like istanbul","coverage","*.lcov","","# nyc test coverage",".nyc_output","","# Grunt intermediate storage (https://gruntjs.com/creating-plugins#storing-task-files)",".grunt","","# Bower dependency directory (https://bower.io/)","bower_components","","# node-waf configuration",".lock-wscript","","# Compiled binary addons (https://nodejs.org/api/addons.html)","build/Release","","# Dependency directories","node_modules/","jspm_packages/","","# Snowpack dependency directory (https://snowpack.dev/)","web_modules/","","# TypeScript cache","*.tsbuildinfo","","# Optional npm cache directory",".npm","","# Optional eslint cache",".eslintcache","","# Optional stylelint cache",".stylelintcache","","# Microbundle cache",".rpt2_cache/",".rts2_cache_cjs/",".rts2_cache_es/",".rts2_cache_umd/","","# Optional REPL history",".node_repl_history","","# Output of 'npm pack'","*.tgz","","# Yarn Integrity file",".yarn-integrity","","# dotenv environment variable files",".env",".env.development.local",".env.test.local",".env.production.local",".env.local","","# parcel-bundler cache (https://parceljs.org/)",".cache",".parcel-cache","","# Next.js build output",".next","out","","# Nuxt.js build / generate output",".nuxt","dist","","# Gatsby files",".cache/","# Comment in the public line in if your project uses Gatsby and not Next.js","# https://nextjs.org/blog/next-9-1#public-directory-support","# public","","# vuepress build output",".vuepress/dist","","# vuepress v2.x temp and cache directory",".temp",".cache","","# Docusaurus cache and generated files",".docusaurus","","# Serverless directories",".serverless/","","# FuseBox cache",".fusebox/","","# DynamoDB Local files",".dynamodb/","","# TernJS port file",".tern-port","","# Stores VSCode versions used for testing VSCode extensions",".vscode-test","","# yarn v2",".yarn/cache",".yarn/unplugged",".yarn/build-state.yml",".yarn/install-state.gz",".pnp.*"},
	}
	for index := 0; index < 3; index++ {
		_, err := os.Stat(filepath.Join(work, names[index]))
		if errors.Is(err, os.ErrNotExist) {
			crt, err = os.Create(filepath.Join(work, names[index]))
			if err != nil {
				return stack.Stack(err, 0)
			}
			defer crt.Close()
			crt.WriteString(strings.Join(contents[index], "\n"))
		}
	}

	return nil
}