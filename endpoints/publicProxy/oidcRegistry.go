package publicProxy

// oidcProviderRegistry maps provider names to their sessionRefresher implementations.
// Each OIDC provider's configure() function registers itself here so that
// tryInlineRefresh can find the correct provider for a given session token.
var oidcProviderRegistry = map[string]sessionRefresher{}
