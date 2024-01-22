package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Engine string


const (
	EnginePDF Engine = "pdf"
	EngineLua Engine = "lua"
	EngineXeTeX Engine = "xe"
)

type BuildOptions struct {
	AuxDir string
	OutDir string
	SrcDir string
	SharedDir string
	Document string
	Engine Engine
	Force bool
	FileLineError bool
	Dependents bool
	BuildMode BuildMode
}

func RunBuild(ctx context.Context, options BuildOptions) (string, error) {
	if options.BuildMode == BuildModeNative {
		return RunBuildNative(ctx, options)
	} else if options.BuildMode == BuildModeDocker {
		// TODO re-add docker build mode
		return "", errors.New("docker build mode not yet implemented")
	}
	return "", errors.New("invalid build mode")
}

func RunBuildNative(ctx context.Context, options BuildOptions) (string, error) {
	var engineArg string
	switch options.Engine {
	case EnginePDF:
		engineArg = "-pdf"
	case EngineLua:
		engineArg = "-pdflua"
	case EngineXeTeX:
		engineArg = "-pdfxe"
	default:
		engineArg = "-pdf"
	}

	auxDir := fmt.Sprintf("-auxdir=%s", options.AuxDir)
	outDir := fmt.Sprintf("-outdir=%s", options.OutDir)

	args := []string{engineArg, auxDir, outDir, "-norc"};

	if options.Document != "" {
		args = append(args, options.Document)
	}

	if options.Force {
		args = append(args, "-f", "-interaction=nonstopmode")
	} else {
		args = append(args, "-interaction=batchmode")
	}

	if options.FileLineError {
		args = append(args, "-file-line-error")
	}

	if options.Dependents {
		args = append(args, "-deps")
	}

	err := os.Chdir(options.SrcDir)
	if err != nil {
		return "", fmt.Errorf("RunBuild: %w", err)
	}

	cmd := exec.CommandContext(ctx, "latexmk", args...)

	cmdOut := new(bytes.Buffer)
	cmd.Stdout = cmdOut
	cmd.Stderr = cmdOut

	log.Printf("Starting build in %s: %v", options.SrcDir, args)
	if err := cmd.Run(); err != nil {
		// If error is type *ExitError, the cmdOut should be populated
		// with an error message
		return cmdOut.String(), fmt.Errorf("RunBuild: %w", err)
	}

	return cmdOut.String(), nil
}
