package main

import (
	"github.com/rendon/testcli"
	"strings"
	"testing"
)

func TestBuild(t *testing.T) {
	tests := []struct {
		args     []string
		stdin    string
		stdErr   []string
		woStdErr []string
		stdOut   []string
		wantErr  bool
	}{
		{
			args: []string{"build", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
			},
		},
		{
			args: []string{"build", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
			},
			stdOut: []string{"foo-goo"},
		},
		{
			args:  []string{"build", "from", "stdin"},
			stdin: "foo goo",
			stdErr: []string{
				"queueing read from reader",
				"queueing build",
			},
			stdOut: []string{"foo-goo"},
		},
		{
			args: []string{"find", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing find",
				"fetching tags",
				"gofunky/docker",
			},
			wantErr: true,
		},
		{
			args: []string{"find", "in", "gofunky/git", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing find",
				"fetching tags",
				"gofunky/git",
			},
			wantErr: true,
		},
		{
			args:    []string{"version"},
			stdErr:  []string{"version"},
			wantErr: true,
		},
		{
			args:    []string{"help"},
			stdErr:  []string{"help"},
			wantErr: true,
		},
		{
			args: []string{"tag", "source", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing tagging",
				"tagged",
				"docker tag source gofunky/docker:master",
			},
		},
		{
			args: []string{"tag", "source", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing tagging",
				"tagged",
				"docker tag source foo-goo",
			},
		},
		{
			args: []string{"tag", "source", "to", "gofunky/repo", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing tagging",
				"tagged",
				"docker tag source gofunky/repo:master",
			},
		},
		{
			args: []string{"tag", "source", "to", "gofunky/repo", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
				"queueing tagging",
				"tagged",
				"docker tag source gofunky/repo:foo-goo",
			},
		},
		{
			args: []string{"push", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing push",
				"docker push gofunky/docker:master",
			},
			woStdErr: []string{
				"tagged",
				"docker tag",
			},
		},
		{
			args: []string{"push", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
				"queueing push",
				"docker push foo-goo",
			},
			woStdErr: []string{
				"tagged",
				"docker tag",
				"docker push :foo-goo",
			},
		},
		{
			args: []string{"push", "to", "gofunky/git", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing push",
				"docker push gofunky/git:master",
			},
			woStdErr: []string{
				"tagged",
				"docker tag",
			},
		},
		{
			args: []string{"push", "to", "gofunky/git", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
				"queueing push",
				"docker push gofunky/git:foo-goo",
			},
			woStdErr: []string{
				"tagged",
				"docker tag",
				"docker push foo-goo",
			},
		},
		{
			args: []string{"push", "source", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing tagging",
				"queueing push",
				"docker tag source gofunky/docker:master",
				"docker push gofunky/docker:master",
			},
		},
		{
			args: []string{"push", "source", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
				"queueing tagging",
				"queueing push",
				"docker tag source foo-goo",
				"docker push foo-goo",
			},
			woStdErr: []string{
				"docker push :foo-goo",
			},
		},
		{
			args: []string{"push", "source", "to", "gofunky/git", "from", "file", "../../test/Dockerfile"},
			stdErr: []string{
				"queueing read from Dockerfile",
				"queueing build",
				"queueing tagging",
				"queueing push",
				"docker tag source gofunky/git:master",
				"docker push gofunky/git:master",
			},
		},
		{
			args: []string{"push", "source", "to", "gofunky/git", "from", "foo", "goo"},
			stdErr: []string{
				"queueing read from slice",
				"queueing build",
				"queueing tagging",
				"queueing push",
				"docker tag source gofunky/git:foo-goo",
				"docker push gofunky/git:foo-goo",
			},
			woStdErr: []string{
				"docker push :foo-goo",
			},
		},
	}
	for _, tt := range tests {
		command := strings.Join(tt.args, " ")
		t.Run(command, func(t *testing.T) {
			cliArgs := append(tt.args, "--verbose", "--simulate")
			c := testcli.Command("tuplip", cliArgs...)
			if tt.stdin != "" {
				c.SetStdin(strings.NewReader(tt.stdin))
			}
			t.Logf("testing command: %s", cliArgs)
			c.Run()
			if c.Success() == tt.wantErr {
				t.Errorf("tuplip error = %v, wantErr %v, error message:\n%v", c.Success(), tt.wantErr, c.Error())
			}
			for _, want := range tt.stdErr {
				if !c.StderrContains(want) {
					t.Fatalf("Expected to contain %q in stderr %v", want, c.Stderr())
				}
			}
			for _, dontWant := range tt.woStdErr {
				if c.StderrContains(dontWant) {
					t.Fatalf("Expected not to contain %q in stderr %v", dontWant, c.Stderr())
				}
			}
			for _, want := range tt.stdOut {
				if !c.StdoutContains(want) {
					t.Fatalf("Expected to contain %q in stdout %v", want, c.Stdout())
				}
			}
		})
	}
}
