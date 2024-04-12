package exec

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func TestCmd_NewCmd(t *testing.T) {
	want := &Cmd{cmd: exec.Command("echo", "foo")}
	got := NewCmd("echo", "foo")

	assert.Equal(t, want, got)
}

func TestCmd_Run_Success(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("echo", "foo")}
	err := cmd.Run()
	assert.NoError(t, err)
}

func TestCmd_Run_Error(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("foo")}
	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executable file not found in $PATH")
}

func TestCmd_SetDir_Success(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("pwd")}
	cmd.SetDir("/tmp")
	stdout := bytes.NewBuffer([]byte(""))
	cmd.cmd.Stdout = stdout
	err := cmd.cmd.Run()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "/tmp")
}

func TestCmd_SetStdout(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("pwd")}
	cmd.SetDir("/tmp")
	stdout := bytes.NewBuffer([]byte(""))
	cmd.SetStdout(stdout)
	err := cmd.cmd.Run()
	assert.NoError(t, err)
	assert.Contains(t, stdout.String(), "/tmp")
}

func TestCmd_SetStderr(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("bash", "-c", "echo foo >&2")}
	stderr := bytes.NewBuffer([]byte(""))
	cmd.SetStderr(stderr)
	err := cmd.cmd.Run()
	assert.NoError(t, err)
	assert.Contains(t, stderr.String(), "foo")
}

func TestCmd_SetEnv(t *testing.T) {
	cmd := &Cmd{cmd: exec.Command("echo")}
	cmd.SetEnv([]string{"FOO=bar"})

	assert.Equal(t, []string{"FOO=bar"}, cmd.cmd.Env)
}
