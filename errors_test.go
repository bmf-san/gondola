package gondola

import "testing"

func TestConfigLoadError(t *testing.T) {
	err := &ConfigLoadError{}
	if err.Error() != "error loading config: <nil>" {
		t.Errorf("Expected error loading config: <nil>, got %v", err.Error())
	}
	if err.Unwrap() != nil {
		t.Errorf("Expected nil, got %v", err.Unwrap())
	}
}

func TestProxyServerError(t *testing.T) {
	err := &ProxyServerError{}
	if err.Error() != "error creating server: <nil>" {
		t.Errorf("Expected error creating server: <nil>, got %v", err.Error())
	}
	if err.Unwrap() != nil {
		t.Errorf("Expected nil, got %v", err.Unwrap())
	}
}
