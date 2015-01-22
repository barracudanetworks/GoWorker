package config

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"testing"

// 	"github.com/barracudanetworks/GoWorker/job"
// )

// var (
// 	configFileName   = os.TempDir() + "/configFile.json"
// 	workerConfigJson = []byte(`
// 		{
// 			"name": "test",
// 			"type": "cli",
// 			"params": {
// 				"one": 1,
// 				"two": "two"
// 			}
// 		}
// 	`)

// 	providerConfigJson = []byte(`
// 		{
// 			"name": "test",
// 			"params": {
// 				"one": "one"
// 			}
// 		}
// 	`)

// 	appConfigJson = []byte(`
// 		{
// 			"providers": {
// 				"test1": {
// 					"name": "test1",
// 					"params": {
// 						"one": "one"
// 					}
// 				},
// 				"test2": {
// 					"name": "test2",
// 					"params": {
// 						"two": "two"
// 					}
// 				}
// 			},
// 			"workers": {
// 				"worker1": {
// 					"name": "1",
// 					"type": "http",
// 					"params": {
// 						"w1": 1
// 					}
// 				}
// 			}
// 		}
// 	`)
// )

// func TestParseAppConfig(t *testing.T) {
// 	a := &AppConfig{}
// 	err := json.Unmarshal(appConfigJson, a)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	if !confirmApp(a) {
// 		t.Fail()
// 	}
// }

// func TestLoadAppConfigFromFile(t *testing.T) {
// 	defer func() {
// 		err := os.Remove(configFileName)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}()
// 	err := ioutil.WriteFile(configFileName, appConfigJson, 0777)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	a, e := LoadAppConfigFromFile(configFileName)
// 	if e != nil {
// 		t.Error(err)
// 	}

// 	if !confirmApp(a) {
// 		t.Fail()
// 	}

// }

// func confirmApp(a *AppConfig) bool {
// 	if a.ProviderConfigs["test1"].Name != "test1" {
// 		return false
// 	}
// 	if a.ProviderConfigs["test2"].Params["two"] != "two" {
// 		return false
// 	}
// 	if a.WorkerConfigs["worker1"].Type != "http" {
// 		return false
// 	}
// 	return true
// }
