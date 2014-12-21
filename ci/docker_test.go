package grader

import (
	"bufio"
	"bytes"
	//"io"
	"strings"
	"sync"
	"testing"
	"time"

	docker "github.com/fsouza/go-dockerclient"
)

var lock sync.Mutex

func TestNewVirtual(t *testing.T) {
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log(err)
	}

	if v.Client == nil {
		t.Fail()
		t.Log("Missing the docker client object")
	}

	recoveryendpoint := endpoint
	endpoint = ""
	v, err = NewVirtual()
	if err == nil {
		t.Fail()
		t.Log(err)
	}

	if v.Client != nil {
		t.Fail()
		t.Log("Managed to create the docker client object, should be nil.")
	}
	endpoint = recoveryendpoint
}

func TestNewContainer(t *testing.T) {
	lock.Lock()
	defer lock.Unlock()
	endpoint = "unix:///var/run/docker.sock"
	var err error
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log(err)
		return
	}

	err = v.NewContainer("autograder")
	if err != nil {
		t.Fail()
		t.Log("Error while creating new container: ", err)
		return
	}

	// cleanup
	v.RemoveContainer()
}

func TestKillContainer(t *testing.T) {
	lock.Lock()
	defer lock.Unlock()

	var err error
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}
	err = v.NewContainer("autograder")
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}

	err = v.KillContainer()
	if err != nil {
		t.Fail()
		t.Log("Error while killing container: ", err)
		return
	}

	// cleanup
	v.RemoveContainer()
}

func TestRemoveContainer(t *testing.T) {
	lock.Lock()
	defer lock.Unlock()
	var err error
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}
	err = v.NewContainer("autograder")
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}

	// cleanup
	defer v.RemoveContainer()

	ID := v.Container.ID

	t.Log("Removing container: ", ID)

	err = v.RemoveContainer()
	if err != nil {
		t.Fail()
		t.Log("Error while removing container: ", err)
	}

	listopt := docker.ListContainersOptions{
		All: true,
	}

	list, err := v.Client.ListContainers(listopt)
	if err != nil {
		t.Fail()
		t.Log("Error loading containers: ", err)
	}

	for _, x := range list {
		if ID == x.ID {
			t.Fail()
			t.Log("Didn't remove container: ", ID)
		}
	}
}

func TestAttachToContainer(t *testing.T) {
	lock.Lock()
	defer lock.Unlock()
	var err error
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}
	err = v.NewContainer("autograder")
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}

	// cleanup
	defer v.RemoveContainer()

	var reader = strings.NewReader("echo hello world; exit;")
	buf := bytes.NewBufferString("test")
	stdout := bufio.NewWriter(buf)

	go v.AttachToContainer(reader, stdout, stdout)

	//	msg := []byte("echo hello world")
	//	n, err := w.Write(msg)
	//	if err != nil {
	//		t.Fail()
	//		t.Log("Couldnt write to the stdin: ", err)
	//		return
	//	}
	//	if n != len(msg) {
	//		t.Fail()
	//		t.Log("Didn't write the whole msg to stdin")
	//		return
	//	}

	// make sure command is executed
	time.Sleep(10000 * time.Microsecond)

	text := buf.String()

	t.Log("Read from container: " + text)

	if text != "hello world" {
		t.Fail()
	}
}

var commandtests = []struct {
	in  string
	out string
}{
	{"echo hello world", "hello world"},
}

func TestExecuteCommands(t *testing.T) {
	lock.Lock()
	defer lock.Unlock()
	var err error
	v, err := NewVirtual()
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}
	err = v.NewContainer("autograder")
	if err != nil {
		t.Fail()
		t.Log("Couldn't set up test env.")
	}

	// cleanup
	defer v.RemoveContainer()

	cmd := commandtests[0].in

	buf := bytes.NewBuffer(make([]byte, 16))

	v.ExecuteCommand(cmd, nil, bufio.NewWriter(buf), bufio.NewWriter(buf))
	if err != nil {
		t.Fail()
		t.Log("Couldn't execute commands on container: ", err)
	}

	scan := bufio.NewScanner(buf)

	result := scan.Scan()
	if !result {
		t.Fail()
		t.Log("Recieved nothing")
	}

	text := scan.Text()
	t.Log("Read from container: \"" + text + "\"")

	if text != commandtests[0].out {
		t.Fail()
		t.Log("Error: " + text + " != " + commandtests[0].out)
	}
}
