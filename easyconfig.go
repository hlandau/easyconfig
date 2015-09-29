package easyconfig // import "gopkg.in/hlandau/easyconfig.v1"

import "os"
import "fmt"
import "gopkg.in/hlandau/svcutils.v1/exepath"
import "gopkg.in/hlandau/configurable.v1"
import "gopkg.in/hlandau/easyconfig.v1/cstruct"
import "gopkg.in/hlandau/easyconfig.v1/adaptflag"
import "gopkg.in/hlandau/easyconfig.v1/adaptconf"
import "gopkg.in/hlandau/easyconfig.v1/adaptenv"
import "flag"

// Easy configurator. Set the ProgramName and call Parse, passing a pointer to
// a structure you want to fill with program-specific configuration values.
type Configurator struct {
	ProgramName    string
	configFilePath string
}

// Parse configuration values. tgt should be a pointer to a structure to be
// filled using cstruct. If nil, no structure is registered using cstruct.
func (cfg *Configurator) Parse(tgt interface{}) error {
	if tgt != nil && cfg.ProgramName != "" {
		if exepath.ProgramNameSetter == "default" {
			exepath.ProgramName = cfg.ProgramName
		}

		configurable.Register(cstruct.MustNew(tgt, cfg.ProgramName))
	}

	adaptflag.Adapt()
	adaptenv.Adapt()
	flag.Parse()
	if cfg.ProgramName != "" {
		err := adaptconf.Load(cfg.ProgramName)
		if err != nil {
			return err
		}
	}

	cfg.configFilePath = adaptconf.LastConfPath()
	return nil
}

// Like Parse, but exits with an error message if an error occurs.
func (cfg *Configurator) ParseFatal(tgt interface{}) {
	err := cfg.Parse(tgt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot load configuration file: %v\n", err)
		os.Exit(1)
	}
}

// After calling Parse successfully, returns the path to the configuration file used, if any.
func (cfg *Configurator) ConfigFilePath() string {
	return cfg.configFilePath
}

// Like Configurator.Parse. cfg may be nil.
func Parse(cfg *Configurator, tgt interface{}) error {
	if cfg == nil {
		cfg = &Configurator{}
	}

	return cfg.Parse(tgt)
}

// Like Configurator.ParseFatal. cfg may be nil.
func ParseFatal(cfg *Configurator, tgt interface{}) {
	if cfg == nil {
		cfg = &Configurator{}
	}

	cfg.ParseFatal(tgt)
}
