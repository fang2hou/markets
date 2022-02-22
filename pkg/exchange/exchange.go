package exchange

import (
	"Markets/pkg/database"
	"time"
)

type Exchanger interface {
	Start() error
	Stop() error
}

type Exchange struct {
	running                  bool
	publicDisconnectedSignal chan bool
	name                     string
	database                 *database.Interactor
	aliveSignalInterval      time.Duration
	currencies               []string
}

func (e *Exchange) GetName() string {
	return e.name
}

func (e *Exchange) IsRunning() bool {
	return e.running
}
