package controller

import (
	"strings"

	"github.com/openziti/zrok/controller/automation"
	zrok_config "github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func GC(inCfg *zrok_config.Config) error {
	cfg = inCfg
	if v, err := store.Open(cfg.Store); err == nil {
		str = v
	} else {
		return errors.Wrap(err, "error opening store")
	}
	defer func() {
		if err := str.Close(); err != nil {
			logrus.Errorf("error closing store: %v", err)
		}
	}()
	tx, err := str.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	sshrs, err := str.FindAllShares(tx)
	if err != nil {
		return err
	}
	liveMap := make(map[string]struct{})
	for _, sshr := range sshrs {
		liveMap[sshr.Token] = struct{}{}
	}
	ziti, err := automation.NewZitiAutomation(cfg.Ziti)
	if err != nil {
		return err
	}
	if err := gcServices(ziti, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting services")
	}
	if err := gcServiceEdgeRouterPolicies(ziti, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting service edge router policies")
	}
	if err := gcServicePolicies(ziti, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting service policies")
	}
	if err := gcConfigs(ziti, liveMap); err != nil {
		return errors.Wrap(err, "error garbage collecting configs")
	}
	return nil
}

func gcServices(ziti *automation.ZitiAutomation, liveMap map[string]struct{}) error {
	filterOpts := &automation.FilterOptions{
		Filter: "tags.zrok != null",
		Limit:  0,
		Offset: 0,
	}

	services, err := ziti.Services.Find(filterOpts)
	if err != nil {
		return errors.Wrap(err, "error listing services")
	}

	for _, svc := range services {
		if _, found := liveMap[*svc.Name]; !found {
			logrus.Infof("garbage collecting, zitiSvcId='%v', zrokSvcId='%v'", *svc.ID, *svc.Name)

			// delete service edge router policies for share
			serpFilter := "name=\"" + *svc.Name + "\""
			if err := ziti.ServiceEdgeRouterPolicies.DeleteWithFilter(serpFilter); err != nil {
				logrus.Errorf("error garbage collecting service edge router policy: %v", err)
			}

			// delete dial service policies for share
			dialFilter := "name=\"" + *svc.Name + "-dial\""
			if err := ziti.ServicePolicies.DeleteWithFilter(dialFilter); err != nil {
				logrus.Errorf("error garbage collecting service dial policy: %v", err)
			}

			// delete bind service policies for share
			bindFilter := "name=\"" + *svc.Name + "-bind\""
			if err := ziti.ServicePolicies.DeleteWithFilter(bindFilter); err != nil {
				logrus.Errorf("error garbage collecting service bind policy: %v", err)
			}

			// delete configs for share
			configFilter := "name=\"" + *svc.Name + "\""
			if err := ziti.Configs.DeleteWithFilter(configFilter); err != nil {
				logrus.Errorf("error garbage collecting config: %v", err)
			}

			// delete service
			if err := ziti.Services.Delete(*svc.ID); err != nil {
				logrus.Errorf("error garbage collecting service: %v", err)
			}
		} else {
			logrus.Infof("remaining live, zitiSvcId='%v', zrokSvcId='%v'", *svc.ID, *svc.Name)
		}
	}
	return nil
}

func gcServiceEdgeRouterPolicies(ziti *automation.ZitiAutomation, liveMap map[string]struct{}) error {
	filterOpts := &automation.FilterOptions{
		Filter: "tags.zrok != null",
		Limit:  0,
		Offset: 0,
	}

	policies, err := ziti.ServiceEdgeRouterPolicies.Find(filterOpts)
	if err != nil {
		return errors.Wrap(err, "error listing service edge router policies")
	}

	for _, serp := range policies {
		if _, found := liveMap[*serp.Name]; !found {
			logrus.Infof("garbage collecting, svcId='%v'", *serp.Name)
			filter := "name=\"" + *serp.Name + "\""
			if err := ziti.ServiceEdgeRouterPolicies.DeleteWithFilter(filter); err != nil {
				logrus.Errorf("error garbage collecting service edge router policy: %v", err)
			}
		} else {
			logrus.Infof("remaining live, svcId='%v'", *serp.Name)
		}
	}
	return nil
}

func gcServicePolicies(ziti *automation.ZitiAutomation, liveMap map[string]struct{}) error {
	filterOpts := &automation.FilterOptions{
		Filter: "tags.zrok != null",
		Limit:  0,
		Offset: 0,
	}

	policies, err := ziti.ServicePolicies.Find(filterOpts)
	if err != nil {
		return errors.Wrap(err, "error listing service policies")
	}

	for _, sp := range policies {
		spName := strings.Split(*sp.Name, "-")[0]
		if _, found := liveMap[spName]; !found {
			logrus.Infof("garbage collecting, svcId='%v'", spName)
			deleteFilter := "id=\"" + *sp.ID + "\""
			if err := ziti.ServicePolicies.DeleteWithFilter(deleteFilter); err != nil {
				logrus.Errorf("error garbage collecting service policy: %v", err)
			}
		} else {
			logrus.Infof("remaining live, svcId='%v'", spName)
		}
	}
	return nil
}

func gcConfigs(ziti *automation.ZitiAutomation, liveMap map[string]struct{}) error {
	filterOpts := &automation.FilterOptions{
		Filter: "tags.zrok != null",
		Limit:  0,
		Offset: 0,
	}

	configs, err := ziti.Configs.Find(filterOpts)
	if err != nil {
		return errors.Wrap(err, "error listing configs")
	}

	for _, c := range configs {
		if _, found := liveMap[*c.Name]; !found {
			configFilter := "name=\"" + *c.Name + "\""
			if err := ziti.Configs.DeleteWithFilter(configFilter); err != nil {
				logrus.Errorf("error garbage collecting config: %v", err)
			}
		} else {
			logrus.Infof("remaining live, svcId='%v'", *c.Name)
		}
	}
	return nil
}
