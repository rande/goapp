package goapp

import (
	"bytes"
	"io/ioutil"
	"os"
	"text/template"
)

// helper function to load a configuration file as template,
// and replace {{ env 'ENV_VARIABLE' }} with the variable from
// the environnement
func LoadConfigurationFromFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)

	PanicOnError(err)

	return LoadConfigurationFromString(string(data[:]))
}

// helper function to load a configuration string as template,
// and replace {{ env 'ENV_VARIABLE' }} with the variable from
// the environnement
func LoadConfigurationFromString(data string) (string, error) {
	var err error

	t := template.New("config")
	t.Funcs(map[string]interface{}{
		"env": os.Getenv,
	})
	_, err = t.Parse(data)

	PanicOnError(err)

	b := bytes.NewBuffer([]byte{})

	err = t.Execute(b, nil)

	PanicOnError(err)

	return b.String(), nil
}
