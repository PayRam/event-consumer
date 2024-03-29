package param

import (
	"github.com/PayRam/event-emitter/service/param"
	"time"
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
	Subject              string
	FromAddress          string
	EmailTemplateName    string
	EmmitEventsOnSuccess []PostEvent
	EmmitEventsOnError   []PostEvent
}

var attributeSpecJSON = `{"refId": true, "test1": true, "postalMessageID": true}`
var Start = -(24 * time.Hour)
var End = -(1 * time.Hour)

var RoutineConfigs = []RoutineConfig{
	{
		QueryBuilder: &param.QueryBuilder{
			EventNames:             []string{"deposit-received"},
			CreatedAtRelativeStart: &Start,
			CreatedAtRelativeEnd:   &End,
			JoinWhereClause: map[string]param.JoinClause{
				"json_extract(attribute, '$.refId')": {
					Clause:  "json_extract(attribute, '$.refId')",
					Exclude: true,
				},
			},
			SubQueryBuilder: &param.QueryBuilder{
				EventNames: []string{"deposit-received-email-sent", "deposit-received-email-failed"},
			},
		},
		FromAddress:       "sam@yourdomain.com",
		EmailTemplateName: "master.tmpl",
		EmmitEventsOnSuccess: []PostEvent{
			{
				EventName:     "deposit-received-email-sent",
				CopyProfileID: true,
				AttributeSpec: map[string]bool{
					"ToAddresses": true,
					"CustomerID":  true,
					"Amount":      true,
					"MemberID":    true,
				},
			},
		},
		EmmitEventsOnError: []PostEvent{
			{
				EventName:         "deposit-received-email-failed",
				CopyProfileID:     true,
				CopyFullAttribute: true,
			},
		},
	},
	// Add more configurations as needed
}
