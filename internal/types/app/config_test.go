package app

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisTLSConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		setupCA        bool
		expectedTLS    bool
		expectedError  bool
		errorContains  string
		validateConfig func(t *testing.T, tlsConfig *tls.Config)
	}{
		{
			name: "SSL disabled - returns nil",
			config: Config{
				RedisUseSsl: false,
			},
			expectedTLS:   false,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.Nil(t, tlsConfig)
			},
		},
		{
			name: "SSL enabled with default CERT_REQUIRED",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
				assert.False(t, tlsConfig.InsecureSkipVerify)
				assert.Nil(t, tlsConfig.RootCAs)
			},
		},
		{
			name: "SSL enabled with CERT_REQUIRED explicit",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_REQUIRED",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with CERT_REQUIRED lowercase",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "cert_required",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with CERT_OPTIONAL",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_OPTIONAL",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with CERT_NONE",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_NONE",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
				assert.True(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with CERT_NONE lowercase",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "cert_none",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.True(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with whitespace in CERT_REQS",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "  CERT_REQUIRED  ",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with invalid CERT_REQS value",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "INVALID_VALUE",
			},
			expectedTLS:   false,
			expectedError: true,
			errorContains: "invalid REDIS_SSL_CERT_REQS value",
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.Nil(t, tlsConfig)
			},
		},
		{
			name: "SSL enabled with custom CA certificate",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_REQUIRED",
			},
			setupCA:       true,
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.NotNil(t, tlsConfig.RootCAs)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
		{
			name: "SSL enabled with non-existent CA certificate file",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_REQUIRED",
				RedisSSLCACerts:  "/nonexistent/ca.crt",
			},
			expectedTLS:   false,
			expectedError: true,
			errorContains: "read REDIS_SSL_CA_CERTS",
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.Nil(t, tlsConfig)
			},
		},
		{
			name: "SSL enabled with empty CA certificate path (whitespace only)",
			config: Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: "CERT_REQUIRED",
				RedisSSLCACerts:  "   ",
			},
			expectedTLS:   true,
			expectedError: false,
			validateConfig: func(t *testing.T, tlsConfig *tls.Config) {
				assert.NotNil(t, tlsConfig)
				assert.Nil(t, tlsConfig.RootCAs)
				assert.False(t, tlsConfig.InsecureSkipVerify)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup CA certificate if needed
			if tt.setupCA {
				tempDir, err := os.MkdirTemp("", "redis-tls-test-*")
				require.NoError(t, err)
				defer os.RemoveAll(tempDir)

				caFile := filepath.Join(tempDir, "ca.crt")
				err = os.WriteFile(caFile, []byte(testCACert), 0644)
				require.NoError(t, err)

				tt.config.RedisSSLCACerts = caFile
			}

			// Call RedisTLSConfig
			tlsConfig, err := tt.config.RedisTLSConfig()

			// Check error
			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			// Check TLS config existence
			if tt.expectedTLS {
				assert.NotNil(t, tlsConfig)
			}

			// Run custom validation
			if tt.validateConfig != nil {
				tt.validateConfig(t, tlsConfig)
			}
		})
	}
}

func TestRedisTLSConfigWithInvalidCAContent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "redis-tls-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	caFile := filepath.Join(tempDir, "invalid-ca.crt")
	err = os.WriteFile(caFile, []byte("invalid certificate content"), 0644)
	require.NoError(t, err)

	config := Config{
		RedisUseSsl:      true,
		RedisSSLCertReqs: "CERT_REQUIRED",
		RedisSSLCACerts:  caFile,
	}

	tlsConfig, err := config.RedisTLSConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to append CA certs")
	assert.Nil(t, tlsConfig)
}

func TestRedisTLSConfigCombinations(t *testing.T) {
	// Test that MinVersion is always set to TLS 1.2 when SSL is enabled
	tests := []struct {
		name     string
		certReqs string
	}{
		{"CERT_NONE", "CERT_NONE"},
		{"CERT_OPTIONAL", "CERT_OPTIONAL"},
		{"CERT_REQUIRED", "CERT_REQUIRED"},
		{"Empty (default)", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: tt.certReqs,
			}

			tlsConfig, err := config.RedisTLSConfig()
			require.NoError(t, err)
			require.NotNil(t, tlsConfig)
			assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion,
				"MinVersion should always be TLS 1.2")
		})
	}
}

func TestRedisTLSConfigCaseInsensitivity(t *testing.T) {
	cases := []string{
		"CERT_NONE",
		"cert_none",
		"Cert_None",
		"CERT_OPTIONAL",
		"cert_optional",
		"Cert_Optional",
		"CERT_REQUIRED",
		"cert_required",
		"Cert_Required",
	}

	for _, certReqs := range cases {
		t.Run(certReqs, func(t *testing.T) {
			config := Config{
				RedisUseSsl:      true,
				RedisSSLCertReqs: certReqs,
			}

			tlsConfig, err := config.RedisTLSConfig()
			assert.NoError(t, err, "Should accept case-insensitive cert requirements")
			assert.NotNil(t, tlsConfig)
		})
	}
}

// Test CA certificate in PEM format (self-signed certificate for testing)
const testCACert = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRqqSMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwMjIyMDUwNzQ4WhcNMjcwMjIwMDUwNzQ4WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAyVuOPQaVxQjBBg7CYBaKzVnfJKG6iFUvuQfJqXL9BKE8bTSASPrHGEFL
WLTh7MvYhE9nMLvPB7FzJFdI5K5pxWRmIFxM6pxEPGvGnF0Bc+cGSN2UFHQZW0Rc
qxZJu5Hbv9YMsGHG+VPx8mWfD8LDHD+M5lKVrNWfXeNXKXlPfnqP8N0vQAUv7z5Y
y0N0OJCcZqW3nFqGNAFLVqL3MzLz+2thqBKs3vG2VQ0NI0aL9T4eqN1qQXqIHWnQ
tLQgLhNCJVQxcLEu2KHyBXUJrI8FnxVAoOyPKq5wJjVPEwqBp5HWpqnNKO6eFXKE
l3BZshBqQ5W8Q6K5LQ0Hwd0qCtSxPwIDAQABo1AwTjAdBgNVHQ4EFgQU8Y1j8vPz
aR3JMXdDrLK9LeV0RjwwHwYDVR0jBBgwFoAU8Y1j8vPzaR3JMXdDrLK9LeV0Rjww
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAeYr0R8wCLxq0ySl3EQ8G
bvF/VLgLCXlvVKEiKwrSkPZTzSmfOJcQfqCAH3pJjVqDOTAZ7H0cV8CzZVpK7q3U
VPl9D5p9hF0VJ3LcMhNLXNhp5C3WBBTqCXF5FQMxgNNwdlJW0cJLrXPG8D8yXhPc
M/qXYqd7K1Q4RiXBNLxSGEqPj5mZnVKZ7JQTKCYqF5uHVx3y8c7gK3nYaXNQfZFa
N8Qs9CZmKVFvJ4KU6nOaW5X8gTCrHvBFMFaQcKKpCmWZfLnPJZMdgZXvxhx5lXXU
9nKKqk7sKB4D6LqHKQ1qRx9HJJVP5LxHMYGpxGnxXNLMaCPjLMxVJpQSZLnVfj5Y
dQ==
-----END CERTIFICATE-----`
