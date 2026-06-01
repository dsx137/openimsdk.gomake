package mageutil

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/openimsdk/gomake/internal/util"
)

func CheckAndReportBinariesStatus() error {
	if err := InitForSSC(); err != nil {
		return err
	}
	err := CheckBinariesRunning()
	if err != nil {
		PrintErrRed("Some programs are not running properly:")
		PrintErrRedNoTimeStamp(err.Error())
		return err
	}
	PrintGreen("All services are running normally.")
	PrintBlue("Display details of the ports listened to by the service:")
	time.Sleep(1 * time.Second)
	err = PrintListenedPortsByBinaries()
	if err != nil {
		PrintErrRed("PrintListenedPortsByBinaries error")
		PrintErrRedNoTimeStamp(err.Error())
		return err
	}
	return nil
}

func StopAndCheckBinaries() error {
	if err := InitForSSC(); err != nil {
		return err
	}
	KillExistBinaries()
	err := attemptCheckBinaries()
	if err != nil {
		PrintErrRed(err.Error())
		return err
	}
	PrintGreen("All services have been stopped")
	return nil
}

func attemptCheckBinaries() error {
	const maxAttempts = 15
	var err error
	for i := 0; i < maxAttempts; i++ {
		err = CheckBinariesStop()
		if err == nil {
			return nil
		}
		PrintYellow("Some services have not been stopped, details are as follows: " + err.Error())
		PrintYellow("Continue to wait for 1 second before checking again")
		if i < maxAttempts-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return fmt.Errorf("already waited for %d seconds, some services have still not stopped", maxAttempts)
}

func StartToolsAndServices(binaries []string, pathOpts *PathOptions) error {
	if pathOpts != nil {
		if err := UpdateGlobalPaths(pathOpts); err != nil {
			PrintErrRed("Failed to update paths: " + err.Error())
			return err
		}
	}

	if len(binaries) > 0 {
		PrintBlue(fmt.Sprintf("Starting specified binaries: %v", binaries))

		var cmdBinaries, toolsBinaries []string

		for _, binary := range binaries {
			if isExecutableFile(GetBinFullPath(binary)) {
				if runtime.GOOS == "windows" {
					binary += ".exe"
				}
				cmdBinaries = append(cmdBinaries, binary)
			}
			if isExecutableFile(GetBinToolsFullPath(binary)) {
				if runtime.GOOS == "windows" {
					binary += ".exe"
				}
				toolsBinaries = append(toolsBinaries, binary)
			}
		}

		if len(cmdBinaries) == 0 && len(toolsBinaries) == 0 {
			PrintYellow("No valid executable binaries found to start. Please build first.")
			return nil
		}

		PrintBlue(fmt.Sprintf("Cmd binaries to start: %v", cmdBinaries))
		PrintBlue(fmt.Sprintf("Tools binaries to start: %v", toolsBinaries))

		if len(toolsBinaries) > 0 {
			PrintBlue("Starting specified tools...")
			if err := StartTools(toolsBinaries...); err != nil {
				PrintErrRed("Some specified tools failed to start:")
				PrintErrRedNoTimeStamp(err.Error())
				return err
			}
			PrintGreen("Specified tools executed successfully")
		}

		if len(cmdBinaries) > 0 {
			KillExistBinaries()
			err := attemptCheckBinaries()
			if err != nil {
				PrintErrRed("Some services running, details are as follows, abort start " + err.Error())
				return err
			}
			err = StartBinaries(cmdBinaries...)
			if err != nil {
				PrintErrRed("Failed to start specified binaries:")
				PrintErrRedNoTimeStamp(err.Error())
				return err
			}
			return CheckAndReportBinariesStatus()
		}
		return nil
	}

	PrintBlue("Starting tools primarily involves component verification and other preparatory tasks.")
	if err := StartTools(); err != nil {
		PrintErrRed("Some tools failed to start, details are as follows, abort start")
		PrintErrRedNoTimeStamp(err.Error())
		return err
	}
	PrintGreen("All tools executed successfully")

	KillExistBinaries()
	err := attemptCheckBinaries()
	if err != nil {
		PrintErrRed("Some services running, details are as follows, abort start " + err.Error())
		return err
	}
	err = StartBinaries()
	if err != nil {
		PrintErrRed("Failed to start all binaries")
		PrintErrRedNoTimeStamp(err.Error())
		return err
	}
	return CheckAndReportBinariesStatus()
}

func isExecutableFile(filePath string) bool {
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(filePath), ".exe") {
		filePath += ".exe"
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if !info.Mode().IsRegular() {
		return false
	}

	if runtime.GOOS == "windows" {
		return true
	}

	return info.Mode()&0111 != 0
}

func Build(binaries []string, pathOpts *PathOptions, buildOpt *BuildOptions) error {
	resolvedBuildOpt := ResolveBuildOptions(buildOpt, &BuildOptions{
		CgoEnabled: util.GetEnvWithNoErr[string]("CGO_ENABLED"),
		Release:    util.GetEnvWithNoErr[bool]("RELEASE"),
		Compress:   util.GetEnvWithNoErr[bool]("COMPRESS"),
		Platforms:  util.GetEnvWithNoErr[[]string]("PLATFORMS"),
	})

	if _, err := os.Stat(StartConfigFile); err == nil {
		if err := InitForSSC(); err != nil {
			return err
		}
	}

	if pathOpts != nil {
		if err := UpdateGlobalPaths(pathOpts); err != nil {
			PrintErrRed("Failed to update paths: " + err.Error())
			return err
		}
	}

	compileBinaries := getBinaries(binaries)
	if cgoEnabled := resolvedBuildOpt.GetCgoEnabled(); cgoEnabled != "" {
		PrintBlue(fmt.Sprintf("CGO_ENABLED %s", cgoEnabled))
	}
	platforms := resolvedBuildOpt.GetPlatforms()
	if len(platforms) == 0 {
		platform, err := DetectPlatform()
		if err != nil {
			return err
		}
		platforms = []string{platform}
	}
	for _, platform := range platforms {
		if err := CompileForPlatform(resolvedBuildOpt, platform, compileBinaries); err != nil {
			return err
		}
	}
	PrintGreen("All specified binaries under cmd and tools were successfully compiled.")
	return nil
}
