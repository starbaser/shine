package config

import (
	"log"
	"os"
	"time"
)

type Watcher struct {
	configPath string
	lastMod    time.Time
	onChange   func(*Config)
	ticker     *time.Ticker
	stop       chan bool
}

func NewWatcher(configPath string, onChange func(*Config)) (*Watcher, error) {
	info, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	return &Watcher{
		configPath: configPath,
		lastMod:    info.ModTime(),
		onChange:   onChange,
		stop:       make(chan bool),
	}, nil
}

// Start begins watching the config file.
// Checks for changes every second using stat polling.
func (w *Watcher) Start() {
	w.ticker = time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.checkForChanges()
			case <-w.stop:
				return
			}
		}
	}()
}

func (w *Watcher) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stop)
}

func (w *Watcher) checkForChanges() {
	info, err := os.Stat(w.configPath)
	if err != nil {
		log.Printf("Error checking config file: %v", err)
		return
	}

	if info.ModTime().After(w.lastMod) {
		w.lastMod = info.ModTime()

		cfg, err := Load(w.configPath)
		if err != nil {
			log.Printf("Error reloading config: %v", err)
			return
		}

		log.Printf("Config file changed, reloading...")
		w.onChange(cfg)
	}
}
