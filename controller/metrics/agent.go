package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Agent struct {
	cfg      AgentConfig
	events   chan ZitiEventMsg
	src      ZitiEventJsonSource
	srcJoin  chan struct{}
	cache    *cache
	durable  UsageSink
	snks     []UsageSink
	ctx      context.Context
	cancel   context.CancelFunc
	join     chan struct{}
	stopOnce sync.Once
	// one legacy sink may remain blocked; never spawn another while it is running.
	callSlot chan struct{}
}

func NewAgent(cfg *AgentConfig, str *store.Store, ifxCfg *InfluxConfig) (*Agent, error) {
	a := &Agent{cfg: cfg.withDefaults(), callSlot: make(chan struct{}, 1)}
	if v, ok := cfg.Source.(ZitiEventJsonSource); ok {
		a.src = v
	} else {
		return nil, errors.New("invalid event json source")
	}
	a.cache = newShareCache(str)
	a.durable = newInfluxWriter(ifxCfg)
	return a, nil
}

func (a *Agent) AddUsageSink(snk UsageSink) { a.snks = append(a.snks, snk) }

func (a *Agent) Start() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.events = make(chan ZitiEventMsg)
	srcJoin, err := a.src.Start(a.events)
	if err != nil {
		a.cancel()
		return err
	}
	a.srcJoin = srcJoin
	a.join = make(chan struct{})
	go a.run()
	return nil
}

func (a *Agent) run() {
	logrus.Info("started")
	defer close(a.join)
	defer logrus.Info("stopped")
	defer func() {
		if writer, ok := a.durable.(*influxWriter); ok {
			writer.idb.Close()
		}
	}()
	for {
		select {
		case event, ok := <-a.events:
			if !ok {
				return
			}
			if a.ctx.Err() != nil {
				// drain so the source can exit; leave deliveries unsettled for the broker to requeue.
				continue
			}
			a.process(event)
		case <-a.srcJoin:
			return
		}
	}
}

func (a *Agent) process(event ZitiEventMsg) {
	var data ZitiEventJson
	defer func() {
		if recovered := recover(); recovered != nil {
			body := string(data)
			if len(body) > 512 {
				body = body[:512] + "..."
			}
			logrus.Errorf("panic processing usage body '%s': %v", body, recovered)
			a.nack(event)
		}
	}()
	data = event.Data()
	usage, err := Ingest(data)
	if err != nil {
		logrus.Errorf("unable to ingest message: %v", err)
		a.nack(event)
		return
	}
	ctx, cancel := context.WithTimeout(a.ctx, a.cfg.RetryBudget)
	defer cancel()
	if usage.ZitiServiceId != "" {
		if err := a.retry(ctx, func() error {
			err := a.cache.addZrokDetail(ctx, usage)
			if errors.Is(err, errNotAShare) || errors.Is(err, errNoAccount) {
				logrus.Debugf("unable to add zrok detail for: %v: %v", usage.String(), err)
				return nil
			}
			return err
		}); err != nil {
			if a.ctx.Err() != nil {
				return
			}
			logrus.Errorf("unable to add zrok detail (message dropped): %v", err)
			a.nack(event)
			return
		}
	}
	if err := a.retry(ctx, func() error { return a.handle(ctx, a.durable, usage) }); err != nil {
		if a.ctx.Err() != nil {
			// shutdown leaves the active delivery for broker recovery too.
			return
		}
		logrus.Errorf("unable to handle usage (message dropped): %v", err)
		a.nack(event)
		return
	}
	for _, snk := range a.snks {
		if err := a.handle(ctx, snk, usage); err != nil {
			logrus.Errorf("unable to handle usage after durable sink succeeded (delivery will be acknowledged): %v", err)
		}
	}
	if err := event.Ack(); err != nil {
		logrus.Errorf("unable to ack handled message: %v", err)
	}
}

func (a *Agent) retry(ctx context.Context, operation func() error) error {
	backoff := min(a.cfg.RetryInitialBackoff, a.cfg.RetryMaxBackoff)
	for attempt := 1; attempt <= a.cfg.RetryAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := operation()
		if err == nil {
			return nil
		}
		if attempt == a.cfg.RetryAttempts {
			return err
		}
		logrus.Warnf("usage processing attempt '%d' failed: %v", attempt, err)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		if backoff > a.cfg.RetryMaxBackoff/2 {
			backoff = a.cfg.RetryMaxBackoff
		} else {
			backoff *= 2
		}
	}
	return ctx.Err()
}

func (a *Agent) handle(ctx context.Context, snk UsageSink, usage *Usage) error {
	if contextual, ok := snk.(interface {
		HandleContext(context.Context, *Usage) error
	}); ok {
		return contextual.HandleContext(ctx, usage)
	}
	// keep the existing sink interface usable without letting a stuck implementation
	// hold shutdown or create an unbounded number of abandoned goroutines.
	select {
	case <-ctx.Done():
		return ctx.Err()
	case a.callSlot <- struct{}{}:
	}
	type sinkResult struct {
		err        error
		panicValue any
	}
	result := make(chan sinkResult, 1)
	go func() {
		outcome := sinkResult{}
		defer func() {
			// carry a legacy sink's panic back to the per-message recovery boundary.
			outcome.panicValue = recover()
			<-a.callSlot
			result <- outcome
		}()
		outcome.err = snk.Handle(usage)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case outcome := <-result:
		if outcome.panicValue != nil {
			panic(outcome.panicValue)
		}
		return outcome.err
	}
}

func (a *Agent) nack(event ZitiEventMsg) {
	if err := event.Nack(false); err != nil {
		logrus.Errorf("unable to nack message: %v", err)
	}
}

func (a *Agent) Stop() {
	if a.join == nil {
		return
	}
	a.stopOnce.Do(func() {
		a.cancel()
		a.src.Stop()
		<-a.srcJoin
		<-a.join
	})
}
