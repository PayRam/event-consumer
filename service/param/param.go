package param

import (
	"github.com/Pacerino/postal-go"
	"github.com/PayRam/event-emitter/service/param"
)

type ConsumerService interface {
	Run() error
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type PostalConfig struct {
	Endpoint string
	APIKey   string
}

type PostEvent struct {
	EventName         string
	CopyProfileID     bool
	CopyFullAttribute bool
	AttributeSpec     map[string]bool
}

type RoutineConfig struct {
	QueryBuilder         *param.QueryBuilder
	EmailTemplateName    string
	SubjectTemplateName  *string
	EmmitEventsOnSuccess []PostEvent
	EmmitEventsOnError   []PostEvent
	SendRequest          postal.SendRequest
}
