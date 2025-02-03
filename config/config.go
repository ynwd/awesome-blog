package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Application ApplicationConfig
	GoogleCloud GoogleCloudConfig
}

type ApplicationConfig struct {
	Name  string   `json:"name"`
	Ports []string `json:"ports"`
}

type GoogleCloudConfig struct {
	ProjectID   string               `json:"project_id"`
	FirestoreDB string               `json:"firestore_db"`
	Collections FirestoreCollections `json:"collections"`
	PubSub      PubSubConfig         `json:"pubsub"`
}

type FirestoreCollections struct {
	Users    string `json:"users"`
	Posts    string `json:"posts"`
	Comments string `json:"comments"`
	Likes    string `json:"likes"`
}

type PubSubConfig struct {
	Topic        string `json:"topic"`
	Subscription string `json:"subscription"`
}

func Load() (*Config, error) {
	config := &Config{
		Application: ApplicationConfig{
			Name:  os.Getenv("APPLICATION_NAME"),
			Ports: strings.Split(os.Getenv("APPLICATION_PORTS"), ","),
		},
		GoogleCloud: GoogleCloudConfig{
			ProjectID:   os.Getenv("GOOGLE_CLOUD_PROJECT_ID"),
			FirestoreDB: os.Getenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID"),
			Collections: FirestoreCollections{
				Users:    os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS"),
				Posts:    os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_POSTS"),
				Comments: os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_COMMENTS"),
				Likes:    os.Getenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_LIKES"),
			},
			PubSub: PubSubConfig{
				Topic:        os.Getenv("GOOGLE_CLOUD_PUBSUB_TOPIC"),
				Subscription: os.Getenv("GOOGLE_CLOUD_PUBSUB_SUBSCRIPTION"),
			},
		},
	}

	return config, validate(config)
}

func validate(c *Config) error {
	if c.Application.Name == "" {
		return fmt.Errorf("APPLICATION_NAME is required")
	}
	if len(c.Application.Ports) == 0 || c.Application.Ports[0] == "" {
		return fmt.Errorf("APPLICATION_PORTS is required")
	}
	if c.GoogleCloud.ProjectID == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PROJECT_ID is required")
	}
	if c.GoogleCloud.FirestoreDB == "" {
		return fmt.Errorf("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID is required")
	}
	if c.GoogleCloud.Collections.Users == "" {
		return fmt.Errorf("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS is required")
	}
	if c.GoogleCloud.Collections.Posts == "" {
		return fmt.Errorf("GOOGLE_CLOUD_FIRESTORE_COLLECTION_POSTS is required")
	}
	if c.GoogleCloud.Collections.Comments == "" {
		return fmt.Errorf("GOOGLE_CLOUD_FIRESTORE_COLLECTION_COMMENTS is required")
	}
	if c.GoogleCloud.Collections.Likes == "" {
		return fmt.Errorf("GOOGLE_CLOUD_FIRESTORE_COLLECTION_LIKES is required")
	}
	if c.GoogleCloud.PubSub.Topic == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PUBSUB_TOPIC is required")
	}
	if c.GoogleCloud.PubSub.Subscription == "" {
		return fmt.Errorf("GOOGLE_CLOUD_PUBSUB_SUBSCRIPTION is required")
	}
	return nil
}
