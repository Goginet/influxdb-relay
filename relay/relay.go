package relay

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Service struct {
	relays map[string]Relay
}

const (
	DefaultCollectInterval = 30 * time.Second
)

var collector MetricsCollector

func New(config Config) (*Service, error) {
	s := new(Service)
	s.relays = make(map[string]Relay)

	if config.MetricsCollector.Url != "" {
		if config.MetricsCollector.Name == "" {
			config.MetricsCollector.Name = config.MetricsCollector.Url
		}

		interval, err := time.ParseDuration(config.MetricsCollector.Interval)
		if err != nil {
			return nil, fmt.Errorf("error parsing metrics collect interval %v", err)
		}
		interval = DefaultCollectInterval

		collector = MetricsCollector{
			name:     config.MetricsCollector.Name,
			url:      config.MetricsCollector.Url,
			interval: interval,
		}
	}

	for _, cfg := range config.HTTPRelays {
		h, err := NewHTTP(cfg)
		if err != nil {
			return nil, err
		}
		if s.relays[h.Name()] != nil {
			return nil, fmt.Errorf("duplicate relay: %q", h.Name())
		}
		s.relays[h.Name()] = h
	}

	for _, cfg := range config.UDPRelays {
		u, err := NewUDP(cfg)
		if err != nil {
			return nil, err
		}
		if s.relays[u.Name()] != nil {
			return nil, fmt.Errorf("duplicate relay: %q", u.Name())
		}
		s.relays[u.Name()] = u
	}

	return s, nil
}

func (s *Service) Run() {
	var wg sync.WaitGroup
	wg.Add(len(s.relays))

	for k := range s.relays {
		relay := s.relays[k]
		go func() {
			defer wg.Done()

			if err := relay.Run(); err != nil {
				log.Printf("Error running relay %q: %v", relay.Name(), err)
			}
		}()
	}

	wg.Wait()
}

func (s *Service) Stop() {
	for _, v := range s.relays {
		v.Stop()
	}
}

type Relay interface {
	Name() string
	Run() error
	Stop() error
}
