package service

import (
	"testing"

	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
)

func TestIsUnauthorizedLanggenius(t *testing.T) {
	// Tests for isUnauthorizedLanggenius function
	// This function is used when ENFORCE_LANGGENIUS_PLUGIN_SIGNATURES=true (default)
	// to prevent unauthorized plugins from impersonating Langgenius
	tests := []struct {
		name         string
		author       string
		verification *decoder.Verification
		want         bool
	}{
		{
			name:   "langgenius author with proper verification",
			author: "langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // properly authorized
		},
		{
			name:   "langgenius author with partner verification",
			author: "langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_PARTNER,
			},
			want: true, // unauthorized - claims langgenius but verified as partner
		},
		{
			name:   "langgenius author with community verification",
			author: "langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: true, // unauthorized - claims langgenius but verified as community
		},
		{
			name:         "langgenius author without verification",
			author:       "langgenius",
			verification: nil,
			want:         true, // unauthorized - claims langgenius but no verification
		},
		{
			name:   "Langgenius author (capital L) with proper verification",
			author: "Langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // properly authorized (case-insensitive)
		},
		{
			name:   "LANGGENIUS author (all caps) with proper verification",
			author: "LANGGENIUS",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // properly authorized (case-insensitive)
		},
		{
			name:         "LANGGENIUS author (all caps) without verification",
			author:       "LANGGENIUS",
			verification: nil,
			want:         true, // unauthorized - claims langgenius but no verification
		},
		{
			name:   "community author with community verification",
			author: "community_developer",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: false, // authorized - doesn't claim langgenius
		},
		{
			name:   "partner author with partner verification",
			author: "partner_company",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_PARTNER,
			},
			want: false, // authorized - doesn't claim langgenius
		},
		{
			name:         "community author without verification",
			author:       "john_doe",
			verification: nil,
			want:         false, // allowed - doesn't claim langgenius
		},
		{
			name:   "empty author with langgenius verification",
			author: "",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // allowed - doesn't claim langgenius
		},
		{
			name:         "empty author without verification",
			author:       "",
			verification: nil,
			want:         false, // allowed - doesn't claim langgenius
		},
		{
			name:   "author contains langgenius but not exact match",
			author: "not_langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: false, // allowed - not exact match
		},
		{
			name:   "author langgenius_team",
			author: "langgenius_team",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: false, // allowed - not exact match
		},
		{
			name:   "author my_langgenius",
			author: "my_langgenius",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: false, // allowed - not exact match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			declaration := &plugin_entities.PluginDeclaration{
				PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
					Author: tt.author,
				},
			}
			
			got := isUnauthorizedLanggenius(declaration, tt.verification)
			if got != tt.want {
				t.Errorf("isUnauthorizedLanggenius() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsUnauthorizedLanggenius_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		author       string
		verification *decoder.Verification
		want         bool
	}{
		{
			name:   "langgenius with spaces",
			author: " langgenius ",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // spaces don't affect the comparison after lowercase
		},
		{
			name:   "langgenius with spaces but no verification",
			author: " langgenius ",
			verification: nil,
			want: false, // with spaces, not exact match after lowercase
		},
		{
			name:   "LaNgGeNiUs mixed case",
			author: "LaNgGeNiUs",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_LANGGENIUS,
			},
			want: false, // properly authorized (case-insensitive)
		},
		{
			name:   "langgenius. with punctuation",
			author: "langgenius.",
			verification: &decoder.Verification{
				AuthorizedCategory: decoder.AUTHORIZED_CATEGORY_COMMUNITY,
			},
			want: false, // not exact match due to punctuation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			declaration := &plugin_entities.PluginDeclaration{
				PluginDeclarationWithoutAdvancedFields: plugin_entities.PluginDeclarationWithoutAdvancedFields{
					Author: tt.author,
				},
			}
			
			got := isUnauthorizedLanggenius(declaration, tt.verification)
			if got != tt.want {
				t.Errorf("isUnauthorizedLanggenius() = %v, want %v for author=%q", got, tt.want, tt.author)
			}
		})
	}
}