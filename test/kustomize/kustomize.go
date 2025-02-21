package kustomize

import (
	"bytes"
	"fmt"
	"os/exec"
)

var logOutput = false

func LogOutput(enabled bool) {
	logOutput = enabled
}

func BuildDir(dir string) (*bytes.Buffer, error) {
	cmd := exec.Command("kustomize", "build")
	cmd.Dir = dir

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()
	buffer := bytes.NewBuffer(output)

	if logOutput {
		fmt.Println(string(output))
	}

	return buffer, err
}

func BuildUrl(url string) (*bytes.Buffer, error) {
	cmd := exec.Command("kustomize", "build", url)

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()
	buffer := bytes.NewBuffer(output)

	if logOutput {
		fmt.Println(string(output))
	}

	return buffer, err
}

func SetNamespace(dir, namespace string) error {
	cmd := exec.Command("kustomize", "edit", "set", "namespace", namespace)
	cmd.Dir = dir

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()

	if logOutput {
		fmt.Println(string(output))
	}

	return err
}

func AddResource(path string) error {
	cmd := exec.Command("kustomize", "edit", "add", "resource", path)
	cmd.Dir = "../testdata/k8ssandra-operator"

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()

	if logOutput {
		fmt.Println(string(output))
	}

	return err
}

func RemoveResource(path string) error {
	cmd := exec.Command("kustomize", "edit", "remove", "resource", path)
	cmd.Dir = "../testdata/k8ssandra-operator"

	fmt.Println(cmd)

	output, err := cmd.CombinedOutput()

	if logOutput {
		fmt.Println(string(output))
	}

	return err
}
