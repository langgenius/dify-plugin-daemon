package bundle_entities

import (
	"testing"
)

func TestGithubDependencyPatternRegex(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid patterns
		{
			name:     "basic version pattern",
			input:    "owner/repo/1.0.0/manifest.json",
			expected: true,
		},
		{
			name:     "version with patch",
			input:    "owner/repo/1.0.1/manifest.yaml",
			expected: true,
		},
		{
			name:     "version with pre-release",
			input:    "owner/repo/1.0.0-beta/manifest.json",
			expected: true,
		},
		{
			name:     "version with x pattern",
			input:    "owner/repo/1.x.x/manifest.json",
			expected: true,
		},
		{
			name:     "version with X pattern",
			input:    "owner/repo/1.X.X/manifest.json",
			expected: true,
		},
		{
			name:     "version with mixed x pattern",
			input:    "owner/repo/1.2.x/manifest.json",
			expected: true,
		},
		{
			name:     "version with tilde",
			input:    "owner/repo/~1.0.0/manifest.json",
			expected: true,
		},
		{
			name:     "version with caret",
			input:    "owner/repo/^1.0.0/manifest.json",
			expected: true,
		},
		{
			name:     "version range",
			input:    "owner/repo/1.0.0-2.0.0/manifest.json",
			expected: true,
		},
		{
			name:     "complex owner and repo names",
			input:    "complex-owner/complex-repo-name/1.0.0/manifest.json",
			expected: true,
		},
		{
			name:     "underscore in names",
			input:    "owner_name/repo_name/1.0.0/manifest.json",
			expected: true,
		},

		// Invalid patterns
		{
			name:     "four digit version",
			input:    "owner/repo/1.0.0.1/manifest.json",
			expected: false,
		},
		{
			name:     "empty owner",
			input:    "/repo/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "empty repo",
			input:    "owner//1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "invalid version format",
			input:    "owner/repo/1.0/manifest.json",
			expected: false,
		},
		{
			name:     "missing manifest file",
			input:    "owner/repo/1.0.0/",
			expected: false,
		},
		{
			name:     "uppercase in owner",
			input:    "Owner/repo/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "uppercase in repo",
			input:    "owner/Repo/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "invalid characters in owner",
			input:    "owner@/repo/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "invalid characters in repo",
			input:    "owner/repo#/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "too long owner name",
			input:    "ownerwithaverylongnamethatshouldnotbeallowedinthiscaseownerwithaverylongnamethatshouldnotbeallowedinthiscase/repo/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "too long repo name",
			input:    "owner/repowithavrepowithaverylongnamethatshouldnotbeallowedinthiscaseandshouldbeshorterthanspecifiedintherequirementsrepowithaverylongnamethatshouldnotbeallowedinthiscaseandshouldbeshorterthanspecifiedintherequirementserylongnamethatshouldnotbeallowedinthiscaseandshouldbeshorterthanspecifiedintherequirements/1.0.0/manifest.json",
			expected: false,
		},
		{
			name:     "invalid version range format",
			input:    "owner/repo/1.0.0-/manifest.json",
			expected: false,
		},
		{
			name:     "invalid pre-release format",
			input:    "owner/repo/1.0.0-toolongprerelease/manifest.json",
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result := GITHUB_DEPENDENCY_PATTERN_REGEX_COMPILED.MatchString(testCase.input)
			if result != testCase.expected {
				t.Errorf("Test case '%s' failed: input '%s' expected %v but got %v, pattern: %s",
					testCase.name, testCase.input, testCase.expected, result, GITHUB_DEPENDENCY_PATTERN_REGEX_COMPILED.String())
			}
		})
	}
}
