package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/goreleaser/goreleaser/internal/artifact"
	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDescription(t *testing.T) {
	assert.NotEmpty(t, Pipe{}.String())
}

func TestRunPipe(t *testing.T) {
	folder, back := testlib.Mktmp(t)
	defer back()
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "darwinamd64"), 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "windowsamd64"), 0755))
	_, err := os.Create(filepath.Join(dist, "darwinamd64", "mybin"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(dist, "windowsamd64", "mybin.exe"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(folder, "README.md"))
	assert.NoError(t, err)
	assert.NoError(t, os.MkdirAll(filepath.Join(folder, "foo", "bar", "foobar"), 0755))
	_, err = os.Create(filepath.Join(filepath.Join(folder, "foo", "bar", "foobar", "blah.txt")))
	assert.NoError(t, err)
	for _, format := range []string{"tar.gz", "zip"} {
		t.Run("Archive format "+format, func(tt *testing.T) {
			var ctx = context.New(
				config.Project{
					Dist:        dist,
					ProjectName: "foobar",
					Archive: config.Archive{
						NameTemplate: defaultNameTemplate,
						Files: []string{
							"README.*",
							"./foo/**/*",
						},
						FormatOverrides: []config.FormatOverride{
							{
								Goos:   "windows",
								Format: "zip",
							},
						},
					},
				},
			)
			ctx.Artifacts.Add(artifact.Artifact{
				Goos:   "darwin",
				Goarch: "amd64",
				Name:   "mybin",
				Path:   filepath.Join(dist, "darwinamd64", "mybin"),
				Type:   artifact.Binary,
				Extra: map[string]string{
					"Binary": "mybin",
				},
			})
			ctx.Artifacts.Add(artifact.Artifact{
				Goos:   "windows",
				Goarch: "amd64",
				Name:   "mybin.exe",
				Path:   filepath.Join(dist, "windowsamd64", "mybin.exe"),
				Type:   artifact.Binary,
				Extra: map[string]string{
					"Binary":    "mybin",
					"Extension": ".exe",
				},
			})
			ctx.Version = "0.0.1"
			ctx.Git.CurrentTag = "v0.0.1"
			ctx.Config.Archive.Format = format
			assert.NoError(tt, Pipe{}.Run(ctx))
			var archives = ctx.Artifacts.Filter(artifact.ByType(artifact.UploadableArchive))
			require.Len(tt, archives.List(), 2)
			darwin := archives.Filter(artifact.ByGoos("darwin")).List()[0]
			windows := archives.Filter(artifact.ByGoos("windows")).List()[0]
			assert.Equal(tt, "foobar_0.0.1_darwin_amd64."+format, darwin.Name)
			assert.Equal(tt, "foobar_0.0.1_windows_amd64.zip", windows.Name)
		})
	}

	// Check archive contents
	assert.Equal(
		t,
		[]string{
			"README.md",
			"foo/bar",
			"foo/bar/foobar",
			"foo/bar/foobar/blah.txt",
			"mybin",
		},
		tarFiles(t, filepath.Join(dist, "foobar_0.0.1_darwin_amd64.tar.gz")),
	)
	assert.Equal(
		t,
		[]string{
			"README.md",
			"foo/bar/foobar/blah.txt",
			"mybin.exe",
		},
		zipFiles(t, filepath.Join(dist, "foobar_0.0.1_windows_amd64.zip")),
	)
}

func zipFiles(t *testing.T, path string) []string {
	f, err := os.Open(path)
	require.NoError(t, err)
	info, err := f.Stat()
	require.NoError(t, err)
	r, err := zip.NewReader(f, info.Size())
	require.NoError(t, err)
	var paths = make([]string, len(r.File))
	for i, zf := range r.File {
		paths[i] = zf.Name
	}
	return paths
}

func tarFiles(t *testing.T, path string) []string {
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()
	gr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gr.Close()
	var r = tar.NewReader(gr)
	var paths []string
	for {
		next, err := r.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		paths = append(paths, next.Name)
	}
	return paths
}

func TestRunPipeBinary(t *testing.T) {
	folder, back := testlib.Mktmp(t)
	defer back()
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "darwinamd64"), 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "windowsamd64"), 0755))
	_, err := os.Create(filepath.Join(dist, "darwinamd64", "mybin"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(dist, "windowsamd64", "mybin.exe"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(folder, "README.md"))
	assert.NoError(t, err)
	var ctx = context.New(
		config.Project{
			Dist: dist,
			Archive: config.Archive{
				Format:       "binary",
				NameTemplate: defaultBinaryNameTemplate,
			},
		},
	)
	ctx.Version = "0.0.1"
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.Artifacts.Add(artifact.Artifact{
		Goos:   "darwin",
		Goarch: "amd64",
		Name:   "mybin",
		Path:   filepath.Join(dist, "darwinamd64", "mybin"),
		Type:   artifact.Binary,
		Extra: map[string]string{
			"Binary": "mybin",
		},
	})
	ctx.Artifacts.Add(artifact.Artifact{
		Goos:   "windows",
		Goarch: "amd64",
		Name:   "mybin.exe",
		Path:   filepath.Join(dist, "windowsamd64", "mybin.exe"),
		Type:   artifact.Binary,
		Extra: map[string]string{
			"Binary": "mybin",
			"Ext":    ".exe",
		},
	})
	assert.NoError(t, Pipe{}.Run(ctx))
	var binaries = ctx.Artifacts.Filter(artifact.ByType(artifact.UploadableBinary))
	darwin := binaries.Filter(artifact.ByGoos("darwin")).List()[0]
	windows := binaries.Filter(artifact.ByGoos("windows")).List()[0]
	assert.Equal(t, "mybin_0.0.1_darwin_amd64", darwin.Name)
	assert.Equal(t, "mybin_0.0.1_windows_amd64.exe", windows.Name)
	assert.Len(t, binaries.List(), 2)
}

func TestRunPipeDistRemoved(t *testing.T) {
	var ctx = context.New(
		config.Project{
			Dist: "/path/nope",
			Archive: config.Archive{
				NameTemplate: "nope",
				Format:       "zip",
			},
		},
	)
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.Artifacts.Add(artifact.Artifact{
		Goos:   "windows",
		Goarch: "amd64",
		Name:   "mybin.exe",
		Path:   filepath.Join("/path/to/nope", "windowsamd64", "mybin.exe"),
		Type:   artifact.Binary,
		Extra: map[string]string{
			"Binary":    "mybin",
			"Extension": ".exe",
		},
	})
	assert.EqualError(t, Pipe{}.Run(ctx), `failed to create directory /path/nope/nope.zip: open /path/nope/nope.zip: no such file or directory`)
}

func TestRunPipeInvalidGlob(t *testing.T) {
	folder, back := testlib.Mktmp(t)
	defer back()
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "darwinamd64"), 0755))
	_, err := os.Create(filepath.Join(dist, "darwinamd64", "mybin"))
	assert.NoError(t, err)
	var ctx = context.New(
		config.Project{
			Dist: dist,
			Archive: config.Archive{
				NameTemplate: "foo",
				Format:       "zip",
				Files: []string{
					"[x-]",
				},
			},
		},
	)
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.Artifacts.Add(artifact.Artifact{
		Goos:   "darwin",
		Goarch: "amd64",
		Name:   "mybin",
		Path:   filepath.Join("dist", "darwinamd64", "mybin"),
		Type:   artifact.Binary,
		Extra: map[string]string{
			"Binary": "mybin",
		},
	})
	assert.EqualError(t, Pipe{}.Run(ctx), `failed to find files to archive: globbing failed for pattern [x-]: file does not exist`)
}

func TestRunPipeWrap(t *testing.T) {
	folder, back := testlib.Mktmp(t)
	defer back()
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "darwinamd64"), 0755))
	_, err := os.Create(filepath.Join(dist, "darwinamd64", "mybin"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(folder, "README.md"))
	assert.NoError(t, err)
	var ctx = context.New(
		config.Project{
			Dist: dist,
			Archive: config.Archive{
				NameTemplate:    "foo",
				WrapInDirectory: true,
				Format:          "tar.gz",
				Files: []string{
					"README.*",
				},
			},
		},
	)
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.Artifacts.Add(artifact.Artifact{
		Goos:   "darwin",
		Goarch: "amd64",
		Name:   "mybin",
		Path:   filepath.Join("dist", "darwinamd64", "mybin"),
		Type:   artifact.Binary,
		Extra: map[string]string{
			"Binary": "mybin",
		},
	})
	assert.NoError(t, Pipe{}.Run(ctx))

	// Check archive contents
	f, err := os.Open(filepath.Join(dist, "foo.tar.gz"))
	assert.NoError(t, err)
	defer func() { assert.NoError(t, f.Close()) }()
	gr, err := gzip.NewReader(f)
	assert.NoError(t, err)
	defer func() { assert.NoError(t, gr.Close()) }()
	r := tar.NewReader(gr)
	for _, n := range []string{"README.md", "mybin"} {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join("foo", n), h.Name)
	}
}

func TestDefault(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			Archive: config.Archive{},
		},
	}
	assert.NoError(t, Pipe{}.Default(ctx))
	assert.NotEmpty(t, ctx.Config.Archive.NameTemplate)
	assert.Equal(t, "tar.gz", ctx.Config.Archive.Format)
	assert.NotEmpty(t, ctx.Config.Archive.Files)
}

func TestDefaultSet(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			Archive: config.Archive{
				NameTemplate: "foo",
				Format:       "zip",
				Files: []string{
					"foo",
				},
			},
		},
	}
	assert.NoError(t, Pipe{}.Default(ctx))
	assert.Equal(t, "foo", ctx.Config.Archive.NameTemplate)
	assert.Equal(t, "zip", ctx.Config.Archive.Format)
	assert.Equal(t, "foo", ctx.Config.Archive.Files[0])
}

func TestDefaultFormatBinary(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			Archive: config.Archive{
				Format: "binary",
			},
		},
	}
	assert.NoError(t, Pipe{}.Default(ctx))
	assert.Equal(t, defaultBinaryNameTemplate, ctx.Config.Archive.NameTemplate)
}

func TestFormatFor(t *testing.T) {
	var ctx = &context.Context{
		Config: config.Project{
			Archive: config.Archive{
				Format: "tar.gz",
				FormatOverrides: []config.FormatOverride{
					{
						Goos:   "windows",
						Format: "zip",
					},
				},
			},
		},
	}
	assert.Equal(t, "zip", packageFormat(ctx, "windows"))
	assert.Equal(t, "tar.gz", packageFormat(ctx, "linux"))
}

func TestBinaryOverride(t *testing.T) {
	folder, back := testlib.Mktmp(t)
	defer back()
	var dist = filepath.Join(folder, "dist")
	assert.NoError(t, os.Mkdir(dist, 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "darwinamd64"), 0755))
	assert.NoError(t, os.Mkdir(filepath.Join(dist, "windowsamd64"), 0755))
	_, err := os.Create(filepath.Join(dist, "darwinamd64", "mybin"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(dist, "windowsamd64", "mybin.exe"))
	assert.NoError(t, err)
	_, err = os.Create(filepath.Join(folder, "README.md"))
	assert.NoError(t, err)
	for _, format := range []string{"tar.gz", "zip"} {
		t.Run("Archive format "+format, func(tt *testing.T) {
			var ctx = context.New(
				config.Project{
					Dist:        dist,
					ProjectName: "foobar",
					Archive: config.Archive{
						NameTemplate: defaultNameTemplate,
						Files: []string{
							"README.*",
						},
						FormatOverrides: []config.FormatOverride{
							{
								Goos:   "windows",
								Format: "binary",
							},
						},
					},
				},
			)
			ctx.Git.CurrentTag = "v0.0.1"
			ctx.Artifacts.Add(artifact.Artifact{
				Goos:   "darwin",
				Goarch: "amd64",
				Name:   "mybin",
				Path:   filepath.Join(dist, "darwinamd64", "mybin"),
				Type:   artifact.Binary,
				Extra: map[string]string{
					"Binary": "mybin",
				},
			})
			ctx.Artifacts.Add(artifact.Artifact{
				Goos:   "windows",
				Goarch: "amd64",
				Name:   "mybin.exe",
				Path:   filepath.Join(dist, "windowsamd64", "mybin.exe"),
				Type:   artifact.Binary,
				Extra: map[string]string{
					"Binary": "mybin",
					"Ext":    ".exe",
				},
			})
			ctx.Version = "0.0.1"
			ctx.Config.Archive.Format = format

			assert.NoError(tt, Pipe{}.Run(ctx))
			var archives = ctx.Artifacts.Filter(artifact.ByType(artifact.UploadableArchive))
			darwin := archives.Filter(artifact.ByGoos("darwin")).List()[0]
			assert.Equal(tt, "foobar_0.0.1_darwin_amd64."+format, darwin.Name)

			archives = ctx.Artifacts.Filter(artifact.ByType(artifact.UploadableBinary))
			windows := archives.Filter(artifact.ByGoos("windows")).List()[0]
			assert.Equal(tt, "foobar_0.0.1_windows_amd64.exe", windows.Name)

		})
	}
}
