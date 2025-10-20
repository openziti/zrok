package controller

import (
	"fmt"
	"net/http"
	"strings"
	"unicode"

	errors2 "github.com/go-openapi/errors"
	"github.com/jaevor/go-nanoid"
	"github.com/jmoiron/sqlx"
	"github.com/michaelquigley/df/dl"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/util"
)

type zrokAuthenticator struct {
	cfg *config.Config
}

func newZrokAuthenticator(cfg *config.Config) *zrokAuthenticator {
	return &zrokAuthenticator{cfg}
}

func (za *zrokAuthenticator) authenticate(token string) (*rest_model_zrok.Principal, error) {
	trx, err := str.Begin()
	if err != nil {
		dl.Errorf("error starting transaction for '%v': %v", token, err)
		return nil, err
	}
	defer func() { _ = trx.Rollback() }()

	if a, err := str.FindAccountWithToken(token, trx); err == nil {
		principal := &rest_model_zrok.Principal{
			ID:        int64(a.Id),
			Token:     a.Token,
			Email:     a.Email,
			Limitless: a.Limitless,
		}
		return principal, nil
	} else {
		// check for admin secret
		if cfg.Admin != nil {
			for _, secret := range cfg.Admin.Secrets {
				if token == secret {
					principal := &rest_model_zrok.Principal{
						ID:    int64(-1),
						Admin: true,
					}
					return principal, nil
				}
			}
		}

		// no match
		dl.Warnf("invalid api key '%v'", token)
		return nil, errors2.New(401, "invalid api key")
	}
}

func createShareToken() (string, error) {
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyz0123456789", 12)
	if err != nil {
		return "", err
	}
	return gen(), nil
}

func CreateToken() (string, error) {
	gen, err := nanoid.CustomASCII("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 12)
	if err != nil {
		return "", err
	}
	return gen(), nil
}

func realRemoteAddress(req *http.Request) string {
	ip := strings.Split(req.RemoteAddr, ":")[0]
	fwdAddress := req.Header.Get("X-Forwarded-For")
	if fwdAddress != "" {
		ip = fwdAddress

		ips := strings.Split(fwdAddress, ", ")
		if len(ips) > 1 {
			ip = ips[0]
		}
	}
	return ip
}

func validatePassword(cfg *config.Config, password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password length: expected (8), got (%d)", len(password))
	}
	if !hasCapital(password) {
		return fmt.Errorf("password requires capital, found none")
	}
	if !hasNumeric(password) {
		return fmt.Errorf("password requires numeric, found none")
	}
	if !strings.ContainsAny(password, "!@#$%^&*()_+-=[]{};':\"\\|,.<>") {
		return fmt.Errorf("password requires special character, found none")
	}
	return nil
}

func hasCapital(check string) bool {
	for _, c := range check {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

func hasNumeric(check string) bool {
	for _, c := range check {
		if unicode.IsDigit(c) {
			return true
		}
	}
	return false
}

// buildFrontendEndpointsForShare retrieves names for a share and builds frontend endpoints
// from those names. Falls back to the deprecated FrontendEndpoint field if no names are
// mapped (for backwards compatibility).
func buildFrontendEndpointsForShare(shareId int, shareToken string, deprecatedEndpoint *string, trx *sqlx.Tx) []string {
	// retrieve names for this share using the new mapping table
	shareNames, err := str.FindNamesForShare(shareId, trx)
	if err != nil {
		dl.Errorf("error finding names for share '%v': %v", shareToken, err)
		// continue without failing the entire request
		shareNames = []*store.NameWithNamespace{}
	}

	// build frontend endpoints from the names
	var frontendEndpoints []string
	for _, sn := range shareNames {
		// use NameInNamespace where namespace.name is the template and name.name is the token
		// this replaces {token} in namespace.name with the actual name.name value
		endpoint := util.NameInNamespace(sn.Name.Name, sn.NamespaceName)
		frontendEndpoints = append(frontendEndpoints, endpoint)
	}

	// fallback to deprecated field if no names are mapped (for backwards compatibility)
	if len(frontendEndpoints) == 0 && deprecatedEndpoint != nil {
		frontendEndpoints = []string{*deprecatedEndpoint}
	}

	return frontendEndpoints
}

// isAccountLimited checks if an account has an active bandwidth limit restriction.
// returns true if the account is currently limited, false otherwise.
func isAccountLimited(accountId int, trx *sqlx.Tx) (bool, error) {
	// check if journal is empty first to avoid unnecessary queries
	jEmpty, err := str.IsBandwidthLimitJournalEmpty(accountId, trx)
	if err != nil {
		return false, err
	}

	// if journal is empty, account is not limited
	if jEmpty {
		return false, nil
	}

	// retrieve the latest journal entry
	je, err := str.FindLatestBandwidthLimitJournal(accountId, trx)
	if err != nil {
		return false, err
	}

	// account is limited if latest entry exists and action is "limit"
	return je != nil && je.Action == store.LimitLimitAction, nil
}
