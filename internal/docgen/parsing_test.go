package docgen

import "testing"

func TestPrefixRuleIDWithPath(t *testing.T) {
	tests := []struct {
		name         string
		relativePath string
		unprefixedID string
		want         string
	}{
		{
			name:         "example 1",
			relativePath: "apex/lang/best-practice/ncino/accessModifiers/GlobalAccessModifiers.yaml",
			unprefixedID: "global-access-modifiers",
			want:         "apex.lang.best-practice.ncino.accessmodifiers.globalaccessmodifiers.global-access-modifiers",
		},
		{
			name:         "example 2",
			relativePath: "javascript/lang/best-practice/leftover_debugging.yaml",
			unprefixedID: "javascript-alert",
			want:         "javascript.lang.best-practice.leftover_debugging.javascript-alert",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prefixRuleIDWithPath(tt.relativePath, tt.unprefixedID); got != tt.want {
				t.Errorf("prefixRuleIDWithPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLastSegment(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "single segment",
			s:    "insecure-use-string-copy-fn",
			want: "insecure-use-string-copy-fn",
		},
		{
			name: "multiple segments",
			s:    "c.lang.security.insecure-use-string-copy-fn.insecure-use-string-copy-fn",
			want: "insecure-use-string-copy-fn",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLastSegment(tt.s); got != tt.want {
				t.Errorf("getLastSegment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetFirstSentence(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want string
	}{
		{
			name: "single sentence",
			s:    "Finding triggers whenever there is a strcpy or strncpy used.",
			want: "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			name: "multiple sentences",
			s:    "Finding triggers whenever there is a strcpy or strncpy used. This is an issue because strcpy does not affirm the size of the destination array and strncpy will not automatically NULL-terminate strings. This can lead to buffer overflows, which can cause program crashes and potentially let an attacker inject code in the program. Fix this by using strcpy_s instead (although note that strcpy_s is an optional part of the C11 standard, and so may not be available).",
			want: "Finding triggers whenever there is a strcpy or strncpy used.",
		},
		{
			name: "sentence with abbreviation or code",
			s:    "Finding triggers whenever there is a str.cpy or strncpy used.",
			want: "Finding triggers whenever there is a str.cpy or strncpy used.",
		},
		{
			name: "sentence longer than 500 characters",
			s:    "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies.",
			want: "Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candies are more tasty than orange candies, Blue candi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFirstSentence(tt.s); got != tt.want {
				t.Errorf("getFirstSentence() = %v, want %v", got, tt.want)
			}
		})
	}
}
