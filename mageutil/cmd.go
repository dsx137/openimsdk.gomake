package mageutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/openimsdk/gomake/internal/priority"
	"github.com/openimsdk/gomake/internal/util"
)

type Cmd struct {
	execCmd *exec.Cmd

	env map[string]string

	priority *priority.Level

	stdoutBuf bytes.Buffer
	stderrBuf bytes.Buffer
}

func NewCmd(command string) *Cmd {
	return &Cmd{execCmd: exec.Command(command)}
}

func (c *Cmd) WithArgs(args ...string) *Cmd {
	c.execCmd.Args = append(c.execCmd.Args, args...)
	return c
}

func (c *Cmd) WithEnv(env map[string]string) *Cmd {
	if len(env) == 0 {
		return c
	}
	if c.env == nil {
		c.env = make(map[string]string, len(env))
	}
	for k, v := range env {
		c.env[k] = v
	}
	return c
}

func (c *Cmd) WithDir(dir string) *Cmd {
	c.execCmd.Dir = strings.TrimSpace(dir)
	return c
}

func (c *Cmd) WithPriority(priority priority.Level) *Cmd {
	c.priority = &priority
	return c
}

func (c *Cmd) WithStdin(stdin io.Reader) *Cmd {
	c.execCmd.Stdin = stdin
	return c
}

func (c *Cmd) WithStdout(stdout io.Writer) *Cmd {
	if stdout == nil {
		stdout = commandOutputWriter(os.Stdout, &c.stdoutBuf)
	} else {
		c.execCmd.Stdout = stdout
	}
	return c
}

func (c *Cmd) WithStderr(stderr io.Writer) *Cmd {
	if stderr == nil {
		stderr = commandOutputWriter(os.Stderr, &c.stderrBuf)
	} else {
		c.execCmd.Stderr = stderr
	}
	return c
}

func (c *Cmd) Start() error {
	c.execCmd.Env = append(os.Environ(), util.FlattenEnvs(c.env)...)

	if err := c.execCmd.Start(); err != nil {
		return err
	}

	if c.priority != nil && c.execCmd.Process != nil {
		if err := priority.Set(c.execCmd.Process.Pid, *c.priority); err != nil {
			PrintYellow(fmt.Sprintf("Failed to set priority for PID %d: %v", c.execCmd.Process.Pid, err))
		}
	}

	return nil
}

func (c *Cmd) Wait() error {
	return c.execCmd.Wait()
}

func (c *Cmd) Run() error {
	err := c.Start()
	if err != nil {
		return err
	}

	return c.Wait()
}

func (c *Cmd) String() string {
	return c.execCmd.String()
}

func (c *Cmd) Output() ([]byte, error) {
	err := c.Run()
	dst := make([]byte, len(c.stdoutBuf.Bytes()))
	copy(dst, c.stdoutBuf.Bytes())
	return dst, err
}

func commandOutputWriter(writers ...io.Writer) io.Writer {
	logFile, err := getSharedLogFile()
	if err != nil {
		PrintYellow(fmt.Sprintf("Warning: failed to open log file for command output: %v", err))
		return io.MultiWriter(writers...)
	}
	return io.MultiWriter(append(writers, logFile)...)
}
