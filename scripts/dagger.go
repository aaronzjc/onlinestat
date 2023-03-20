package scripts

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"
)

const (
	Version = "1.2"
)

func Build(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	envs := map[string]string{
		"GO111MODULE": "on",
		"GOPROXY":     "https://goproxy.cn,direct",
		"CGO_ENABLED": "0",
		"GOOS":        "linux",
		"GOARCH":      "amd64",
	}

	// 获取本地项目路径
	src := client.Host().Workdir()
	golang := client.Container().From("golang:1.19-alpine3.15")
	golang = golang.WithMountedDirectory("/src", src).WithWorkdir("/src")
	for k, v := range envs {
		golang = golang.WithEnvVariable(k, v)
	}
	path := "bin/"
	golang = golang.Exec(dagger.ContainerExecOpts{
		Args: []string{"go", "build", "-o", path + "online-stat", "main.go"},
	})

	if _, err := golang.Directory(path).Export(ctx, path); err != nil {
		return err
	}

	return nil
}

func Image(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	src := client.Host().Workdir()
	docker := client.Container()
	docker = docker.Build(src, dagger.ContainerBuildOpts{Dockerfile: "./scripts/dockerfile"})
	resp, err := docker.Publish(ctx, "aaronzjc/online-stat:"+Version)
	if err != nil {
		return err
	}
	fmt.Println(resp)

	return nil
}

func Deploy(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return err
	}
	defer client.Close()

	// 处理版本
	oldTag, newTag := "latest", Version
	file := "./scripts/k8s.yaml"
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	out := strings.ReplaceAll(string(data), oldTag, newTag)
	os.WriteFile(file, []byte(out), 0666)
	defer os.WriteFile(file, data, 0666)

	kubectl := client.Container().From("bitnami/kubectl")
	kubeconfig := client.Host().Workdir().File("./scripts/kubeconf.yaml")
	kubectl = kubectl.WithMountedFile("/.kube/config", kubeconfig)
	deployment := client.Host().Workdir().File(file)
	kubectl = kubectl.WithMountedFile("/tmp/deployment.yaml", deployment)

	kubectl = kubectl.Exec(dagger.ContainerExecOpts{
		Args: []string{"apply", "-f", "/tmp/deployment.yaml", "-n", "k3s-apps"},
	})
	logs, err := kubectl.Stdout().Contents(ctx)
	if err != nil {
		return err
	}
	fmt.Println(logs)
	return nil
}
