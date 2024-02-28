package main

import (
	"testing"

	"github.com/lookvr/edge-tts-go/edgeTTS"
)

func Test_printVoices(t *testing.T) {
	type args struct {
		locale string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test-1",
			args: args{
				locale: "zh-",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edgeTTS.PrintVoices(tt.args.locale)
		})
	}
}
