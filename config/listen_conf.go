package config

import (
	"fmt"
	"io/ioutil"
	"net"

	"gopkg.in/yaml.v2"
)

// Conf List of config entries
type Conf struct {
	Uid      string `yaml:"uid"`
	Gid      string `yaml:"gid"`
	Http     []ListenConf
}

type ListenConf struct {
	Listen string   `yaml:"listen"`
	Allow  []subnet `yaml:"allow"`
	Deny   []subnet `yaml:"deny"`
	RateLimit RateLimit `yaml:"ratelimit"`
}

// RateLimit -- per host and global
type RateLimit struct {
	Global  int `yaml:"global"`
	PerHost int `yaml:"perhost"`
}

// An IP/Subnet
type subnet struct {
	net.IPNet
}

// UnmarshalYAML Custom unmarshaler for IPNet
func (ipn *subnet) UnmarshalYAML(unm func(v interface{}) error) error {
	var s string

	// First unpack the bytes as a string. We then parse the string
	// as a CIDR
	err := unm(&s)
	if err != nil {
		return err
	}

	_, nets, err := net.ParseCIDR(s)
	if err == nil {
		ipn.IP = nets.IP
		ipn.Mask = nets.Mask
	}
	return err
}

// ReadYAML Parses config file in YAML format and return
func ReadYAML(fn string) (*Conf, error) {
	yml, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("can't read config file %s: %s", fn, err)
	}

	var cfg Conf
	err = yaml.Unmarshal(yml, &cfg)
	if err != nil {
		return nil, fmt.Errorf("can't parse config file %s: %s", fn, err)
	}

	return &cfg, nil
}
