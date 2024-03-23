package serviceimpl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PayRam/event-consumer/service/param"
	service2 "github.com/PayRam/event-emitter/service"
	param2 "github.com/PayRam/event-emitter/service/param"
	"gorm.io/gorm"
	"html/template"
	"log"
	"net/smtp"
	"strconv"
	"strings"
)

type service struct {
	configs      []param.RoutineConfig
	eventService param2.EventService
	templates    *template.Template
	smtpConfig   *param.SMTPConfig
	smtpAuth     smtp.Auth
}

func NewConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventServiceWithDB(db),
		templates:    templates,
		smtpConfig:   smtpConfig,
		smtpAuth:     smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host),
	}
}

func NewConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventService(dbPath),
		templates:    templates,
		smtpConfig:   smtpConfig,
		smtpAuth:     smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host),
	}
}

// CreateEvent adds a new event to the database.
func (s *service) Run() error {
	for _, config := range s.configs {
		events, err := s.eventService.QueryEvents(*config.QueryBuilder)
		if err != nil {
			return err
		}
		for _, event := range events {
			// Unmarshal the JSON string into a map
			var attrs map[string]interface{}
			if err := json.Unmarshal([]byte(event.Attribute), &attrs); err != nil {
				log.Fatalf("Error unmarshaling JSON: %v", err)
			}
			// Generate the email body
			emailBody := new(bytes.Buffer)
			if err := s.templates.ExecuteTemplate(emailBody, config.EmailTemplateName, attrs); err != nil {
				log.Fatal(err)
			}

			// Assuming config.ToAddress is a []string for simplicity in this example
			toAddresses := strings.Join(config.ToAddress, ", ")

			// Prepare email headers and body
			var emailContent bytes.Buffer
			emailContent.WriteString(fmt.Sprintf("From: %s\r\n", config.FromAddress))
			emailContent.WriteString(fmt.Sprintf("To: %s\r\n", toAddresses))
			emailContent.WriteString("Subject: Your Subject Here\r\n") // Add a subject
			emailContent.WriteString("\r\n")                           // End of headers, start of body
			emailContent.WriteString(emailBody.String())

			// Sending email
			err := smtp.SendMail(
				s.smtpConfig.Host+":"+strconv.Itoa(s.smtpConfig.Port),
				s.smtpAuth,
				config.FromAddress,
				config.ToAddress,
				emailContent.Bytes(),
			)
			if err != nil {
				for _, postEvent := range config.EmmitEventsOnError {
					// Extract fields from the event data
					emmitEvent(postEvent, event, attrs, s)
				}
			} else {
				for _, postEvent := range config.EmmitEventsOnSuccess {
					// Extract fields from the event data
					emmitEvent(postEvent, event, attrs, s)
				}
			}
		}
	}
	return nil
}

func emmitEvent(postEvent param.PostEvent, event param2.EEEvent, attrs map[string]interface{}, s *service) {
	var attrsJsonStr string
	if postEvent.CopyFullAttribute {
		attrsJsonStr = event.Attribute
	} else {
		attrsFiltered := extractFields(attrs, postEvent.AttributeSpec)
		attrsJSON, _ := json.Marshal(attrsFiltered)
		attrsJsonStr = string(attrsJSON)
	}
	if postEvent.CopyProfileID {
		s.eventService.CreateEvent(postEvent.EventName, *event.ProfileID, attrsJsonStr)
	} else {
		s.eventService.CreateGenericEvent(postEvent.EventName, attrsJsonStr)
	}
}

func extractFields(data interface{}, spec interface{}) interface{} {
	switch specTyped := spec.(type) {
	case map[string]interface{}:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil // Data does not match spec structure
		}
		result := make(map[string]interface{})
		for key, val := range specTyped {
			if val == true {
				if dataVal, exists := dataMap[key]; exists {
					result[key] = dataVal
				}
			} else {
				// Recursive case for nested structures
				if dataVal, exists := dataMap[key]; exists {
					result[key] = extractFields(dataVal, val)
				}
			}
		}
		return result
	default:
		return data // Non-object values are included directly
	}
}
