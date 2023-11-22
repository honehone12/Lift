package gsprocess

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"lift/gsmap/gsparams"
	"lift/logger"
	"os/exec"
	"sync"
)

type GSProcess struct {
	cmd    *exec.Cmd
	params *gsparams.GSParams
	stdout io.ReadCloser
	stderr io.ReadCloser
	logger logger.Logger

	canceled        bool
	cancelFunc      context.CancelFunc
	processDoneFunc func()
	closingWait     sync.WaitGroup
	closeChLog      chan bool
	closeChErr      chan bool
}

var (
	ErrorMessageOnKill = "signal: killed"
)

func NewGSProcess(params *gsparams.GSParams, l logger.Logger) (*GSProcess, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, params.ProcessName(), params.ToArgs()...)
	p := &GSProcess{
		cmd:         cmd,
		params:      params,
		logger:      l,
		canceled:    true,
		cancelFunc:  cancel,
		closingWait: sync.WaitGroup{},
		closeChLog:  make(chan bool),
		closeChErr:  make(chan bool),
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

func (p *GSProcess) Start(processDoneFunc func()) error {
	if err := p.cmd.Start(); err != nil {
		return err
	}
	p.processDoneFunc = processDoneFunc
	p.canceled = false
	p.closingWait.Add(2)
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
	p.closeChLog <- true
	p.closeChErr <- true
	p.cancelFunc()
}

func (p *GSProcess) wait() {
	err := p.cmd.Wait()
	if err.Error() == ErrorMessageOnKill {
		p.logger.Info(p.params.LogFormat("killed"))
	} else if err != nil {
		p.logger.Error(p.params.LogFormat(err.Error()))
	}
	p.Cancel()
	p.closingWait.Wait()
	p.logger.Info(p.params.LogFormat("gs process successfully closed"))
	p.processDoneFunc()
}

func (p *GSProcess) recoverStdoutLog() {
	if r := recover(); r != nil {
		p.logger.Warn(p.params.LogFormat("recovering logging goroutine"))
		go p.stdoutLog()
	}
}

func (p *GSProcess) stdoutLog() {
	defer p.recoverStdoutLog()

	reader := bufio.NewReader(p.stdout)
	closing := false
LOOP:
	for {
		select {
		case <-p.closeChLog:
			break LOOP
		default:
			if closing {
				continue
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				p.logger.Error(p.params.LogFormat(fmt.Sprintf(
					"error: %s, waiting for closing logging goroutine", err.Error(),
				)))
				closing = true
				continue
			}

			p.logger.Info(p.params.LogFormat(line))
		}
	}

	p.logger.Debug(p.params.LogFormat("logging goroutine successfully closed"))
	p.closingWait.Done()
}

func (p *GSProcess) recoverStderrLog() {
	if r := recover(); r != nil {
		p.logger.Warn(p.params.LogFormat("recovering error logging goroutine"))
		go p.stderrLog()
	}
}

func (p *GSProcess) stderrLog() {
	defer p.recoverStderrLog()

	reader := bufio.NewReader(p.stderr)
	closing := false
LOOP:
	for {
		select {
		case <-p.closeChErr:
			break LOOP
		default:
			if closing {
				continue
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				p.logger.Error(p.params.LogFormat(fmt.Sprintf(
					"error: %s, waiting for closing error logging goroutine", err.Error(),
				)))
				closing = true
				continue
			}

			p.logger.Error(p.params.LogFormat(line))
		}
	}

	p.logger.Debug(p.params.LogFormat("error logging goroutine successfully closed"))
	p.closingWait.Done()
}
