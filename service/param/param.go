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

type PostEvent struct {
	EventName         string
	CopyProfileID     bool
	CopyFullAttribute bool
	AttributeSpec     *string
}

type RoutineConfig struct {
	QueryBuilder         *param.QueryBuilder
	FromAddress          string
	ToAddress            []string
	EmailTemplateName    string
	EmmitEventsOnSuccess []PostEvent
	EmmitEventsOnError   []PostEvent
}

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
				EventName: []string{"deposit-received-email-sent"},
			},
		},
		FromAddress:       "noreply@example.com",
		ToAddress:         []string{"user@example.com"},
		EmailTemplateName: "master.tmpl",
		EmmitEventsOnSuccess: []PostEvent{
			{
				EventName:     "event1",
				CopyProfileID: true,
			},
		},
		EmmitEventsOnError: []PostEvent{
			{
				EventName:     "event1",
				CopyProfileID: true,
			},
		},
	},
	// Add more configurations as needed
}
