package gcp

import (
	"io/ioutil"
	"net"
	"net/http"

	"github.com/phuslu/log"
)

func ExternalIP() (net.IP, error) {
	// curl -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip", nil)
	if err != nil {
		log.Warn().Err(err).Msg("ExternalIP")
		return nil, err
	}
	req.Header.Set("Metadata-Flavor", "Google")
	res, err := client.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("ExternalIP")
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warn().Err(err).Msg("ExternalIP")
		return nil, err
	}
	return net.ParseIP(string(body)), nil
}
