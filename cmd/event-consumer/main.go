package main

import (
	"encoding/json"
	"github.com/PayRam/event-consumer/service"
	"github.com/PayRam/event-consumer/service/param"
	"html/template"
	"io/ioutil"
	"log"
)

func main() {
	//jsonData, err := json.MarshalIndent(param.RoutineConfigs, "", "    ")
	//if err != nil {
	//	log.Fatalf("Error serializing RoutineConfigs: %v", err)
	//}
	//
	//fmt.Println(string(jsonData))

	jsonData, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var configs []param.RoutineConfig
	if err := json.Unmarshal(jsonData, &configs); err != nil {
		log.Fatalf("Error deserializing JSON to RoutineConfigs: %v", err)
	}

	templates := template.Must(template.ParseGlob("templates/*.tmpl"))

	service := service.NewConsumerService(configs, "path.db", templates, &param.SMTPConfig{
		Host:     "host",
		Port:     587,
		Username: "dd",
		Password: "fd",
	})

	service.Run()

}
