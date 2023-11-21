package gsprocess

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"lift/gs/gsparams"
	"lift/logger"
	"os/exec"
)

type GSProcess struct {
	cmd    *exec.Cmd
	params *gsparams.GSParams
	stdout io.ReadCloser
	stderr io.ReadCloser
	logger logger.Logger

	canceled   bool
	cancelFunc context.CancelFunc
	closeChLog chan bool
	closeChErr chan bool
}

var (
	ErrorMessageOnKill = "signal: killed"
)

func NewGSProcess(params *gsparams.GSParams, l logger.Logger) (*GSProcess, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, params.ProcessName(), params.ToArgs()...)
	p := &GSProcess{
		cmd:        cmd,
		params:     params,
		logger:     l,
		canceled:   true,
		cancelFunc: cancel,
		closeChLog: make(chan bool),
		closeChErr: make(chan bool),
	}
	outPipe, err := p.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errPipe, err := p.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	p.stdout = outPipe
	p.stderr = errPipe
	return p, nil
}

func (p *GSProcess) logFormat(msg string) string {
	return fmt.Sprintf("GS PROCESS [%s] ", p.params.Uuid.String()) + msg
}

func (p *GSProcess) Start() error {
	if err := p.cmd.Start(); err != nil {
		return err
	}
	p.canceled = false
	go p.stdoutLog()
	go p.stderrLog()
	go p.wait()
	return nil
}

func (p *GSProcess) Cancel() {
	if p.canceled {
		return
	}

	p.canceled = true
	p.cancelFunc()
	p.closeChLog <- true
	p.closeChErr <- true
}

func (p *GSProcess) wait() {
	err := p.cmd.Wait()
	if err.Error() == ErrorMessageOnKill {
		p.logger.Info(p.logFormat("killed"))
	} else if err != nil {
		p.logger.Error(p.logFormat(err.Error()))
	}
	p.Cancel()
}

func (p *GSProcess) recoverStdoutLog() {
	if r := recover(); r != nil {
		p.logger.Warn(p.logFormat("recovering logging goroutine"))
		go p.stdoutLog()
	}
}

func (p *GSProcess) stdoutLog() {
	defer p.recoverStdoutLog()

	reader := bufio.NewReader(p.stdout)
	eof := false
LOOP:
	for {
		select {
		case <-p.closeChLog:
			break LOOP
		default:
			if !eof {
				line, err := reader.ReadString('\n')
				if errors.Is(err, io.EOF) {
					p.logger.Info(p.logFormat(
						"EOF, waiting for closing logging goroutine",
					))
					eof = true
					continue
				} else if err != nil {
					panic(p.logFormat(err.Error()))
				}
				p.logger.Info(p.logFormat(line))
			}
		}
	}

	p.logger.Info(p.logFormat("logging goroutine successfully closed"))
}

func (p *GSProcess) recoverStderrLog() {
	if r := recover(); r != nil {
		p.logger.Warn(p.logFormat("recovering error logging goroutine"))
		go p.stderrLog()
	}
}

func (p *GSProcess) stderrLog() {
	defer p.recoverStderrLog()

	reader := bufio.NewReader(p.stderr)
	eof := false
LOOP:
	for {
		select {
		case <-p.closeChErr:
			break LOOP
		default:
			if !eof {
				line, err := reader.ReadString('\n')
				if errors.Is(err, io.EOF) {
					p.logger.Info(p.logFormat(
						"EOF, waiting for closing error logging goroutine",
					))
					eof = true
					continue
				} else if err != nil {
					panic(p.logFormat(err.Error()))
				}
				p.logger.Error(p.logFormat(line))
			}
		}
	}

	p.logger.Info(p.logFormat("error logging goroutine successfully closed"))
}
