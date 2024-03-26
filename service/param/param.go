package param

import "github.com/PayRam/event-emitter/service/param"

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
	AttributeSpec     *string
}

type RoutineConfig struct {
	QueryBuilder         *param.QueryBuilder
	Subject              string
	FromAddress          string
	ToAddress            []string
	EmailTemplateName    string
	EmmitEventsOnSuccess []PostEvent
	EmmitEventsOnError   []PostEvent
}

var attributeSpecJSON = `{"refId": true, "test1": true, "postalMessageID": true}`

var RoutineConfigs = []RoutineConfig{
	{
		QueryBuilder: &param.QueryBuilder{
			EventName: []string{"deposit-received"},
			JoinWhereClause: map[string]param.JoinClause{
				"json_extract(attribute, '$.refId')": {
					Clause:  "json_extract(attribute, '$.refId')",
					Exclude: true,
				},
			},
			QueryBuilderParam: &param.QueryBuilder{
				EventName: []string{"deposit-received-email-sent", "deposit-received-email-failed"},
			},
		},
		FromAddress:       "sam@yourdomain.com",
		ToAddress:         []string{"sam@payram.app"},
		EmailTemplateName: "master.tmpl",
		EmmitEventsOnSuccess: []PostEvent{
			{
				EventName:     "deposit-received-email-sent",
				CopyProfileID: true,
				AttributeSpec: &attributeSpecJSON,
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
