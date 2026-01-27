package yak

import "testing"

func TestValidateName_RejectsBackslash(t *testing.T) {
	err := ValidateName("foo\\bar")
	if err == nil {
		t.Error("expected error for backslash, got nil")
	}
}

func TestValidateName_RejectsColon(t *testing.T) {
	err := ValidateName("foo:bar")
	if err == nil {
		t.Error("expected error for colon, got nil")
	}
}

func TestValidateName_RejectsAsterisk(t *testing.T) {
	err := ValidateName("foo*bar")
	if err == nil {
		t.Error("expected error for asterisk, got nil")
	}
}

func TestValidateName_RejectsQuestionMark(t *testing.T) {
	err := ValidateName("foo?bar")
	if err == nil {
		t.Error("expected error for question mark, got nil")
	}
}

func TestValidateName_RejectsPipe(t *testing.T) {
	err := ValidateName("foo|bar")
	if err == nil {
		t.Error("expected error for pipe, got nil")
	}
}

func TestValidateName_RejectsLessThan(t *testing.T) {
	err := ValidateName("foo<bar")
	if err == nil {
		t.Error("expected error for less than, got nil")
	}
}

func TestValidateName_RejectsGreaterThan(t *testing.T) {
	err := ValidateName("foo>bar")
	if err == nil {
		t.Error("expected error for greater than, got nil")
	}
}

func TestValidateName_RejectsQuote(t *testing.T) {
	err := ValidateName("foo\"bar")
	if err == nil {
		t.Error("expected error for quote, got nil")
	}
}

func TestValidateName_AcceptsForwardSlash(t *testing.T) {
	err := ValidateName("parent/child")
	if err != nil {
		t.Errorf("expected no error for forward slash, got %v", err)
	}
}

func TestValidateName_AcceptsValidName(t *testing.T) {
	err := ValidateName("my yak task")
	if err != nil {
		t.Errorf("expected no error for valid name, got %v", err)
	}
}
