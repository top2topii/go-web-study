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
	def := config{addr: "127.0.0.1:8080", dbPath: "data/shopadmin.db", logLevel: slog.LevelInfo, cookieSecure: true}

	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    func(c config) config // 기본값에서 달라지는 부분만 적는다
		wantErr bool
	}{
		{
			name: "defaults",
			want: func(c config) config { return c },
		},
		{
			name: "env overrides default",
			env:  map[string]string{"ADDR": ":9000", "DB_PATH": "/tmp/x.db", "LOG_LEVEL": "debug", "LOG_JSON": "1", "COOKIE_SECURE": "0"},
			want: func(c config) config {
				c.addr, c.dbPath, c.logLevel, c.logJSON, c.cookieSecure = ":9000", "/tmp/x.db", slog.LevelDebug, true, false
				return c
			},
		},
		{
			name: "flag overrides env",
			args: []string{"-addr", ":7000", "-cookie-secure=false"},
			env:  map[string]string{"ADDR": ":9000"},
			want: func(c config) config {
				c.addr, c.cookieSecure = ":7000", false
				return c
			},
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
			if tt.wantErr {
				return
			}
			if want := tt.want(def); got != want {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}
