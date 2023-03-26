package config

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/k0kubun/pp"
)

var c *Config

type Config struct {
	UUID      string
	TunnelKey string
	ProxyKey  string
}

var configDir = path.Join(xdg.ConfigHome, "lpweb")
var configFile = path.Join(configDir, "config.json")

func init() {

	initConfig()
}

func initConfig() {
	os.Mkdir(configDir, os.ModePerm)

	_, err := os.Stat(configFile)
	if err == nil {
		return
	}

	if os.IsNotExist(err) {
		pp.Println("@config_not_found_init_new", configFile)

		// resp, err := http.Get("https://discovery.etcd.io/new?size=3")
		// if err != nil {
		// 	log.Fatal("couldnot generate uuid", err)
		// 	return
		// }

		// if resp.StatusCode != 200 {
		// 	log.Fatal(resp)
		// }

		// uuidBytes, err := io.ReadAll(resp.Body)
		// if err != nil {
		// 	log.Fatal(err)
		// 	return
		// }

		// uuid := strings.Split(string(uuidBytes), "/")

		config := &Config{
			UUID:      "a8a3d89ff73fc23a33f775441d8b51ac",
			TunnelKey: randomString(32),
			ProxyKey:  randomString(32),
		}

		out, err := json.Marshal(config)
		if err != nil {
			log.Fatal(err)
			return
		}

		os.WriteFile(configFile, out, 0600)
	}
}

func Get() *Config {

	if c != nil {
		return c
	}

	pp.Println("@reading_config_from", configFile)

	out, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	conf := &Config{}
	err = json.Unmarshal(out, conf)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	c = conf
	return c

}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
