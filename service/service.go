package service

import (
	"github.com/PayRam/event-consumer/internal/serviceimpl"
	"github.com/PayRam/event-consumer/service/param"
	"gorm.io/gorm"
	"html/template"
)

func NewSMTPConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return serviceimpl.NewSMTPConsumerServiceWithDB(configs, db, templates, smtpConfig)
}

func NewSMTPConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return serviceimpl.NewSMTPConsumerService(configs, dbPath, templates, smtpConfig)
}

func NewPostalConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, postalConfig *param.PostalConfig) param.ConsumerService {
	return serviceimpl.NewPostalConsumerServiceWithDB(configs, db, templates, postalConfig)
}

func NewPostalConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, postalConfig *param.PostalConfig) param.ConsumerService {
	return serviceimpl.NewPostalConsumerService(configs, dbPath, templates, postalConfig)
}
