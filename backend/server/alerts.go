package server

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
)

func (s *LogService) sendAlert(message string) {
	log.Printf("ALERT: %s", message)
	// Integrate email or Slack notifications here
}

type AlertRule struct {
	Keyword     string
	Threshold   int
	TimeWindow  time.Duration
	Description string
}

type Alerts struct {
	client    *elastic.Client
	rules     []AlertRule
	broadcast chan string
}

func NewAlerts(client *elastic.Client) *Alerts {
	return &Alerts{
		client:    client,
		rules:     []AlertRule{},
		broadcast: make(chan string, 100), // Channel to send alert notifications
	}
}

// AddRule adds a new alert rule
func (a *Alerts) AddRule(keyword string, threshold int, timeWindow time.Duration, description string) {
	a.rules = append(a.rules, AlertRule{
		Keyword:     keyword,
		Threshold:   threshold,
		TimeWindow:  timeWindow,
		Description: description,
	})
	log.Printf("Added alert rule: %s", description)
}

// StartMonitoring starts monitoring logs based on alert rules
func (a *Alerts) StartMonitoring() {
	go func() {
		for {
			for _, rule := range a.rules {
				a.checkRule(rule)
			}
			time.Sleep(1 * time.Minute) // Check every minute
		}
	}()
}

// checkRule checks if an alert rule is triggered
func (a *Alerts) checkRule(rule AlertRule) {
	cutoff := time.Now().Add(-rule.TimeWindow).Unix() * 1000 // Convert to milliseconds

	// Search logs for the specified keyword within the time window
	query := elastic.NewBoolQuery().
		Must(elastic.NewMatchQuery("message", rule.Keyword)).
		Filter(elastic.NewRangeQuery("timestamp").Gte(cutoff))

	result, err := a.client.Search().
		Index("logs").
		Query(query).
		Do(context.Background())
	if err != nil {
		log.Printf("Error querying logs for alerts: %v", err)
		return
	}

	if int(result.TotalHits()) >= rule.Threshold {
		alertMessage := rule.Description
		log.Printf("ALERT TRIGGERED: %s", alertMessage)
		a.broadcast <- alertMessage
	}
}

// GetBroadcastChannel returns the alert notification channel
func (a *Alerts) GetBroadcastChannel() chan string {
	return a.broadcast
}
