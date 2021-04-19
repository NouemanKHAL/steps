package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stackpulse/steps-sdk-go/env"
	"github.com/stackpulse/steps-sdk-go/step"

	datadog "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
)

const stackpulseSource = "STACKPULSE"

type EventPost struct {
	ApiKey         string   `env:"API_KEY,required"`
	UseEu          bool     `env:"USE_EU"`
	Title          string   `env:"TITLE,required"`
	Text           string   `env:"TEXT,required"`
	Tags           []string `env:"TAGS"`
	AggregationKey string   `env:"AGGREGATION_KEY"`
	RelatedEventId int      `env:"RELATED_EVENT_ID"`
}

type createRequest struct {
	Title          string   `json:"title"`
	Text           string   `json:"text"`
	Tags           []string `json:"tags,omitempty"`
	AggregationKey string   `json:"aggregation_key,omitempty"`
	RelatedEventId int      `json:"related_event_id,omitempty"`
	SourceTypeName string   `json:"source_type_name,omitempty"`
}

type output struct {
	ID int `json:"id"`
	step.Outputs
}

func (s *EventPost) Init() error {
	err := env.Parse(s)
	if err != nil {
		return err
	}

	return nil
}

func (s *EventPost) getURL(useEU bool) string {
	hostExt := "com"
	if useEU {
		hostExt = "eu"
	}
	return fmt.Sprintf("https://api.datadoghq.%s/api/v1/events", hostExt)
}

func (s *EventPost) Run() (int, []byte, error) {
	//prepare request body
	Title := s.Title
	Text := s.Text
	Tags := s.Tags
	AggregationKey := s.AggregationKey
	RelatedEventId := int64(s.RelatedEventId)
	stackpulseSource := string(stackpulseSource)

	body := datadog.EventCreateRequest{
		Title:          Title,
		Text:           Text,
		Tags:           &Tags,
		AggregationKey: &AggregationKey,
		RelatedEventId: &RelatedEventId,
		SourceTypeName: &stackpulseSource,
	}

	ctx := context.WithValue(
		context.Background(),
		datadog.ContextAPIKeys,
		map[string]datadog.APIKey{
			"apiKeyAuth": {
				Key: s.ApiKey,
			},
		},
	)

	configuration := datadog.NewConfiguration()

	// send request
	apiClient := datadog.NewAPIClient(configuration)
	responseBody, r, err := apiClient.EventsApi.CreateEvent(ctx).Body(body).Execute()
	if err != nil {
		return step.ExitCodeFailure, nil, err
	}

	// response from `CreateEvent`: EventCreateResponse
	rawRespBody, err := json.MarshalIndent(responseBody, "", "  ")
	if err != nil {
		return step.ExitCodeFailure, nil, err
	}

	//check HTTP error
	if r.StatusCode != http.StatusAccepted {
		return step.ExitCodeFailure, nil, fmt.Errorf("failed to create event. got response code: %d: %s",
			r.StatusCode, string(rawRespBody))
	}

	var eventID int
	if responseBody.Id != nil {
		eventID = int(*responseBody.Id)
	}

	//prepare output
	stepOutput := output{
		ID:      eventID,
		Outputs: step.Outputs{Object: responseBody},
	}
	jsonOutput, err := json.Marshal(stepOutput)
	if err != nil {
		return step.ExitCodeFailure, nil, err
	}

	return step.ExitCodeOK, jsonOutput, nil
}

func main() {
	step.Run(&EventPost{})
}
