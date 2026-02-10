package flags

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Options
		wantErr bool
	}{
		{
			name: "no args returns defaults",
			args: []string{},
			want: Options{},
		},
		{
			name: "all flags set",
			args: []string{
				"--lang", "go",
				"--framework", "gin",
				"--name", "myproject",
				"--dir", "/tmp/projects",
				"--dry-run",
				"--no-tui",
				"--config", "/etc/pi.yaml",
			},
			want: Options{
				Language:   "go",
				Framework:  "gin",
				Name:       "myproject",
				Dir:        "/tmp/projects",
				DryRun:     true,
				NoTUI:      true,
				ConfigPath: "/etc/pi.yaml",
			},
		},
		{
			name: "lang flag only",
			args: []string{"--lang", "rust"},
			want: Options{Language: "rust"},
		},
		{
			name: "framework flag only",
			args: []string{"--framework", "echo"},
			want: Options{Framework: "echo"},
		},
		{
			name: "name flag only",
			args: []string{"--name", "cool-app"},
			want: Options{Name: "cool-app"},
		},
		{
			name: "dir flag only",
			args: []string{"--dir", "/home/user/projects"},
			want: Options{Dir: "/home/user/projects"},
		},
		{
			name: "dry-run flag only",
			args: []string{"--dry-run"},
			want: Options{DryRun: true},
		},
		{
			name: "no-tui flag only",
			args: []string{"--no-tui"},
			want: Options{NoTUI: true},
		},
		{
			name: "config flag only",
			args: []string{"--config", "config.yaml"},
			want: Options{ConfigPath: "config.yaml"},
		},
		{
			name:    "invalid flag returns error",
			args:    []string{"--nonexistent", "value"},
			wantErr: true,
		},
		{
			name: "multiple flags combined",
			args: []string{"--lang", "python", "--name", "webapp", "--dry-run"},
			want: Options{
				Language: "python",
				Name:     "webapp",
				DryRun:   true,
			},
		},
		{
			name: "single-dash flags",
			args: []string{"-lang", "java", "-name", "api"},
			want: Options{
				Language: "java",
				Name:     "api",
			},
		},
		{
			name: "equals syntax",
			args: []string{"--lang=typescript", "--name=frontend"},
			want: Options{
				Language: "typescript",
				Name:     "frontend",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
