package easyconfig // import "gopkg.in/hlandau/easyconfig.v1"

import "os"
import "fmt"
import "gopkg.in/hlandau/configurable.v1"
import "gopkg.in/hlandau/easyconfig.v1/cstruct"
import "gopkg.in/hlandau/easyconfig.v1/adaptflag"
import "gopkg.in/hlandau/easyconfig.v1/adaptconf"
import "gopkg.in/hlandau/easyconfig.v1/adaptenv"
import "flag"

type Configurator struct {
	ProgramName    string
	configFilePath string
}

func (cfg *Configurator) Parse(tgt interface{}) error {
	c := cstruct.MustNew(tgt, cfg.ProgramName)
	configurable.Register(c)
	adaptflag.Adapt()
	adaptenv.Adapt()
	flag.Parse()
	err := adaptconf.Load(cfg.ProgramName)
	if err != nil {
		return err
	}

	cfg.configFilePath = adaptconf.LastConfPath()
	return nil
}

func (cfg *Configurator) ParseFatal(tgt interface{}) {
	err := cfg.Parse(tgt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load configuration file: %v\n", err)
		os.Exit(1)
	}
}

func (cfg *Configurator) ConfigFilePath() string {
	return cfg.configFilePath
}
