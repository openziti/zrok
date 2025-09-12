package dynamicProxy

import (
	"context"
	"sync"
	"time"

	"github.com/michaelquigley/df"
	"github.com/openziti/zrok/controller/dynamicProxyController"
	"github.com/openziti/zrok/dynamicProxyModel"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type mappings struct {
	cfg     *config
	amqp    *amqpSubscriber
	ctrl    *controllerClient
	ctx     context.Context
	cancel  context.CancelFunc
	mutex   sync.RWMutex
	nameMap map[string]*dynamicProxyController.FrontendMapping
}

func buildMappings(app *df.Application[*config]) error {
	mappings := newMappings()
	df.Set(app.C, mappings)
	return nil
}

func newMappings() *mappings {
	return &mappings{
		nameMap: make(map[string]*dynamicProxyController.FrontendMapping),
	}
}

func (m *mappings) getMapping(name string) (*dynamicProxyController.FrontendMapping, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	mapping, exists := m.nameMap[name]
	return mapping, exists
}

func (m *mappings) Link(c *df.Container) error {
	var found bool
	m.cfg, found = df.Get[*config](c)
	if !found {
		return errors.New("no config found")
	}

	m.amqp, found = df.Get[*amqpSubscriber](c)
	if !found {
		return errors.New("no amqp subscriber found")
	}

	m.ctrl, found = df.Get[*controllerClient](c)
	if !found {
		return errors.New("no controller client found")
	}
	return nil
}

func (m *mappings) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	go m.run()
	return nil
}

func (m *mappings) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}

func (m *mappings) run() {
	logrus.Infof("started")
	defer logrus.Infof("stopped")

	// load initial mappings
	start := time.Now()
	mappings, err := m.ctrl.getAllFrontendMappings(m.cfg.AmqpSubscriber.FrontendToken, 0)
	if err != nil {
		logrus.Fatal(err)
	}
	m.updateMappings(mappings)
	logrus.Infof("retrieved '%d' mappings in '%v'", len(mappings), time.Since(start))

	// periodic update loop
	ticker := time.NewTicker(m.cfg.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// periodically refresh mappings
			logrus.Info("refreshing")
			mappings, err := m.ctrl.getAllFrontendMappings(m.cfg.AmqpSubscriber.FrontendToken, m.getHighestVersion())
			if err != nil {
				logrus.Errorf("failed to refresh mappings: %v", err)
				continue
			}
			m.updateMappings(mappings)
			logrus.Debugf("updated '%d' mappings", len(mappings))

		case update := <-m.amqp.Updates():
			// handle real-time mapping updates from AMQP
			m.handleMappingUpdate(update)
		}
	}
}

func (m *mappings) getHighestVersion() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var highestVersion int64
	for _, mapping := range m.nameMap {
		if version := mapping.GetVersion(); version > highestVersion {
			highestVersion = version
		}
	}
	return highestVersion
}

func (m *mappings) handleMappingUpdate(update *dynamicProxyModel.Mapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch update.Operation {
	case dynamicProxyModel.OperationBind:
		mapping := &dynamicProxyController.FrontendMapping{
			Name:       update.Name,
			Version:    update.Version,
			ShareToken: update.ShareToken,
		}
		m.nameMap[mapping.Name] = mapping
		logrus.Infof("added mapping: '%v' -> '%v' (%v)", mapping.Name, mapping.ShareToken, mapping.Version)

	case dynamicProxyModel.OperationUnbind:
		delete(m.nameMap, update.Name)
		logrus.Infof("removed mapping: '%v'", update.Name)

	default:
		logrus.Errorf("unknown mapping operation '%v'", update.Operation)
	}

	logrus.Infof("'%d' mappings in table", len(m.nameMap))
}

func (m *mappings) updateMappings(frontendMappings []*dynamicProxyController.FrontendMapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// clear existing mappings
	m.nameMap = make(map[string]*dynamicProxyController.FrontendMapping)

	// populate with new mappings
	for _, mapping := range frontendMappings {
		if mapping.Name != "" {
			m.nameMap[mapping.Name] = mapping
		}
		logrus.Infof("added mapping: '%v' -> '%v' (%v)", mapping.Name, mapping.ShareToken, mapping.Version)
	}

	logrus.Infof("'%d' mappings in table", len(m.nameMap))
}
