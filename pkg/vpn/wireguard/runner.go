package wireguard

import "os/exec"

// ExecRunner abstracts running system commands so it can be mocked in tests.
type ExecRunner interface {
	Run(name string, args ...string) error
}

// DefaultExecRunner runs commands using os/exec.
type DefaultExecRunner struct{}

func (r *DefaultExecRunner) Run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}
