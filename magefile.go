//go:build mage

package main

import (
	"flag"

	"github.com/openimsdk/gomake/mageutil"
)

var Default = Build

var Aliases = map[string]any{
	"buildcc": BuildWithCustomConfig,
	"startcc": StartWithCustomConfig,
}

var (
	customRootDir = "."
	// customSrcDir  = "work_cmd"
	customSrcDir    = "cmd"
	customOutputDir = "_output"
	customConfigDir = "config"
	customToolsDir  = "tools"

	customExportProjectName = "gomake"
	customExportBuildOpt    *mageutil.BuildOptions
)

// Build support specifical binary build.
//
// Example: `mage build openim-api openim-rpc-user seq`
func Build() error {
	flag.Parse()
	bin := flag.Args()
	if len(bin) != 0 {
		bin = bin[1:]
	}

	return mageutil.WithSpinnerR("Building binaries...", func() error {
		return mageutil.Build(bin, nil, nil)
	})
}

func BuildWithCustomConfig() error {
	flag.Parse()
	bin := flag.Args()
	if len(bin) != 0 {
		bin = bin[1:]
	}

	config := &mageutil.PathOptions{
		RootDir:   &customRootDir,   // default is "."(current directory)
		OutputDir: &customOutputDir, // default is "_output"
		SrcDir:    &customSrcDir,    // default is "cmd"
		ToolsDir:  &customToolsDir,  // default is "tools"
	}

	return mageutil.WithSpinnerR("Building binaries with custom config...", func() error {
		return mageutil.Build(bin, config, nil)
	})
}

func Start() error {
	if err := mageutil.InitForSSC(); err != nil {
		return err
	}
	err := setMaxOpenFiles()
	if err != nil {
		mageutil.PrintRed("setMaxOpenFiles failed " + err.Error())
		return err
	}

	flag.Parse()
	bin := flag.Args()
	if len(bin) != 0 {
		bin = bin[1:]
	}

	return mageutil.WithSpinnerR("Starting tools and services...", func() error {
		return mageutil.StartToolsAndServices(bin, nil)
	})
}

func StartWithCustomConfig() error {
	if err := mageutil.InitForSSC(); err != nil {
		return err
	}
	err := setMaxOpenFiles()
	if err != nil {
		mageutil.PrintRed("setMaxOpenFiles failed " + err.Error())
		return err
	}

	flag.Parse()
	bin := flag.Args()
	if len(bin) != 0 {
		bin = bin[1:]
	}

	config := &mageutil.PathOptions{
		RootDir:   &customRootDir,   // default is "."(current directory)
		OutputDir: &customOutputDir, // default is "_output"
		ConfigDir: &customConfigDir, // default is "config"
	}

	return mageutil.WithSpinnerR("Starting tools and services with custom config...", func() error {
		return mageutil.StartToolsAndServices(bin, config)
	})
}

func Stop() error {
	return mageutil.WithSpinnerR("Checking service status...", mageutil.StopAndCheckBinaries)
}

func Check() error {
	return mageutil.WithSpinnerR("Checking service status...", mageutil.CheckAndReportBinariesStatus)
}

func Protocol() error {
	return mageutil.WithSpinnerR("Generating protocol artifacts...", mageutil.Protocol)
}

func Export() error {
	exportOpt := &mageutil.ExportOptions{
		ProjectName: &customExportProjectName,
		BuildOpt:    customExportBuildOpt,
	}
	err := mageutil.WithSpinnerR("Exporting launcher archive...", func() error {
		return mageutil.ExportMageLauncherArchived(nil, exportOpt)
	})
	if err != nil {
		mageutil.PrintRed("export failed " + err.Error())
		return err
	}
	return nil
}
