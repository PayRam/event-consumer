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
			// Convert all string fields in attrs to template.HTML
			for k, v := range attrs {
				if strVal, ok := v.(string); ok {
					attrs[k] = template.HTML(strVal)
				}
			}
			// Generate the email body
			emailBody := new(bytes.Buffer)
			if err := s.templates.ExecuteTemplate(emailBody, config.EmailTemplateName, attrs); err != nil {
				logger.Error("Error executing email template: %v", err)
			}

			var subject = config.SendRequest.Subject

			if config.SubjectTemplateName != nil {
				subjectBuffer := new(bytes.Buffer)
				if err := s.templates.ExecuteTemplate(subjectBuffer, *config.SubjectTemplateName, attrs); err != nil {
					subject = config.SendRequest.Subject
				} else {
					// If there's no error, use the executed template as the subject
					subject = subjectBuffer.String()
				}
			}

			var err error

			if s.consumerType == "POSTAL" {
				attrs, err = s.sendEmailUsingPostal(config, subject, emailBody, attrs)
			} else {
				err = s.sendEmailUsingSMTP(config, subject, emailBody, attrs)
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

func (s *service) sendEmailUsingSMTP(config param.RoutineConfig, subject string, emailBody *bytes.Buffer, attrs map[string]interface{}) error {
	// Assuming config.ToAddress is a []string for simplicity in this example
	toAddresses := strings.Join(getToAddresses(attrs), ", ")

	// Prepare email headers and body
	var emailContent bytes.Buffer
	emailContent.WriteString(fmt.Sprintf("From: %s\r\n", config.SendRequest.From))
	emailContent.WriteString(fmt.Sprintf("To: %s\r\n", toAddresses))
	emailContent.WriteString("Subject: " + subject + "\r\n") // Add a subject
	emailContent.WriteString("\r\n")                         // End of headers, start of body
	emailContent.WriteString(emailBody.String())

	err := smtp.SendMail(
		s.smtpConfig.Host+":"+strconv.Itoa(s.smtpConfig.Port),
		s.smtpAuth,
		config.SendRequest.From,
		getToAddresses(attrs),
		emailContent.Bytes(),
	)
	if err != nil {
		logger.Error("Error sending email(SMTP): %v", err)
		return err
	}
	return nil
}

func (s *service) sendEmailUsingPostal(config param.RoutineConfig, subject string, emailBody *bytes.Buffer, attrs map[string]interface{}) (map[string]interface{}, error) {

	toAddresses := getToAddresses(attrs)

	message := &config.SendRequest

	message.Subject = subject

	if toAddresses != nil && len(toAddresses) > 0 {
		message.To = toAddresses
	}

	if emailBody.String() != "" {
		message.HTMLBody = emailBody.String()
	}

	var resp *postal.SendResponse
	resp, _, err := s.client.Send.Send(context.TODO(), message)
	if err != nil {
		logger.Error("Error sending email(POSTAL): %v", err)
		return attrs, err
	} else {
		attrs["PostalMessageID"] = resp.MessageID
	}
	return attrs, nil
}

func getToAddresses(attrs map[string]interface{}) []string {
	if toAddresses, ok := attrs["ToAddresses"].([]interface{}); ok {
		var strToAddresses []string
		for _, addr := range toAddresses {
			if strAddr, ok := addr.(string); ok {
				strToAddresses = append(strToAddresses, strAddr)
			} else {
				// Handle the case where the conversion is not possible
				logger.Warn("Warning: Non-string value encountered in ToAddresses")
			}
		}
		return strToAddresses
	} else {
		logger.Error("Error: ToAddresses is not a slice of interface{}")
	}
	return nil
}

func emmitEvent(postEvent param.PostEvent, event param2.EEEvent, attrs map[string]interface{}, s *service) {
	var attrsJsonStr string
	if postEvent.CopyFullAttribute {
		attrsJsonStr = event.Attribute
	} else {
		if postEvent.AttributeSpec != nil {
			attrsFiltered := extractFields(attrs, postEvent.AttributeSpec)
			if attrsFiltered == nil {
				attrsJsonStr = event.Attribute
			} else {
				attrsJSON, _ := json.Marshal(attrsFiltered)
				attrsJsonStr = string(attrsJSON)
			}
		}

	}
	if postEvent.CopyProfileID {
		_, err := s.eventService.CreateEvent(postEvent.EventName, attrsJsonStr, event.ProfileID)
		if err != nil {
			logger.Error("Error creating event: %v", err)
		}
	} else {
		_, err := s.eventService.CreateSimpleEvent(postEvent.EventName, attrsJsonStr)
		if err != nil {
			logger.Error("Error creating event: %v", err)
		}
	}
}

func extractFields(data interface{}, spec interface{}) interface{} {
	switch specTyped := spec.(type) {
	case map[string]bool:
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
