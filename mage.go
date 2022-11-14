//go:build mage

package main

import (
	"context"

	"github.com/aaronzjc/onlinestat/scripts"
)

// Build run go build
func Build() error {
	return scripts.Build(context.Background())
}

// Image run docker build & push
func Image() error {
	return scripts.Image(context.Background())
}

// Deploy run deployment
func Deploy() error {
	return scripts.Deploy(context.Background())
}
