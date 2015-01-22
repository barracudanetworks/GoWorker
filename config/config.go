package config

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

const (
	DEFAULT_STATS_PORT = ":9090"
)

var (
	DEFAULT_LUA_PATH  = os.Getenv("GOPATH") + "/src/github.com/barracudanetworks/GoWorker/lua"
	WRONG_CONFIG_TYPE = errors.New("config: wrong config type")
	LUA_PATH          = DEFAULT_LUA_PATH
)

func init() {
	gob.Register(AppConfig{})
	gob.Register(Config{})
}

// Configer has the ability to return a config struct and a method to apply that configuartion
type Configer interface {
	ConfigStruct() interface{}
	Init(i interface{}) error
}

type ConfigBlock []map[string]json.RawMessage
type ConfigPair struct {
	Type   string
	Config Config
}

// AppConfig contains all of the information needed to configure the app
type AppConfig struct {
	ProviderConfigs        []ConfigPair
	WorkerConfigs          []ConfigPair
	FailureHanldlerConfigs []ConfigPair
	ManagerToManager       string      `json:"manager_to_manager_port"`
	StatsPort              string      `json:"stats_port"`
	LuaPath                string      `json:"lua_path"`
	RawProviders           ConfigBlock `json:"providers"`
	RawWorkers             ConfigBlock `json:"workers"`
	RawFailureHandler      ConfigBlock `json:"failure_handler"`
}

// defaultAppConfig returns a app config with defaults params
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		ProviderConfigs: []ConfigPair{},
		WorkerConfigs:   []ConfigPair{},
		LuaPath:         DEFAULT_LUA_PATH,
		StatsPort:       DEFAULT_STATS_PORT,
	}
}

// LoadAppConfigFromFile loads configuration needed for the initilization of the manager
func LoadAppConfigFromFile(fileName string) (*AppConfig, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	a := DefaultAppConfig()
	err = json.Unmarshal(b, a)

	// cast the raw configs
	a.ProviderConfigs = castConfigs(a.RawProviders)
	a.WorkerConfigs = castConfigs(a.RawWorkers)
	a.FailureHanldlerConfigs = castConfigs(a.RawFailureHandler)
	LUA_PATH = a.LuaPath
	return a, err
}

// Holds generic config information
type Config json.RawMessage

// Apply decode the config into the given interface and call it's init function
func (c Config) Apply(conf Configer) error {
	i := conf.ConfigStruct()
	err := json.Unmarshal([]byte(c), i)
	if err != nil {
		return err
	}
	return conf.Init(i)
}

// Encode endcode the given interface onto a Config struct
func (c Config) Encode(i interface{}) error {
	b, err := json.Marshal(i)
	if err != nil {
		return err
	}
	c = Config(b)
	return nil
}

func castConfigs(c ConfigBlock) []ConfigPair {
	n := make([]ConfigPair, len(c))
	for i := range n {
		for k, v := range c[i] {
			n[i] = ConfigPair{
				Type:   k,
				Config: Config(v),
			}
			break
		}
	}
	return n
}
