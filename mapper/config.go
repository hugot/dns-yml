package mapper

import (
	"fmt"
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ErrMissingConfigParameter struct {
	Name string
}

func (e *ErrMissingConfigParameter) Error() string {
	return fmt.Sprintf("Missing required config parameter \"%s\".", e.Name)
}

type Config struct {
	Database struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Address  string `yaml:"address"`
		Name     string `yaml:"name"`
	} `yaml:"database"`
}

func ConfigFromYaml(r io.Reader) (*Config, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	config := &Config{}

	err = yaml.Unmarshal(data, config)

	if err != nil {
		return config, err
	}

	switch "" {
	case config.Database.Username:
		return config, &ErrMissingConfigParameter{Name: "database.username"}
	case config.Database.Password:
		return config, &ErrMissingConfigParameter{Name: "database.password"}
	case config.Database.Address:
		return config, &ErrMissingConfigParameter{Name: "database.address"}
	case config.Database.Name:
		return config, &ErrMissingConfigParameter{Name: "database.name"}
	}

	return config, nil
}
