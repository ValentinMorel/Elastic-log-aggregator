package server

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/olivere/elastic/v7"
)

func (s *LogService) sendAlert(message string) {
	log.Printf("ALERT: %s", message)
	// Integrate email or anything else notifications here
}

type AlertRule struct {
	Keyword     string
	Threshold   int
	TimeWindow  time.Duration
	Description string
}

type Alerts struct {
	client         *elastic.Client
	rules          []AlertRule
	broadcast      chan string
	triggeredCount atomic.Int32
}

func NewAlerts(client *elastic.Client) *Alerts {
	return &Alerts{
		client:    client,
		rules:     []AlertRule{},
		broadcast: make(chan string), // Channel to send alert notifications
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

func (a *Alerts) StartMonitoring() {
	go func() {
		for {
			for _, rule := range a.rules {
				a.checkRule(rule)
			}
			time.Sleep(1 * time.Minute)
		}
	}()
}

func (a *Alerts) checkRule(rule AlertRule) {
	cutoff := time.Now().Add(-rule.TimeWindow).Unix() * 1000 // Convert to milliseconds

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
		// Lock before incrementing
		a.triggeredCount.Add(1)

		alertMessage := rule.Description
		log.Printf("ALERT TRIGGERED: %s", alertMessage)

		// Broadcast the alert to WebSocket clients
		a.broadcast <- alertMessage
	}
}

func (a *Alerts) GetBroadcastChannel() chan string {
	return a.broadcast
}

func (a *Alerts) GetAlertsTriggered() int {
	return int(a.triggeredCount.Load())
}
