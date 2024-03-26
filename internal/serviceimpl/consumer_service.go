package serviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Pacerino/postal-go"
	"github.com/PayRam/event-consumer/internal/logger"
	"github.com/PayRam/event-consumer/service/param"
	service2 "github.com/PayRam/event-emitter/service"
	param2 "github.com/PayRam/event-emitter/service/param"
	"gorm.io/gorm"
	"html/template"
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
	client       *postal.Client
	consumerType string //SMTP or POSTAL
}

func NewSMTPConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventServiceWithDB(db),
		templates:    templates,
		smtpConfig:   smtpConfig,
		smtpAuth:     smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host),
		consumerType: "SMTP",
	}
}

func NewSMTPConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventService(dbPath),
		templates:    templates,
		smtpConfig:   smtpConfig,
		smtpAuth:     smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host),
		consumerType: "SMTP",
	}
}

func NewPostalConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, postalConfig *param.PostalConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventServiceWithDB(db),
		templates:    templates,
		client:       postal.NewClient(postalConfig.Endpoint, postalConfig.APIKey),
		consumerType: "POSTAL",
	}
}

func NewPostalConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, postalConfig *param.PostalConfig) param.ConsumerService {
	return &service{
		configs:      configs,
		eventService: service2.NewEventService(dbPath),
		templates:    templates,
		client:       postal.NewClient(postalConfig.Endpoint, postalConfig.APIKey),
		consumerType: "POSTAL",
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
				logger.Error("Error unmarshalling event attribute: %v", err)
				continue
			}
			// Generate the email body
			emailBody := new(bytes.Buffer)
			if err := s.templates.ExecuteTemplate(emailBody, config.EmailTemplateName, attrs); err != nil {
				logger.Error("Error executing email template: %v", err)
			}

			var subject = config.Subject
			if attrs["emailSubject"] != nil {
				subject = attrs["emailSubject"].(string)
			}

			var err error

			if s.consumerType == "POSTAL" {
				err = s.sendEmailUsingPostal(config, subject, emailBody, attrs)
			} else {
				err = s.sendEmailUsingSMTP(config, subject, emailBody)
			}
			if err != nil {
				for _, postEvent := range config.EmmitEventsOnError {
					emmitEvent(postEvent, event, attrs, s)
				}
			} else {
				for _, postEvent := range config.EmmitEventsOnSuccess {
					emmitEvent(postEvent, event, attrs, s)
				}
			}
		}
	}
	return nil
}

func (s *service) sendEmailUsingSMTP(config param.RoutineConfig, subject string, emailBody *bytes.Buffer) error {
	// Assuming config.ToAddress is a []string for simplicity in this example
	toAddresses := strings.Join(config.ToAddress, ", ")

	// Prepare email headers and body
	var emailContent bytes.Buffer
	emailContent.WriteString(fmt.Sprintf("From: %s\r\n", config.FromAddress))
	emailContent.WriteString(fmt.Sprintf("To: %s\r\n", toAddresses))
	emailContent.WriteString("Subject: " + subject + "\r\n") // Add a subject
	emailContent.WriteString("\r\n")                         // End of headers, start of body
	emailContent.WriteString(emailBody.String())

	err := smtp.SendMail(
		s.smtpConfig.Host+":"+strconv.Itoa(s.smtpConfig.Port),
		s.smtpAuth,
		config.FromAddress,
		config.ToAddress,
		emailContent.Bytes(),
	)
	if err != nil {
		logger.Error("Error sending email(SMTP): %v", err)
		return err
	}
	return nil
}

func (s *service) sendEmailUsingPostal(config param.RoutineConfig, subject string, emailBody *bytes.Buffer, attrs map[string]interface{}) error {
	message := &postal.SendRequest{
		To:       config.ToAddress,
		From:     config.FromAddress,
		Subject:  subject,
		HTMLBody: emailBody.String(),
	}
	var resp *postal.SendResponse
	resp, _, err := s.client.Send.Send(context.TODO(), message)
	if err != nil {
		logger.Error("Error sending email(POSTAL): %v", err)
		return err
	} else {
		attrs["postalMessageID"] = resp.MessageID
	}
	return nil
}

func emmitEvent(postEvent param.PostEvent, event param2.EEEvent, attrs map[string]interface{}, s *service) {
	var attrsJsonStr string
	if postEvent.CopyFullAttribute {
		attrsJsonStr = event.Attribute
	} else {
		var specAttrs map[string]interface{}
		if err := json.Unmarshal([]byte(*postEvent.AttributeSpec), &specAttrs); err != nil {
			logger.Error("Error unmarshalling attribute spec: %v", err)
			attrsJsonStr = event.Attribute
		} else {
			attrsFiltered := extractFields(attrs, specAttrs)
			if attrsFiltered == nil {
				attrsJsonStr = event.Attribute
			} else {
				attrsJSON, _ := json.Marshal(attrsFiltered)
				attrsJsonStr = string(attrsJSON)
			}
		}
	}
	if postEvent.CopyProfileID {
		err := s.eventService.CreateEvent(postEvent.EventName, *event.ProfileID, attrsJsonStr)
		if err != nil {
			logger.Error("Error creating event: %v", err)
		}
	} else {
		err := s.eventService.CreateGenericEvent(postEvent.EventName, attrsJsonStr)
		if err != nil {
			logger.Error("Error creating event: %v", err)
		}
	}
}

func extractFields(data interface{}, spec interface{}) interface{} {
	switch specTyped := spec.(type) {
	case map[string]interface{}:
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			logger.Error("Data does not match spec structure")
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
