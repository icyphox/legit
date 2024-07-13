package service

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
)

// Mostly from charmbracelet/soft-serve and sosedoff/gitkit.

type ServiceCommand struct {
	Dir    string
	Stdin  io.Reader
	Stdout http.ResponseWriter
}

func (c *ServiceCommand) InfoRefs() error {
	cmd := exec.Command("git", []string{
		"upload-pack",
		"--stateless-rpc",
		"--advertise-refs",
		".",
	}...)

	cmd.Dir = c.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Printf("git: failed to start git-upload-pack (info/refs): %s", err)
		return err
	}

	if err := packLine(c.Stdout, "# service=git-upload-pack\n"); err != nil {
		log.Printf("git: failed to write pack line: %s", err)
		return err
	}

	if err := packFlush(c.Stdout); err != nil {
		log.Printf("git: failed to flush pack: %s", err)
		return err
	}

	buf := bytes.Buffer{}
	if _, err := io.Copy(&buf, stdoutPipe); err != nil {
		log.Printf("git: failed to copy stdout to tmp buffer: %s", err)
		return err
	}

	if err := cmd.Wait(); err != nil {
		out := strings.Builder{}
		_, _ = io.Copy(&out, &buf)
		log.Printf("git: failed to run git-upload-pack; err: %s; output: %s", err, out.String())
		return err
	}

	if _, err := io.Copy(c.Stdout, &buf); err != nil {
		log.Printf("git: failed to copy stdout: %s", err)
	}

	return nil
}

func (c *ServiceCommand) UploadPack() error {
	cmd := exec.Command("git", []string{
		"-c", "uploadpack.allowFilter=true",
		"upload-pack",
		"--stateless-rpc",
		".",
	}...)
	cmd.Dir = c.Dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	defer stdoutPipe.Close()

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdinPipe.Close()

	if err := cmd.Start(); err != nil {
		log.Printf("git: failed to start git-upload-pack: %s", err)
		return err
	}

	if _, err := io.Copy(stdinPipe, c.Stdin); err != nil {
		log.Printf("git: failed to copy stdin: %s", err)
		return err
	}
	stdinPipe.Close()

	if _, err := io.Copy(newWriteFlusher(c.Stdout), stdoutPipe); err != nil {
		log.Printf("git: failed to copy stdout: %s", err)
		return err
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("git: failed to wait for git-upload-pack: %s", err)
		return err
	}

	return nil
}

func packLine(w io.Writer, s string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(s)+4, s)
	return err
}

func packFlush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}
