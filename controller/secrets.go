package controller

import (
	"context"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/config"
	"github.com/openziti/zrok/controller/secretsGrpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func startSecretsListener(cfg *config.Config) {
	if cfg != nil && cfg.Secrets != nil {
		zcfg, err := ziti.NewConfigFromFile(cfg.Secrets.IdentityPath)
		if err != nil {
			logrus.Errorf("error loading secrets listener identity '%v': %v", cfg.Secrets.IdentityPath, err)
			return
		}
		zctx, err := ziti.NewContext(zcfg)
		if err != nil {
			logrus.Errorf("error creating ziti context: %v", err)
			return
		}
		l, err := zctx.Listen(cfg.Secrets.ServiceName)
		if err != nil {
			logrus.Errorf("error listening on '%v': %v", cfg.Secrets.ServiceName, err)
			return
		}

		srv := grpc.NewServer()
		secretsGrpc.RegisterSecretsServer(srv, &secretsGrpcImpl{})
		logrus.Infof("starting secrets listener")
		if err := srv.Serve(l); err != nil {
			logrus.Errorf("error serving '%v': %v", cfg.Secrets.ServiceName, err)
			return
		}

	} else {
		logrus.Warnf("secrets listener disabled")
	}
}

type secretsGrpcImpl struct {
	secretsGrpc.UnimplementedSecretsServer
}

func (i *secretsGrpcImpl) FetchSecrets(_ context.Context, req *secretsGrpc.SecretsRequest) (*secretsGrpc.SecretsResponse, error) {
	logrus.Infof("request for secrets for '%v'", req.ShareToken)

	trx, err := str.Begin()
	if err != nil {
		return nil, err
	}
	defer trx.Rollback()

	shr, err := str.FindShareWithToken(req.ShareToken, trx)
	if err != nil {
		return nil, err
	}

	secrets, err := str.GetSecrets(shr.Id, trx)
	if err != nil {
		return nil, err
	}

	out := &secretsGrpc.SecretsResponse{}
	for _, secret := range secrets.Secrets {
		out.Secrets = append(out.Secrets, &secretsGrpc.Secret{Key: secret.Key, Value: secret.Value})
	}

	return out, nil
}
