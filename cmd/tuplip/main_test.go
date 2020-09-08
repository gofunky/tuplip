package main

import (
	"github.com/rendon/testcli"
	"strings"
	"testing"
)

const WithoutRepository = "../../test/WithoutRepository.Dockerfile"
const WithRepository = "../../test/WithRepository.Dockerfile"

func TestBuild(t *testing.T) {
	type testBuild struct {
		args     []string
		stdin    string
		stdErr   map[string]bool
		stdOut   map[string]bool
		wantErr  bool
		criteria string
		replace  bool
	}
	matrix := []testBuild{
		{},
		{
			criteria: "source from foo goo",
			args:     []string{"-s"},
			stdErr: map[string]bool{
				"queueing build":           false,
				"foo-goo":                  false,
				"straight channel enabled": true,
			},
		},
		{
			criteria: "build",
			args:     []string{"--filter=foo"},
			stdOut: map[string]bool{
				"goo":                false,
				"foo":                true,
				"foo-goo":            true,
				"2.4":                false,
				"gofunky/ignore:2.4": false,
				"gofunky/git:2.4":    false,
				"latest":             false,
			},
		},
		{
			criteria: WithRepository,
			args:     []string{"--filter=golang,foo"},
			stdOut: map[string]bool{
				"2.4":                                  false,
				"gofunky/git:2.4":                      false,
				"gofunky/ignore:2.4":                   false,
				"docker18.9.0-foo-golang":              false,
				"gofunky/git:docker18.9.0-foo-golang1": true,
				"foo":                                  false,
				"gofunky/git:foo":                      false,
				"golang":                               false,
				"gofunky/git:golang":                   false,
			},
		},
		{
			criteria: "file",
			args:     []string{"-r 6.3.8"},
			stdOut: map[string]bool{
				"2.4":                  false,
				"6.3.8":                true,
				"gofunky/ignore:2.4":   false,
				"gofunky/ignore:6.3.8": true,
				"gofunky/git:2.4":      false,
				"gofunky/git:6.3.8":    true,
				"latest":               false,
			},
			stdErr: map[string]bool{
				"exclusive latest tag was found": false,
			},
		},
		{
			criteria: WithoutRepository,
			args:     []string{WithRepository},
			stdOut: map[string]bool{
				"foo":                      false,
				"2.4":                      false,
				"latest":                   false,
				"golang":                   false,
				"docker18.9.0-foo-golang1": false,
				"gofunky/ignore:latest":    true,
				"gofunky/ignore:2.4":       true,
				"gofunky/ignore:foo":       true,
			},
			replace: true,
		},
		{
			criteria: "build",
			args:     []string{"build", "to", "gofunky/ignore"},
			stdOut: map[string]bool{
				"latest":                false,
				"2.4":                   false,
				"goo":                   false,
				"foo":                   false,
				"foo-goo":               false,
				"gofunky/ignore:foo":    true,
				"gofunky/ignore:2.4":    true,
				"gofunky/ignore:latest": true,
			},
			replace: true,
		},
	}
	tests := []testBuild{
		{
			args: []string{"build", "from", "file", WithoutRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile": true,
				"queueing build":                true,
			},
			stdOut: map[string]bool{
				"foo":                true,
				"2.4":                true,
				"6.3.8":              false,
				"gofunky/ignore:foo": false,
				"gofunky/ignore:2.4": false,
			},
		},
		{
			args: []string{"build", "from", "file", WithoutRepository, "--root-version=latest", "-e"},
			stdErr: map[string]bool{
				"queueing read from Dockerfile":  true,
				"queueing build":                 true,
				"exclusive latest tag was found": true,
			},
			stdOut: map[string]bool{
				"2.4":                   false,
				"latest":                true,
				"gofunky/ignore:latest": false,
			},
		},
		{
			args: []string{"build", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice": true,
				"queueing build":           true,
			},
			stdOut: map[string]bool{
				"goo":                true,
				"foo":                true,
				"foo-goo":            true,
				"gofunky/ignore:foo": false,
			},
		},
		{
			args:  []string{"build", "from", "stdin"},
			stdin: "foo goo",
			stdErr: map[string]bool{
				"queueing read from reader": true,
				"queueing build":            true,
			},
			stdOut: map[string]bool{
				"foo-goo":            true,
				"foo":                true,
				"goo":                true,
				"gofunky/ignore:foo": false,
			},
		},
		{
			args: []string{"find", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile": true,
				"queueing find":                 true,
				"fetching tags":                 true,
				"gofunky/ignore":                true,
			},
			wantErr: true,
		},
		{
			args: []string{"find", "in", "gofunky/git", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice": true,
				"queueing find":            true,
				"fetching tags":            true,
				"gofunky/git":              true,
			},
			wantErr: true,
		},
		{
			args:    []string{"version"},
			stdErr:  map[string]bool{"version": true},
			wantErr: true,
		},
		{
			args:    []string{"help"},
			stdErr:  map[string]bool{"help": true},
			wantErr: true,
		},
		{
			args: []string{"tag", "source", "from", "file", WithoutRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile": true,
				"queueing build":                true,
				"queueing tagging":              true,
				"tagged":                        true,
				"straight channel enabled":      false,
			},
			stdOut: map[string]bool{
				"golang":                   true,
				"foo":                      true,
				"docker18.9.0-foo-golang1": true,
				"2.4":                      true,
				"gofunky/ignore:foo":       false,
				"6.3.8":                    false,
			},
		},
		{
			args: []string{"tag", "source", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice":  true,
				"queueing build":            true,
				"queueing tagging":          true,
				"tagged":                    true,
				"docker tag source foo-goo": true,
			},
		},
		{
			args: []string{"tag", "source", "to", "gofunky/git", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile":     true,
				"queueing build":                    true,
				"queueing tagging":                  true,
				"tagged":                            true,
				"docker tag source gofunky/git:foo": true,
				"straight channel enabled":          false,
			},
			stdOut: map[string]bool{
				"gofunky/git:2.4":                     true,
				"gofunky/git:golang":                  true,
				"gofunky/git:foo":                     true,
				"gofunky/git:docker18.9.0-foo-golang": true,
				"2.4":                                 false,
				"gofunky/git:6.3.8":                   false,
			},
		},
		{
			args: []string{"tag", "source", "to", "gofunky/repo", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice":               true,
				"queueing build":                         true,
				"queueing tagging":                       true,
				"tagged":                                 true,
				"docker tag source gofunky/repo:foo-goo": true,
				"docker tag source foo-goo":              false,
			},
		},
		{
			args: []string{"push", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile":  true,
				"queueing build":                 true,
				"queueing push":                  true,
				"docker push gofunky/ignore:foo": true,
				"tagged":                         false,
				"docker tag":                     false,
			},
			stdOut: map[string]bool{
				"gofunky/ignore:2.4":   true,
				"gofunky/ignore:6.3.8": false,
			},
		},
		{
			args: []string{"push", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice": true,
				"queueing build":           true,
				"queueing push":            true,
				"docker push foo-goo":      true,
				"tagged":                   false,
				"docker tag":               false,
				"docker push :foo-goo":     false,
			},
		},
		{
			args: []string{"push", "to", "gofunky/git", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile": true,
				"queueing build":                true,
				"queueing push":                 true,
				"docker push gofunky/git:foo":   true,
				"tagged":                        false,
				"docker tag":                    false,
			},
		},
		{
			args: []string{"push", "to", "gofunky/git", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice":        true,
				"queueing build":                  true,
				"queueing push":                   true,
				"docker push gofunky/git:foo-goo": true,
				"tagged":                          false,
				"docker tag":                      false,
				"docker push foo-goo":             false,
			},
		},
		{
			args: []string{"push", "source", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile":        true,
				"queueing build":                       true,
				"queueing tagging":                     true,
				"queueing push":                        true,
				"docker tag source gofunky/ignore:foo": true,
				"docker push gofunky/ignore:foo":       true,
			},
		},
		{
			args: []string{"push", "source", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice":  true,
				"queueing build":            true,
				"queueing tagging":          true,
				"queueing push":             true,
				"docker tag source foo-goo": true,
				"docker push foo-goo":       true,
				"docker push :foo-goo":      false,
			},
		},
		{
			args: []string{"push", "source", "to", "gofunky/git", "from", "file", WithRepository},
			stdErr: map[string]bool{
				"queueing read from Dockerfile":     true,
				"queueing build":                    true,
				"queueing tagging":                  true,
				"queueing push":                     true,
				"docker tag source gofunky/git:foo": true,
				"docker push gofunky/git:foo":       true,
			},
		},
		{
			args: []string{"push", "source", "to", "gofunky/git", "from", "foo", "goo"},
			stdErr: map[string]bool{
				"queueing read from slice":              true,
				"queueing build":                        true,
				"queueing tagging":                      true,
				"queueing push":                         true,
				"docker tag source gofunky/git:foo-goo": true,
				"docker push gofunky/git:foo-goo":       true,
				"docker push :foo-goo":                  false,
			},
		},
	}
	for _, rawTT := range tests {
		for _, mod := range matrix {
			rawCommand := strings.Join(rawTT.args, " ")
			if strings.Contains(rawCommand, mod.criteria) {
				var fullArgs []string
				if mod.replace {
					var i = 0
					for _, arg := range rawTT.args {
						if arg == mod.criteria {
							for _, replaceArg := range mod.args {
								fullArgs = append(fullArgs, replaceArg)
								i++
							}
						} else {
							fullArgs = append(fullArgs, arg)
							i++
						}
					}
				} else {
					fullArgs = append(rawTT.args, mod.args...)
				}
				tt := &testBuild{
					args:    fullArgs,
					wantErr: rawTT.wantErr || mod.wantErr,
					stdin:   rawTT.stdin + mod.stdin,
					stdOut:  make(map[string]bool, len(rawTT.stdOut)),
					stdErr:  make(map[string]bool, len(rawTT.stdErr)),
				}
				for rawK, rawV := range rawTT.stdOut {
					tt.stdOut[rawK] = rawV
					for modK, modV := range mod.stdOut {
						if rawK == modK {
							tt.stdOut[rawK] = modV
						}
					}
				}
				for rawK, rawV := range rawTT.stdErr {
					tt.stdErr[rawK] = rawV
					for modK, modV := range mod.stdErr {
						if strings.Contains(rawK, modK) {
							tt.stdErr[rawK] = modV
						}
					}
				}
				command := strings.Join(tt.args, " ")
				t.Run(command, func(t *testing.T) {
					cliArgs := append(tt.args, "--verbose", "--simulate")
					c := testcli.Command("tuplip", cliArgs...)
					if tt.stdin != "" {
						c.SetStdin(strings.NewReader(tt.stdin))
					}
					c.Run()
					if c.Success() == tt.wantErr {
						t.Errorf("tuplip error = %v, wantErr %v, error message:\n%v", c.Success(), tt.wantErr, c.Error())
					}
					for key, want := range tt.stdErr {
						if c.StderrContains(key) != want {
							t.Fatalf("Expected %q = %v in stderr:\n%v", key, want, c.Stderr())
						}
					}
					for key, want := range tt.stdOut {
						stdOut := strings.Split(c.Stdout(), "\n")
						var found bool
						for _, stdOutLine := range stdOut {
							if stdOutLine == key {
								found = true
							}
						}
						if found != want {
							t.Fatalf("Expected %q = %v in stdout:\n%v", key, want, c.Stdout())
						}
					}
				})
			}
		}
	}
}
