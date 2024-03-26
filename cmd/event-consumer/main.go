package main

import (
	"encoding/json"
	"github.com/PayRam/event-consumer/service"
	"github.com/PayRam/event-consumer/service/param"
	"html/template"
	"log"
	"os"
)

func main() {

	// *** Prepare your action config Start ***

	//jsonData, err := json.MarshalIndent(param.RoutineConfigs, "", "    ")
	//if err != nil {
	//	log.Fatalf("Error serializing RoutineConfigs: %v", err)
	//}
	//
	//fmt.Println(string(jsonData))

	// *** Prepare your action config End ***

	jsonData, err := os.ReadFile("action_config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var configs []param.RoutineConfig
	if err := json.Unmarshal(jsonData, &configs); err != nil {
		log.Fatalf("Error deserializing JSON to RoutineConfigs: %v", err)
	}

	templates := template.Must(template.ParseGlob("templates/*.tmpl"))

	service := service.NewPostalConsumerService(configs, "your.db", templates, &param.PostalConfig{
		Endpoint: "https://postal.yourdomain.com",
		APIKey:   "fdslfkjoeriewi",
	})

	service.Run()

}
