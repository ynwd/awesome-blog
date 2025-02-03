package utils

import (
	"encoding/json"
	"log"

	"github.com/ynwd/awesome-blog/pkg/module"
)

func EventParser(event module.BaseEvent) ([]byte, string) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return nil, ""
	}

	var baseEvent module.BaseEvent
	if err := json.Unmarshal(data, &baseEvent); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return nil, ""
	}

	// Convert payload to map
	payloadData, err := json.Marshal(baseEvent.Payload)
	if err != nil {
		log.Printf("Error marshaling payload: %v", err)
		return nil, ""
	}

	return payloadData, baseEvent.Timestamp
}
