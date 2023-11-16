package docgen

import "testing"

func TestPrefixRuleIDWithPath(t *testing.T) {
	tests := []struct {
		test_name    string
		relativePath string
		unprefixedID string
		expected     string
	}{
		{
			test_name:    "example 1",
			relativePath: "apex/lang/best-practice/ncino/accessModifiers/GlobalAccessModifiers.yaml",
			unprefixedID: "global-access-modifiers",
			expected:     "apex.lang.best-practice.ncino.accessmodifiers.globalaccessmodifiers.global-access-modifiers",
		},
		{
			test_name:    "example 2",
			relativePath: "javascript/lang/best-practice/leftover_debugging.yaml",
			unprefixedID: "javascript-alert",
			expected:     "javascript.lang.best-practice.leftover_debugging.javascript-alert",
		},
	}
	for _, test := range tests {
		t.Run(test.test_name, func(t *testing.T) {
			if got := prefixRuleIDWithPath(test.relativePath, test.unprefixedID); got != test.expected {
				t.Errorf("prefixRuleIDWithPath() = %v, expected %v", got, test.expected)
			}
		})
	}
}

func TestGetLastSegment(t *testing.T) {
	tests := []struct {
		test_name string
		input     string
		expected  string
	}{
		{
			test_name: "single segment",
			input:     "insecure-use-string-copy-fn",
			expected:  "insecure-use-string-copy-fn",
		},
		{
			test_name: "multiple segments",
			input:     "c.lang.security.insecure-use-string-copy-fn.insecure-use-string-copy-fn",
			expected:  "insecure-use-string-copy-fn",
		},
	}
	for _, test := range tests {
		t.Run(test.test_name, func(t *testing.T) {
			if got := getLastSegment(test.input); got != test.expected {
				t.Errorf("getLastSegment() = %v, expected %v", got, test.expected)
			}
		})
	}
}

func TestGetFirstSentence(t *testing.T) {
	tests := []struct {
		test_name string
		input     string
		expected  string
	}{
		{
			test_name: "single sentence",
			input:     "Finding triggers whenever there is a strcpy or strncpy used.",
			expected:  "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			test_name: "multiple sentences",
			input:     "Finding triggers whenever there is a strcpy or strncpy used. This is an issue because strcpy does not affirm the size of the destination array and strncpy will not automatically NULL-terminate strings. This can lead to buffer overflows, which can cause program crashes and potentially let an attacker inject code in the program. Fix this by using strcpy_s instead (although note that strcpy_s is an optional part of the C11 standard, and so may not be available).",
			expected:  "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			test_name: "sentence with abbreviation or code",
			input:     "Finding triggers whenever there is a str.cpy or strncpy used.",
			expected:  "Finding triggers whenever there is a str.cpy or strncpy used.",
		},
		{
			test_name: "sentence longer than 500 characters",
			input:     "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies.",
			expected:  "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candi",
		},
	}
	for _, test := range tests {
		t.Run(test.test_name, func(t *testing.T) {
			if got := GetFirstSentence(test.input); got != test.expected {
				t.Errorf("getFirstSentence() = %v, expected %v", got, test.expected)
			}
		})
	}
}
