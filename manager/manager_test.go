package manager

import (
	"testing"

	"github.com/barracudanetworks/GoWorker/config"
)

var (
	TEST_PROVIDER_CONFIG = config.Config{}
	TEST_WORKER_CONFIG   = config.Config{}

	TEST_MANAGER = func() *Manager {
		m := NewManager()
		m.Init(config.DefaultAppConfig())
		return m
	}()
)

func TestManagerCreation(t *testing.T) {
	m := TEST_MANAGER
	if m == nil {
		t.Fail()
	}

	// make sure no values are nil
	if m.KillChan == nil {
		t.Fail()
	}
	if m.Providers == nil {
		t.Fail()
	}
	if m.allWorkers == nil {
		t.Fail()
	}
	if m.readyWorkers == nil {
		t.Fail()
	}
	if m.currentWorkers == nil {
		t.Fail()
	}
}

func TestManagerKill(t *testing.T) {
	go TEST_MANAGER.Manage()
	TEST_MANAGER.KillChan <- struct{}{}
}
