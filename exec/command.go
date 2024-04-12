package exec

import (
	"io"
	"os/exec"
)

var (
	_ Executable = &Cmd{}
)

var NewCmd = func(name string, arg ...string) Executable {
	return &Cmd{cmd: exec.Command(name, arg...)}
}

type Executable interface {
	SetDir(dir string)
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
	SetEnv(env []string)
	Run() error
}

type Cmd struct {
	cmd *exec.Cmd
}

func (c Cmd) Run() error {
	return c.cmd.Run()
}

func (c Cmd) SetDir(dir string) {
	c.cmd.Dir = dir
}

func (c Cmd) SetStdout(stdout io.Writer) {
	c.cmd.Stdout = stdout
}

func (c Cmd) SetStderr(stderr io.Writer) {
	c.cmd.Stderr = stderr
}

func (c Cmd) SetEnv(env []string) {
	c.cmd.Env = env
}
