package ci

import (
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var (
	endpoint = "unix:///var/run/docker.sock"
	cmdlock  sync.Mutex
)

// Virtual encapsulates a docker environment where commands can be executed.
type Virtual struct {
	Client    *docker.Client
	Imagename string
	Container *docker.Container
}

// NewVirtual returns a new Virtual object.
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

// NewContainer creates a new container from a image name.
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
	return err
}

// KillContainer will kill a running container.
func (v *Virtual) KillContainer() (err error) {
	if v.Container == nil {
		return
	}

	killopt := docker.KillContainerOptions{
		ID: v.Container.ID,
	}
	return v.Client.KillContainer(killopt)
}

// RemoveContainer will force a removal of a container in the docker system.
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

//TODO this function is never used in autograder.
// AttachToContainer will attach given readers and writers to streams from a running docker container.
func (v *Virtual) xAttachToContainer(stdin io.Reader, stdout, stderr io.Writer) (err error) {
	if v.Container == nil {
		return errors.New("container is not running yet")
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

// Execute the given command in a running docker container.
func (v *Virtual) Execute(command string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmdlock.Lock()
	defer cmdlock.Unlock()
	if v.Container == nil {
		return errors.New("container is not running yet")
	}
	err := v.Client.StartContainer(v.Container.ID, &docker.HostConfig{})
	if _, ok := err.(*docker.ContainerAlreadyRunning); err != nil && !ok {
		return err
	}

	cmd := exec.Command("/bin/bash", "-c", "docker exec "+v.Container.ID+" "+command+"")

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(5 * time.Minute):
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
		<-done // allow goroutine to exit
		return errors.New("Process killed on timeout after 5 min.")
	case err = <-done:
		return err
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
}
