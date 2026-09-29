package controller

import (
	"strings"

	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/v2/controller/automation"
	"github.com/pkg/errors"
)

type zitiObjectKind string

const (
	zitiConfig                  zitiObjectKind = "config"
	zitiService                 zitiObjectKind = "service"
	zitiServicePolicy           zitiObjectKind = "service policy"
	zitiServiceEdgeRouterPolicy zitiObjectKind = "service edge router policy"
)

type zitiObject struct {
	kind zitiObjectKind
	id   string
}

// zitiCompensation records the ziti objects one request created, so a request that fails after
// allocating deletes exactly those objects. it deletes by recorded id and never by tag: two requests
// racing for the same private share token both pass the availability check and carry the same token
// tag, so a tag-scoped cleanup by the loser would delete the winner's objects.
type zitiCompensation struct {
	token   string
	objects []zitiObject
}

func newZitiCompensation(token string) *zitiCompensation {
	return &zitiCompensation{token: token}
}

func (c *zitiCompensation) add(kind zitiObjectKind, id string) {
	c.objects = append(c.objects, zitiObject{kind, id})
}

// run deletes the recorded objects in reverse order of creation. it continues past failures, since
// a partial compensation is better than none, and never changes the caller's response.
func (c *zitiCompensation) run(ziti *automation.ZitiAutomation) {
	var deleted []string
	for i := len(c.objects) - 1; i >= 0; i-- {
		obj := c.objects[i]
		err := c.delete(ziti, obj)
		switch {
		case err == nil:
			deleted = append(deleted, obj.id)
		case automation.IsNotFound(err):
			dl.Debugf("compensating share '%v': %v '%v' already deleted", c.token, obj.kind, obj.id)
		default:
			dl.Errorf("compensating share '%v': error deleting %v '%v': %v", c.token, obj.kind, obj.id, err)
		}
	}
	dl.Infof("compensated failed share '%v': deleted ziti objects '%v'", c.token, strings.Join(deleted, ", "))
}

func (c *zitiCompensation) delete(ziti *automation.ZitiAutomation, obj zitiObject) error {
	switch obj.kind {
	case zitiConfig:
		return ziti.Configs.Delete(obj.id)
	case zitiService:
		return ziti.Services.Delete(obj.id)
	case zitiServicePolicy:
		return ziti.ServicePolicies.Delete(obj.id)
	case zitiServiceEdgeRouterPolicy:
		return ziti.ServiceEdgeRouterPolicies.Delete(obj.id)
	default:
		return errors.Errorf("unknown ziti object kind '%v'", obj.kind)
	}
}
