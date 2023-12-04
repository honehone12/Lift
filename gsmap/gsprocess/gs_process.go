package gsprocess

import (
	"bufio"
	"context"
	"io"
	"lift/gsmap/gsinfo"
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

	status          uint8
	canceled        bool
	cancel          context.CancelFunc
	onProcessClosed func()
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
		status:      gsinfo.ProcessStatusNotStarted,
		canceled:    true,
		cancel:      cancel,
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

func (p *GSProcess) Start(onProcessClosed func()) error {
	if err := p.cmd.Start(); err != nil {
		return err
	}
	p.onProcessClosed = onProcessClosed
	p.status = gsinfo.ProcessStatusOK
	p.canceled = false
	p.closingWait.Add(2)
	go p.stdoutLog(true)
	go p.stdoutLog(false)
	go p.wait()
	return nil
}

func (p *GSProcess) Status() uint8 {
	return p.status
}

func (p *GSProcess) Close() {
	if p.canceled {
		return
	}

	if p.status != gsinfo.ProcessStatusError {
		p.status = gsinfo.ProcessStatusCanceled
	}
	p.canceled = true
	p.cancel()
	p.closeChLog <- true
	p.closeChErr <- true
}

func (p *GSProcess) wait() {
	err := p.cmd.Wait()
	if err.Error() == ErrorMessageOnKill {
		p.logger.Info(p.params.LogWithId("killed"))
	} else if err != nil {
		p.status = gsinfo.ProcessStatusError
		p.logger.Error(p.params.LogWithId(err.Error()))
	}
	p.Close()
	p.closingWait.Wait()
	p.logger.Debug(p.params.LogWithId("gs process successfully closed"))
	p.onProcessClosed()
}

func (p *GSProcess) recoverStdoutLog(errLog bool) {
	if r := recover(); r != nil {
		if errLog {
			p.logger.Warn(p.params.LogWithId("recovering error logging goroutine"))
		} else {
			p.logger.Warn(p.params.LogWithId("recovering logging goroutine"))

		}

		go p.stdoutLog(errLog)
	}
}

func (p *GSProcess) stdoutLog(errLog bool) {
	defer p.recoverStdoutLog(errLog)

	var reader *bufio.Reader
	if errLog {
		reader = bufio.NewReader(p.stderr)
	} else {
		reader = bufio.NewReader(p.stdout)
	}

	var closeCh chan bool
	if errLog {
		closeCh = p.closeChErr
	} else {
		closeCh = p.closeChLog
	}

	closing := false
LOOP:
	for {
		select {
		case <-closeCh:
			break LOOP
		default:
			if closing {
				continue
			}

			line, err := reader.ReadString('\n')
			if errLog {
				if err != nil {
					p.logger.Errorf(p.params.LogWithId(
						"error: %s, waiting for closing error logging goroutine"),
						err.Error(),
					)
					closing = true
					continue
				}

				p.logger.Error(p.params.LogWithId(line))
				p.status = gsinfo.ProcessStatusError
			} else {
				if err != nil {
					p.logger.Errorf(p.params.LogWithId(
						"error: %s, waiting for closing logging goroutine"),
						err.Error(),
					)
					closing = true
					p.status = gsinfo.ProcessStatusError
					continue
				}

				p.logger.Info(p.params.LogWithId(line))
			}
		}
	}

	if errLog {
		p.logger.Debug(p.params.LogWithId("error logging goroutine successfully closed"))
	} else {
		p.logger.Debug(p.params.LogWithId("logging goroutine successfully closed"))
	}
	p.closingWait.Done()
}
