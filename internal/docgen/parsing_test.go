package docgen

import "testing"

func TestPrefixRuleIDWithPath(t *testing.T) {
	tests := []struct {
		testcase     string
		relativePath string
		unprefixedID string
		expected     string
	}{
		{
			testcase:     "example 1",
			relativePath: "apex/lang/best-practice/ncino/accessModifiers/GlobalAccessModifiers.yaml",
			unprefixedID: "global-access-modifiers",
			expected:     "apex.lang.best-practice.ncino.accessmodifiers.globalaccessmodifiers.global-access-modifiers",
		},
		{
			testcase:     "example 2",
			relativePath: "javascript/lang/best-practice/leftover_debugging.yaml",
			unprefixedID: "javascript-alert",
			expected:     "javascript.lang.best-practice.leftover_debugging.javascript-alert",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testcase, func(t *testing.T) {
			if got := prefixRuleIDWithPath(tt.relativePath, tt.unprefixedID); got != tt.expected {
				t.Errorf("prefixRuleIDWithPath() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestGetLastSegment(t *testing.T) {
	tests := []struct {
		testcase string
		input    string
		expected string
	}{
		{
			testcase: "single segment",
			input:    "insecure-use-string-copy-fn",
			expected: "insecure-use-string-copy-fn",
		},
		{
			testcase: "multiple segments",
			input:    "c.lang.security.insecure-use-string-copy-fn.insecure-use-string-copy-fn",
			expected: "insecure-use-string-copy-fn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testcase, func(t *testing.T) {
			if got := getLastSegment(tt.input); got != tt.expected {
				t.Errorf("getLastSegment() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestGetFirstSentence(t *testing.T) {
	tests := []struct {
		testcase string
		input    string
		expected string
	}{
		{
			testcase: "single sentence",
			input:    "Finding triggers whenever there is a strcpy or strncpy used.",
			expected: "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			testcase: "multiple sentences",
			input:    "Finding triggers whenever there is a strcpy or strncpy used. This is an issue because strcpy does not affirm the size of the destination array and strncpy will not automatically NULL-terminate strings. This can lead to buffer overflows, which can cause program crashes and potentially let an attacker inject code in the program. Fix this by using strcpy_s instead (although note that strcpy_s is an optional part of the C11 standard, and so may not be available).",
			expected: "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			testcase: "sentence with abbreviation or code",
			input:    "Finding triggers whenever there is a str.cpy or strncpy used.",
			expected: "Finding triggers whenever there is a str.cpy or strncpy used.",
		},
		{
			testcase: "sentence longer than 500 characters",
			input:    "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies.",
			expected: "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testcase, func(t *testing.T) {
			if got := getFirstSentence(tt.input); got != tt.expected {
				t.Errorf("getFirstSentence() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
