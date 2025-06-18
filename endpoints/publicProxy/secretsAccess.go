package publicProxy

import (
	"context"
	"net"
	"time"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/zrok/controller/secretsGrpc"
	"github.com/viccon/sturdyc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

type Secret struct {
	Key   string
	Value string
}

func GetSecrets(shareToken string, cfg *Config) ([]Secret, error) {
	cacheClient := sturdyc.New[[]Secret](cfg.SecretsCache.Capacity, cfg.SecretsCache.Shards, cfg.SecretsCache.TTL, cfg.SecretsCache.EvictionPercentage)
	fetch := func(ctx context.Context) ([]Secret, error) {
		opts := []grpc.DialOption{
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				zcfg, err := ziti.NewConfigFromFile(cfg.SecretsAccess.IdentityPath)
				if err != nil {
					return nil, err
				}
				zctx, err := ziti.NewContext(zcfg)
				if err != nil {
					return nil, err
				}
				conn, err := zctx.DialWithOptions(addr, &ziti.DialOptions{ConnectTimeout: 30 * time.Second})
				if err != nil {
					return nil, err
				}
				return conn, nil
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}
		resolver.SetDefaultScheme("passthrough")
		conn, err := grpc.NewClient(cfg.SecretsAccess.ServiceName, opts...)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		client := secretsGrpc.NewSecretsClient(conn)
		resp, err := client.FetchSecrets(ctx, &secretsGrpc.SecretsRequest{ShareToken: shareToken})
		if err != nil {
			return nil, err
		}
		var secrets []Secret
		for _, secret := range resp.GetSecrets() {
			secrets = append(secrets, Secret{Key: secret.Key, Value: secret.Value})
		}
		return secrets, nil
	}
	return cacheClient.GetOrFetch(context.Background(), shareToken, fetch)
}
