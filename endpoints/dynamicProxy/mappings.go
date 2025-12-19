package dynamicProxy

import (
	"context"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/michaelquigley/df/da"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/dynamicProxyController"
	"github.com/pkg/errors"
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

func buildMappings(app *da.Application[*config]) error {
	mappings := newMappings()
	mappings.cfg = app.Cfg
	da.Set(app.C, mappings)
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

func (m *mappings) Link(c *da.Container) error {
	var found bool
	m.amqp, found = da.Get[*amqpSubscriber](c)
	if !found {
		return errors.New("no amqp subscriber found")
	}

	m.ctrl, found = da.Get[*controllerClient](c)
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
	dl.Infof("started")
	defer dl.Infof("stopped")

	// load initial mappings
	start := time.Now()
	mappings, err := m.ctrl.getAllFrontendMappings(m.cfg.FrontendToken, 0)
	if err != nil {
		dl.Fatal(err)
	}
	m.updateMappings(mappings)
	dl.Infof("retrieved '%d' mappings in '%v'", len(mappings), time.Since(start))

	// periodic update loop
	ticker := time.NewTicker(m.cfg.MappingRefreshInterval)
	defer ticker.Stop()

	for {
		dl.ChannelLog("mappings").Debugf("\n%s", m.dumpMappings())
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			// periodically refresh mappings
			start := time.Now()
			highestId := m.getHighestId()
			mappings, err := m.ctrl.getAllFrontendMappings(m.cfg.FrontendToken, highestId)
			if err != nil {
				dl.Errorf("failed to refresh mappings (highest version '%v'): %v", highestId, err)
				continue
			}
			if len(mappings) > 0 {
				m.updateMappings(mappings)
				dl.Warnf("refresh updated '%d' mappings (highest version '%v') in '%v'", len(mappings), highestId, time.Since(start))
			} else {
				dl.Debugf("refresh found no new mappings (highest version '%v') in '%v'", highestId, time.Since(start))
			}

		case update := <-m.amqp.Updates():
			// handle real-time mapping updates from AMQP
			m.handleMappingUpdate(update)
		}
	}
}

func (m *mappings) getHighestId() int64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var highestId int64
	for _, mapping := range m.nameMap {
		if version := mapping.GetId(); version > highestId {
			highestId = version
		}
	}
	return highestId
}

func (m *mappings) handleMappingUpdate(update *dynamicProxyController.Mapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	switch update.Operation {
	case dynamicProxyController.OperationBind:
		mapping := &dynamicProxyController.FrontendMapping{
			Id:         update.Id,
			Name:       update.Name,
			ShareToken: update.ShareToken,
		}
		m.nameMap[mapping.Name] = mapping
		dl.Infof("added mapping: '%v' -> '%v'", mapping.Name, mapping.ShareToken)

	case dynamicProxyController.OperationUnbind:
		delete(m.nameMap, update.Name)
		dl.Infof("removed mapping: '%v'", update.Name)

	default:
		dl.Errorf("unknown mapping operation '%v'", update.Operation)
	}
}

func (m *mappings) updateMappings(frontendMappings []*dynamicProxyController.FrontendMapping) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// populate new mappings
	for _, mapping := range frontendMappings {
		if mapping.Name != "" {
			m.nameMap[mapping.Name] = mapping
		}
		dl.Infof("added mapping: '%v' -> '%v'", mapping.Name, mapping.ShareToken)
	}
}

// dumpMappings returns a formatted table string containing all mapping details
func (m *mappings) dumpMappings() string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetCaption("%d mappings", len(m.nameMap))
	t.AppendHeader(table.Row{"name", "share token"})
	for key, mapping := range m.nameMap {
		t.AppendRow(table.Row{key, mapping.ShareToken})
	}
	return t.Render()
}
