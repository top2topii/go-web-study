package main

import (
	"log/slog"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// map을 getenv 함수로 바꿔 주입한다. 실제 환경 변수를 건드리지 않는다.
	env := func(m map[string]string) func(string) string {
		return func(k string) string { return m[k] }
	}

	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    config
		wantErr bool
	}{
		{
			name: "defaults",
			want: config{addr: "127.0.0.1:8080", logLevel: slog.LevelInfo},
		},
		{
			name: "env overrides default",
			env:  map[string]string{"ADDR": ":9000", "LOG_LEVEL": "debug", "LOG_JSON": "1"},
			want: config{addr: ":9000", logLevel: slog.LevelDebug, logJSON: true},
		},
		{
			name: "flag overrides env",
			args: []string{"-addr", ":7000"},
			env:  map[string]string{"ADDR": ":9000"},
			want: config{addr: ":7000", logLevel: slog.LevelInfo},
		},
		{
			name:    "invalid log level",
			args:    []string{"-log-level", "loud"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.args, env(tt.env))
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
