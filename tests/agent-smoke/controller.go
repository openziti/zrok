//go:build agent_smoke

package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/openziti/zrok/v2/build"
	"github.com/openziti/zrok/v2/rest_client_zrok"
	"github.com/openziti/zrok/v2/rest_client_zrok/environment"
	"github.com/openziti/zrok/v2/rest_client_zrok/metadata"
	"github.com/openziti/zrok/v2/rest_client_zrok/share"
	"github.com/openziti/zrok/v2/rest_model_zrok"
)

// controllerClient makes bounded, authenticated calls with the account token through the generated client. it
// is built from the endpoint alone, so cleanup does not depend on the local environment.
type controllerClient struct {
	zrok  *rest_client_zrok.Zrok
	auth  runtime.ClientAuthInfoWriter
	bound time.Duration
}

func newControllerClient(ctx context.Context, endpoint, accountToken string, bound time.Duration) (*controllerClient, error) {
	apiURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse api endpoint '%v': %w", endpoint, err)
	}
	if apiURL.Scheme == "" || apiURL.Host == "" {
		return nil, fmt.Errorf("api endpoint '%v' needs a scheme and host", endpoint)
	}
	transport := httptransport.New(apiURL.Host, "/api/v2", []string{apiURL.Scheme})
	transport.Producers["application/zrok.v1+json"] = runtime.JSONProducer()
	transport.Consumers["application/zrok.v1+json"] = runtime.JSONConsumer()
	c := &controllerClient{
		zrok:  rest_client_zrok.New(transport, strfmt.Default),
		auth:  httptransport.APIKeyAuth("X-TOKEN", "header", accountToken),
		bound: bound,
	}
	reqCtx, cancel := context.WithTimeout(ctx, bound)
	defer cancel()
	params := metadata.NewClientVersionCheckParamsWithContext(reqCtx)
	params.Body.ClientVersion = build.String()
	if _, err := c.zrok.Metadata.ClientVersionCheck(params); err != nil {
		return nil, fmt.Errorf("controller refused client version '%v': %w", build.String(), err)
	}
	return c, nil
}

func (c *controllerClient) overview(ctx context.Context) (*rest_model_zrok.Overview, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.bound)
	defer cancel()
	resp, err := c.zrok.Metadata.Overview(metadata.NewOverviewParamsWithContext(reqCtx), c.auth)
	if err != nil {
		return nil, fmt.Errorf("account overview failed: %w", err)
	}
	if resp == nil || resp.Payload == nil {
		return nil, errors.New("account overview returned no payload")
	}
	return resp.Payload, nil
}

func (c *controllerClient) names(ctx context.Context) ([]*rest_model_zrok.Name, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.bound)
	defer cancel()
	resp, err := c.zrok.Share.ListAllNames(share.NewListAllNamesParamsWithContext(reqCtx), c.auth)
	if err != nil {
		return nil, fmt.Errorf("account name listing failed: %w", err)
	}
	if resp == nil {
		return nil, errors.New("account name listing returned no response")
	}
	return resp.Payload, nil
}

func (c *controllerClient) unshare(ctx context.Context, envZID, token string) error {
	reqCtx, cancel := context.WithTimeout(ctx, c.bound)
	defer cancel()
	params := share.NewUnshareParamsWithContext(reqCtx)
	params.Body.EnvZID = envZID
	params.Body.ShareToken = token
	if _, err := c.zrok.Share.Unshare(params, c.auth); err != nil {
		return fmt.Errorf("unshare of '%v' failed: %w", token, err)
	}
	return nil
}

func (c *controllerClient) deleteName(ctx context.Context, namespace, name string) error {
	reqCtx, cancel := context.WithTimeout(ctx, c.bound)
	defer cancel()
	params := share.NewDeleteShareNameParamsWithContext(reqCtx)
	params.Body.NamespaceToken = namespace
	params.Body.Name = name
	if _, err := c.zrok.Share.DeleteShareName(params, c.auth); err != nil {
		return fmt.Errorf("deletion of name '%v:%v' failed: %w", namespace, name, err)
	}
	return nil
}

func (c *controllerClient) disable(ctx context.Context, envZID string) error {
	reqCtx, cancel := context.WithTimeout(ctx, c.bound)
	defer cancel()
	params := environment.NewDisableParamsWithContext(reqCtx)
	params.Body.Identity = envZID
	if _, err := c.zrok.Environment.Disable(params, c.auth); err != nil {
		return fmt.Errorf("disable of environment '%v' failed: %w", envZID, err)
	}
	return nil
}

// accountView is one successful pair of reads: the overview and the account's name listing.
type accountView struct {
	overview *rest_model_zrok.Overview
	names    []*rest_model_zrok.Name
}

func (c *controllerClient) view(ctx context.Context) (*accountView, error) {
	ov, err := c.overview(ctx)
	if err != nil {
		return nil, err
	}
	names, err := c.names(ctx)
	if err != nil {
		return nil, err
	}
	return &accountView{overview: ov, names: names}, nil
}

func (v *accountView) environment(zID string) *rest_model_zrok.EnvironmentAndResources {
	for _, resources := range v.overview.Environments {
		if resources != nil && resources.Environment != nil && resources.Environment.ZID == zID {
			return resources
		}
	}
	return nil
}

func (v *accountView) environmentsDescribed(description string) []string {
	var zIDs []string
	for _, resources := range v.overview.Environments {
		if resources != nil && resources.Environment != nil && resources.Environment.Description == description {
			zIDs = append(zIDs, resources.Environment.ZID)
		}
	}
	return zIDs
}

// tokenLocations lists every place the account overview mentions a share token: shares, frontends and names in
// any environment.
func (v *accountView) tokenLocations(token string) []string {
	var found []string
	for _, resources := range v.overview.Environments {
		if resources == nil {
			continue
		}
		zID := ""
		if resources.Environment != nil {
			zID = resources.Environment.ZID
		}
		for _, shr := range resources.Shares {
			if shr != nil && shr.ShareToken == token {
				found = append(found, fmt.Sprintf("share in environment '%v'", zID))
			}
		}
		for _, fe := range resources.Frontends {
			if fe != nil && fe.ShareToken == token {
				found = append(found, fmt.Sprintf("frontend '%v' in environment '%v'", fe.FrontendToken, zID))
			}
		}
	}
	for _, name := range v.overview.Names {
		if name != nil && name.ShareToken == token {
			found = append(found, fmt.Sprintf("overview name '%v:%v'", name.NamespaceToken, name.Name))
		}
	}
	for _, name := range v.names {
		if name != nil && name.ShareToken == token {
			found = append(found, fmt.Sprintf("listed name '%v:%v'", name.NamespaceToken, name.Name))
		}
	}
	return found
}

func (v *accountView) matchingNames(namespace, name string) []*rest_model_zrok.Name {
	var found []*rest_model_zrok.Name
	for _, entry := range v.names {
		if entry != nil && entry.NamespaceToken == namespace && entry.Name == name {
			found = append(found, entry)
		}
	}
	return found
}
