package limits

import (
	"testing"

	"github.com/openziti/zrok/controller/emailUi"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/sdk/golang/sdk"
	"github.com/stretchr/testify/assert"
)

func TestNewWarningActionSimple(t *testing.T) {
	emailCfg := &emailUi.Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "password",
		From:     "noreply@example.com",
	}
	str := &store.Store{}

	action := newWarningAction(emailCfg, str)

	assert.NotNil(t, action)
	assert.Equal(t, str, action.str)
	assert.Equal(t, emailCfg, action.cfg)
}

func TestWarningAction_InterfaceCompliance(t *testing.T) {
	emailCfg := &emailUi.Config{}
	str := &store.Store{}

	action := newWarningAction(emailCfg, str)

	// verify it implements the AccountAction interface
	var _ AccountAction = action

	acct := &store.Account{
		Model: store.Model{Id: 1},
		Email: "test@example.com",
	}

	bwc := &configBandwidthClass{
		periodInMinutes: 60,
		bw: &Bandwidth{
			Rx:    100,
			Tx:    100,
			Total: 200,
		},
		limitAction: store.WarningLimitAction,
	}

	ul := &userLimits{
		scopes: make(map[sdk.BackendMode]*store.LimitClass),
	}

	// test with no email config should succeed (just logs warning)
	action.cfg = nil
	err := action.HandleAccount(acct, 50, 75, bwc, ul, nil)
	assert.NoError(t, err)
}

func TestDetailMessageMethods(t *testing.T) {
	dm := newDetailMessage()

	assert.NotNil(t, dm)
	assert.Empty(t, dm.lines)

	// test append
	dm = dm.append("First line")
	assert.Len(t, dm.lines, 1)
	assert.Equal(t, "First line", dm.lines[0])

	dm = dm.append("Second line with %s and %d", "format", 42)
	assert.Len(t, dm.lines, 2)
	assert.Equal(t, "Second line with format and 42", dm.lines[1])

	// test plain text output
	plainText := dm.plain()
	expectedPlain := "First line\n\nSecond line with format and 42\n\n"
	assert.Equal(t, expectedPlain, plainText)

	// test HTML output
	htmlText := dm.html()
	expectedHTML := `<p style="text-align: left;">First line</p>
<p style="text-align: left;">Second line with format and 42</p>
`
	assert.Equal(t, expectedHTML, htmlText)
}

func TestDetailMessage_EmptyMessage(t *testing.T) {
	dm := newDetailMessage()

	plainText := dm.plain()
	assert.Equal(t, "", plainText)

	htmlText := dm.html()
	assert.Equal(t, "", htmlText)
}