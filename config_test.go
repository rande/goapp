package goapp

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

var filecontent = `
type    = "master"
dsn     = "{{ env "PG_USER" }}:{{ env "PG_PASSWORD" }}"
`

var expected = `
type    = "master"
dsn     = "foo:bar"
`

func Test_LoadConfigurationFromString_WithEnv(t *testing.T) {
	os.Setenv("PG_USER", "foo")
	os.Setenv("PG_PASSWORD", "bar")

	defer func() {
		os.Unsetenv("PG_USER")
		os.Unsetenv("PG_PASSWORD")
	}()

	data, err := LoadConfigurationFromString(filecontent)

	assert.Nil(t, err)
	assert.Equal(t, expected, data)
}

func Test_LoadConfigurationFromFile_WithEnv(t *testing.T) {
	os.Setenv("PG_USER", "foo")
	os.Setenv("PG_PASSWORD", "bar")

	filename := os.TempDir() + "/test_goapp.toml"

	defer func() {
		os.Unsetenv("PG_USER")
		os.Unsetenv("PG_PASSWORD")

		os.Remove(filename)
	}()

	err := ioutil.WriteFile(filename, []byte(filecontent), 0755)

	PanicOnError(err)

	data, err := LoadConfigurationFromFile(filename)

	assert.Nil(t, err)
	assert.Equal(t, expected, data)
}
