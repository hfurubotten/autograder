package ci

import (
	"errors"
	"io"
	"os/exec"
	"sync"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	endpoint string = "unix:///var/run/docker.sock"
	cmdlock  sync.Mutex
)

type Virtual struct {
	Client    *docker.Client
	Imagename string
	Container *docker.Container
}

func NewVirtual() (v Virtual, err error) {
	c, err := docker.NewClient(endpoint)
	if err != nil {
		return
	}

	v = Virtual{
		Client: c,
	}

	return
}

func (v *Virtual) NewContainer(imagename string) (err error) {
	if v.Container != nil {
		v.RemoveContainer()
	}

	conf := docker.Config{
		Image:        imagename,
		AttachStdout: true,
		AttachStdin:  true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/bash"},
		OpenStdin:    true,
		StdinOnce:    true,
	}
	contopt := docker.CreateContainerOptions{
		Config: &conf,
	}

	v.Container, err = v.Client.CreateContainer(contopt)
	if err != nil {
		return
	}

	return
}

func (v *Virtual) KillContainer() (err error) {
	if v.Container == nil {
		return
	}

	killopt := docker.KillContainerOptions{
		ID: v.Container.ID,
	}
	err = v.Client.KillContainer(killopt)
	if err != nil {
		return
	}

	return
}

func (v *Virtual) RemoveContainer() (err error) {
	if v.Container == nil {
		return
	}

	rmopt := docker.RemoveContainerOptions{
		ID:            v.Container.ID,
		RemoveVolumes: true,
		Force:         true,
	}

	err = v.Client.RemoveContainer(rmopt)
	if err != nil {
		return
	}

	v.Container = nil
	return
}

func (v *Virtual) AttachToContainer(stdin io.Reader, stdout io.Writer, stderr io.Writer) (err error) {
	if v.Container == nil {
		return errors.New("Does not have any container started up yet.")
	}

	attopt := docker.AttachToContainerOptions{
		Container:    v.Container.ID,
		InputStream:  stdin,
		OutputStream: stdout,
		ErrorStream:  stderr,
		Stdin:        true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
		RawTerminal:  true,
	}

	err = v.Client.StartContainer(v.Container.ID, &docker.HostConfig{})
	if err != nil {
		return
	}

	err = v.Client.ResizeContainerTTY(v.Container.ID, 20, 20)
	if err != nil {
		return
	}

	err = v.Client.AttachToContainer(attopt)
	if err != nil {
		return
	}

	return
}

func (v *Virtual) ExecuteCommand(commands string, stdin io.Reader, stdout, stderr io.Writer) (err error) {
	cmdlock.Lock()
	defer cmdlock.Unlock()
	if v.Container == nil {
		return errors.New("Does not have any container started up yet.")
	}

	err = v.Client.StartContainer(v.Container.ID, nil)
	if _, ok := err.(*docker.ContainerAlreadyRunning); err != nil && !ok {
		return
	}

	cmd := exec.Command("/bin/bash", "-c", "docker exec "+v.Container.ID+" "+commands+"")

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Run()
	if err != nil {
		return
	}

	/*
		// blocks the system
		c := make(chan struct{})
		cexecopt := docker.CreateExecOptions{
			Cmd:          commands,
			Tty:          true,
			Container:    v.Container.ID,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
		}

		exec, err := v.Client.CreateExec(cexecopt)
		if err != nil {
			return
		}

		sexecopt := docker.StartExecOptions{
			Detach:       false,
			RawTerminal:  true,
			Tty:          true,
			InputStream:  stdin,
			OutputStream: stdout,
			ErrorStream:  stderr,
			Success:      c,
		}

		err = v.Client.StartExec(exec.ID, sexecopt)
		if err != nil {
			return
		}*/

	return
}
