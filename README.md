# event-consumer

## Events Generated using event emitter
- deposit-received
```database
1. deposit-received	            123	{"refId": "123456","test":"value","test1":"value1"}
2. deposit-received	            323	{"refId": "123457","paymentLink":"https://google.com", "test":"yes","test1":"uels"}
3. deposit-received	            123	{"refId": "123458"}
```
- deposit-received-email-sent: entry for the email sent for event no 1
```database
4. deposit-received-email-sent	123	{"postalMessageID":"resp.MessageID","refId":"123456","
```
- deposit-received-email-failed: entry for the email sending failed for event no 2
```database
5. deposit-received-email-failed	123	{"refId": "123458"}
```

## Usage

- Prepare RoutineConfig: this is query builder to query on events
```
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
	// You can create multiple routines
}
```
- Import in your project: create an action_config.json file
```
jsonData, err := json.MarshalIndent(param.RoutineConfigs, "", "    ")
if err != nil {
    log.Fatalf("Error serializing RoutineConfigs: %v", err)
}

fmt.Println(string(jsonData))
```
- Run Consumer
```
jsonData, err := ioutil.ReadFile("action_config.json")
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var configs []param.RoutineConfig
	if err := json.Unmarshal(jsonData, &configs); err != nil {
		log.Fatalf("Error deserializing JSON to RoutineConfigs: %v", err)
	}

	templates := template.Must(template.ParseGlob("templates/*.tmpl"))

	service := service.NewPostalConsumerService(configs, "your.db", templates, &param.PostalConfig{
		Endpoint: "https://postal.yourdomain.com",
		APIKey:   "fdslfkjoeriewi",
	})

	service.Run()
```

## Email Templates

- body.tmpl
```
<main>
    <p>Hello Test,</p>
    <p>Welcome to our service!</p>
</main>
```
- header.tmpl
```
<!DOCTYPE html>
<html>
<head>
    <title>Test</title>
</head>
<body>
<header>
    <h1>Our Company Newsletter</h1>
</header>
```
- footer.tmpl
```
<footer>
    <p>Copyright © Our Company. All rights reserved.</p>
</footer>
</body>
</html>
```
- master.tmpl
```
{{template "header.tmpl" .}}
{{template "body.tmpl" .}}
{{template "footer.tmpl" .}}
```

