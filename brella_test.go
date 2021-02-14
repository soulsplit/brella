package main

import (
	"reflect"
	"testing"

	goex "github.com/soulsplit/goex"
)

func Test_getCredentials(t *testing.T) {
	tests := []struct {
		name string
		want credentials
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getCredentials(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getAPIHandle(t *testing.T) {
	tests := []struct {
		name string
		want goex.API
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getAPIHandle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getAPIHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}
