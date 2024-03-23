package service

import (
	"github.com/PayRam/event-consumer/internal/serviceimpl"
	"github.com/PayRam/event-consumer/service/param"
	"gorm.io/gorm"
	"html/template"
)

func NewConsumerServiceWithDB(configs []param.RoutineConfig, db *gorm.DB, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return serviceimpl.NewConsumerServiceWithDB(configs, db, templates, smtpConfig)
}

func NewConsumerService(configs []param.RoutineConfig, dbPath string, templates *template.Template, smtpConfig *param.SMTPConfig) param.ConsumerService {
	return serviceimpl.NewConsumerService(configs, dbPath, templates, smtpConfig)
}
