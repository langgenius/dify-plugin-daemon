package plugin_entities

type DatasourceDeclaration struct {
	// Identity etc.
	// TDB

	CredentialsSchema []ProviderConfig `json:"credentials_schema" yaml:"credentials_schema" validate:"omitempty,dive"`
	OAuthSchema       *OAuthSchema     `json:"oauth_schema" yaml:"oauth_schema" validate:"omitempty,dive"`
}
