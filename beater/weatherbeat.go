package beater

import (
	"fmt"
	"time"
	"flag"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/libbeat/publisher"

	"github.com/rfraposa/weatherbeat/config"
)

type Weatherbeat struct {
	done   chan struct{}
	config config.Config
	client publisher.Client
}

// Creates beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Weatherbeat{
		done: make(chan struct{}),
		config: config,
	}
	return bt, nil
}


func (bt *Weatherbeat) Run(b *beat.Beat) error {
	logp.Info("weatherbeat is running! Hit CTRL-C to stop it.")

	flag.Parse()
	location := flag.Args()[0]
	if len(location) == 0 {
		fmt.Printf("***Location not set, using NewYork***")
		location = "NewYork"
	}

	resp, err := http.Get("http://api.openweathermap.org/data/2.5/find?q=" + location + "&units=imperial&appid=64d32793037e0b7054a9733f816de2f0")
	if err != nil {
		fmt.Printf("Error occurred: %s\n", err)
	}
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	//Decode the response since its a JSON object
	var weather map[string]interface{}
	if err := json.Unmarshal(out, &weather); err != nil {
        	panic(err)
    	}	

	bt.client = b.Publisher.Connect()
	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := common.MapStr{
			"@timestamp": common.Time(time.Now()),
			"type":       b.Name,
			"weather":    weather,
		}
		bt.client.PublishEvent(event)
		logp.Info("Event sent")
	}
}

func (bt *Weatherbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
