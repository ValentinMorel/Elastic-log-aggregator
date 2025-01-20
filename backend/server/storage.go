package server

import (
	"context"
	"encoding/json"
	"log"

	proto "log-aggregator/pb/logmsg"

	"github.com/olivere/elastic/v7"
)

type Storage struct {
	client *elastic.Client
}

func NewStorage() *Storage {
	client, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"))
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}

	// Ensure the logs index exists
	index := "logs"
	exists, err := client.IndexExists(index).Do(context.Background())
	if err != nil {
		log.Fatalf("Failed to check if index exists: %v", err)
	}

	if !exists {
		_, err := client.CreateIndex(index).BodyString(`
		{
			"settings": {
				"number_of_shards": 1,
				"number_of_replicas": 0
			},
			"mappings": {
				"properties": {
					"source": { "type": "keyword" },
					"log_level": { "type": "keyword" },
					"message": { "type": "text" },
					"timestamp": { "type": "date" }
				}
			}
		}`).Do(context.Background())
		if err != nil {
			log.Fatalf("Failed to create index: %v", err)
		}
		log.Printf("Created index: %s", index)
	}

	return &Storage{client: client}
}

func (s *Storage) SaveLog(logMsg *proto.LogMessage) {
	_, err := s.client.Index().
		Index("logs").
		BodyJson(logMsg).
		Do(context.Background())
	if err != nil {
		log.Printf("Failed to save log: %v", err)
	}
}

func (s *Storage) QueryLogs(query *proto.LogQuery) []*proto.LogMessage {
	search := elastic.NewBoolQuery().
		Must(elastic.NewTermQuery("app_name", query.Source)).
		Must(elastic.NewRangeQuery("timestamp").Gte(query.StartTime).Lte(query.EndTime))

	result, err := s.client.Search().
		Index("logs").
		Query(search).
		Do(context.Background())
	if err != nil {
		log.Printf("Failed to query logs: %v", err)
		return nil
	}

	var logs []*proto.LogMessage
	for _, hit := range result.Hits.Hits {
		var logMsg proto.LogMessage
		json.Unmarshal(hit.Source, &logMsg)
		logs = append(logs, &logMsg)
	}
	return logs
}
