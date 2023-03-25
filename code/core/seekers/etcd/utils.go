package etcd

import (
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/adrg/xdg"
)

var configDir = path.Join(xdg.ConfigHome, "libweb")
var uuidFile = path.Join(configDir, "uuid")

func InitilizeUUID() {
	os.Mkdir(configDir, os.ModePerm)

	_, err := os.Stat(uuidFile)
	if err == nil {
		return
	}

	if os.IsNotExist(err) {
		resp, err := http.Get("https://discovery.etcd.io/new?size=3")
		if err != nil {
			log.Fatal(err)
			return
		}

		if resp.StatusCode != 200 {
			log.Fatal(resp)
		}

		out, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			return
		}

		os.WriteFile(uuidFile, out, 0600)
	}
}
