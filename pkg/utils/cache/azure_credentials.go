package cache

import (
	"fmt"

	entraid "github.com/redis/go-redis-entraid"
	"github.com/redis/go-redis-entraid/identity"
	"github.com/redis/go-redis/v9/auth"
)

// NewAzureEntraIDCredentialsProvider creates a StreamingCredentialsProvider
// that authenticates to Azure Cache for Redis via Entra ID
// (DefaultAzureCredential), matching the behaviour of the Python
// redis-entraid library used in the main Dify API server.
func NewAzureEntraIDCredentialsProvider() (auth.StreamingCredentialsProvider, error) {
	provider, err := entraid.NewDefaultAzureCredentialsProvider(
		entraid.DefaultAzureCredentialsProviderOptions{
			DefaultAzureIdentityProviderOptions: identity.DefaultAzureIdentityProviderOptions{
				Scopes: []string{"https://redis.azure.com/.default"},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("azure entraid: failed to create credentials provider: %w", err)
	}
	return provider, nil
}
