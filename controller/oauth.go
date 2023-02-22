package controller

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/go-openapi/runtime/middleware"
	"github.com/openziti/zrok/rest_server_zrok/operations/share"
	"github.com/sirupsen/logrus"
)

type oauthHandler struct{}

func newOauthHandler() *oauthHandler {
	return &oauthHandler{}
}

func (h *oauthHandler) Handle(params share.OauthAuthenticateParams) middleware.Responder {
	awsUrl := "https:///oauth2/token" // COGNITO URL OR WHATEVER OAUTH PROVIDER URL
	clientId := ""                    // PROVIDER CLIENT ID
	secret := ""                      // PROVIDER CLIENT SECRET
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientId, secret)))
	grant := "authorization_code"
	redirectUri := "http://localhost:18080/api/v1/oauth/authorize"
	// scope := "email"
	data := url.Values{}
	data.Set("client_id", clientId)
	data.Set("grant_type", grant)
	// data.Set("scope", scope)
	data.Set("code", params.Code)
	data.Set("redirect_uri", redirectUri)
	encodedData := data.Encode()

	c := http.Client{}
	req := &http.Request{}
	req.Method = http.MethodPost
	req.URL, _ = url.Parse(awsUrl)
	req.Body = io.NopCloser(strings.NewReader(encodedData))
	req.Header = http.Header{}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
	resp, err := c.Do(req)
	// resp, err := http.Post(awsUrl, "application/x-www-form-urlencoded", strings.NewReader(encodedData))
	logrus.Error(err)
	logrus.Error(resp)
	b, err := ioutil.ReadAll(resp.Body)
	logrus.Error(err)
	logrus.Error(string(b))
	//user, err := cog.GetUser(&cognitoidentityprovider.GetUserInput{
	//	AccessToken: aws.String(params.Code),
	//})
	//if err != nil {
	//	logrus.Error(err)
	//}
	//logrus.Error(user)
	logrus.Error("--------------")
	return share.NewOauthAuthenticateOK()
}

func old(params share.OauthAuthenticateParams) {
	sdkConfig, err := config.LoadDefaultConfig(context.TODO())
	sdkConfig.Region = "us-east-1"
	if err != nil {
		fmt.Println("Couldn't load default configuration. Have you set up your AWS account?")
		fmt.Println(err)
		// return share.NewOauthAuthenticateOK()
	}

	cog := cognitoidentityprovider.NewFromConfig(sdkConfig)
	user, err := cog.GetUser(context.TODO(), &cognitoidentityprovider.GetUserInput{
		AccessToken: &params.Code,
	})
	logrus.Error(err)
	logrus.Error(user)
}
