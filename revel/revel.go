// Copyright (c) 2012-2016 The Revel Framework Authors, All rights reserved.
// Revel Framework source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

// The command line tool for running Revel apps.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/agtorre/gocolorize"
	"github.com/revel/cmd/logger"
	"github.com/revel/cmd/model"
	"github.com/revel/cmd/utils"
)

const (
	// RevelCmdImportPath Revel framework cmd tool import path
	RevelCmdImportPath = "github.com/revel/cmd"

	// RevelCmdImportPath Revel framework cmd tool import path
	RevelSkeletonsImportPath = "github.com/revel/skeletons"

	// DefaultRunMode for revel's application
	DefaultRunMode = "dev"
)

// Command structure cribbed from the genius organization of the "go" command.
type Command struct {
	UpdateConfig           func(c *model.CommandConfig, args []string) bool
	RunWith                func(c *model.CommandConfig) error
	UsageLine, Short, Long string
}

// Name returns command name from usage line
func (cmd *Command) Name() string {
	name := cmd.UsageLine
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// The commands
var Commands = []*Command{
	nil, // Safety net, prevent missing index from running
	cmdNew,
	cmdRun,
	cmdBuild,
	cmdPackage,
	cmdClean,
	cmdTest,
	cmdVersion,
}

func main() {
	if runtime.GOOS == "windows" {
		gocolorize.SetPlain(true)
	}
	c := &model.CommandConfig{}
	wd, _ := os.Getwd()

	utils.InitLogger(wd, logger.LvlError)

	parser := flags.NewParser(c, flags.HelpFlag|flags.PassDoubleDash)
	if err := ParseArgs(c, parser, os.Args[1:]); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		parser.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	// Switch based on the verbose flag
	if len(c.Verbose)>1 {
		utils.InitLogger(wd, logger.LvlDebug)
	} else if len(c.Verbose)>0 {
		utils.InitLogger(wd, logger.LvlInfo)
	} else {
		utils.InitLogger(wd, logger.LvlWarn)
	}

	if !c.UpdateImportPath() {
		utils.Logger.Fatal("Unable to determine application path")
	}

	command := Commands[c.Index]
	println("Revel executing:", command.Short)

	// Setting go paths
	c.InitGoPaths()

	// Setup package resolver
	c.InitPackageResolver()

	if err := command.RunWith(c); err != nil {
		utils.Logger.Error("Unable to execute","error",err)
		os.Exit(1)
	}
}

// Parse the arguments passed into the model.CommandConfig
func ParseArgs(c *model.CommandConfig, parser *flags.Parser, args []string) (err error) {
	var extraArgs []string
	if ini := flag.String("ini", "none", ""); *ini != "none" {
		if err = flags.NewIniParser(parser).ParseFile(*ini); err != nil {
			return
		}
	} else {
		if extraArgs, err = parser.ParseArgs(args); err != nil {
			return
		} else {
			switch parser.Active.Name {
			case "new":
				c.Index = model.NEW
			case "run":
				c.Index = model.RUN
			case "build":
				c.Index = model.BUILD
			case "package":
				c.Index = model.PACKAGE
			case "clean":
				c.Index = model.CLEAN
			case "test":
				c.Index = model.TEST
			case "version":
				c.Index = model.VERSION
			}
		}
	}

	if c.Index == 0 {
		err = fmt.Errorf("Unknown command %v", extraArgs)
	} else if len(extraArgs) > 0 {
		utils.Logger.Info("Found additional arguements, setting them")
		if !Commands[c.Index].UpdateConfig(c, extraArgs) {
			err = fmt.Errorf("Invalid command line arguements %v", extraArgs)
		}
	}

	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
