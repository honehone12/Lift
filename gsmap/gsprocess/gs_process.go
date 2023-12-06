package gsprocess

import (
	"bufio"
	"context"
	"io"
	"lift/gsmap/gsparams"
	"lift/logger"
	"os/exec"
	"sync"
	"sync/atomic"
)

type GSProcess struct {
	cmd    *exec.Cmd
	params *gsparams.GSParams
	stdout io.ReadCloser
	stderr io.ReadCloser
	logger logger.Logger

	cancelProcess   context.CancelFunc
	canceled        *atomic.Bool
	onProcessClosed func()

	closingWait sync.WaitGroup
	closeChLog  chan bool
	closeChErr  chan bool
}

var (
	ErrorMessageOnKill = "signal: killed"
)

func NewGSProcess(params *gsparams.GSParams, l logger.Logger) (*GSProcess, error) {
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, params.ProcessName(), params.ToArgs()...)
	b := &atomic.Bool{}
	b.Store(true)
	p := &GSProcess{
		cmd:             cmd,
		params:          params,
		logger:          l,
		cancelProcess:   cancel,
		canceled:        b,
		onProcessClosed: nil,
		closingWait:     sync.WaitGroup{},
		closeChLog:      make(chan bool),
		closeChErr:      make(chan bool),
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
	p.canceled.Store(false)
	p.closingWait.Add(2)
	go p.stdoutLog(true)
	go p.stdoutLog(false)
	go p.wait()
	return nil
}

func (p *GSProcess) Close() {
	if p.canceled.Load() {
		return
	}

	p.canceled.Store(true)
	p.cancelProcess()
	p.closeChLog <- true
	p.closeChErr <- true
}

func (p *GSProcess) wait() {
	err := p.cmd.Wait()
	if err != nil && err.Error() != ErrorMessageOnKill {
		p.logger.Errorf(p.params.LogWithId(
			"%s: this means gs process was down first, make sure gs is closed successfully"),
			err.Error(),
		)
	}
	p.Close()
	p.closingWait.Wait()
	p.logger.Debug(p.params.LogWithId("gs process successfully closed"))
	p.onProcessClosed()
}

func (p *GSProcess) stdoutLog(errLog bool) {
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

	pipeBroken := false

LOOP:
	for {
		select {
		case <-closeCh:
			break LOOP
		default:
			if pipeBroken {
				continue
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if errLog {
					p.logger.Errorf(p.params.LogWithId(
						"error: %s, waiting for closing error logging goroutine"),
						err.Error(),
					)
				} else {
					p.logger.Errorf(p.params.LogWithId(
						"error: %s, waiting for closing logging goroutine"),
						err.Error(),
					)
				}

				pipeBroken = true
				continue
			}

			if errLog {
				p.logger.Error(p.params.LogWithId(line))
			} else {
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
