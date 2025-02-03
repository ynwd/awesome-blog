package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ynwd/awesome-blog/config"
	"github.com/ynwd/awesome-blog/internal/app"
	"github.com/ynwd/awesome-blog/pkg/utils"
	"github.com/ynwd/awesome-blog/tests/helper"
)

var testApp *app.App

func TestMain(m *testing.M) {
	helper.SetTestEnv()
	if err := helper.CleanDatabase(); err != nil {
		fmt.Printf("Failed to clean database: %v\n", err)
		os.Exit(1)
	}

	if err := setupTest(); err != nil {
		fmt.Printf("Failed to setup test: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := helper.CleanDatabase(); err != nil {
		fmt.Printf("Failed to clean database: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func TestBlogServiceFlow(t *testing.T) {
	defer closeApp()
	userName := utils.GenerateRandomString(10)
	password := "Test123!"
	var authToken string
	var postID string

	// Create Tenant Owner
	t.Run("Register User", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"username": userName,
			"password": "Test123!",
		}

		w := helper.PerformRequest(testApp.Router(), "POST", "/register", requestBody, "")
		if w.Code != 201 {
			t.Fatalf("Expected status code 201, got %d", w.Code)
		}

	})

	t.Run("Login User", func(t *testing.T) {
		payload := map[string]interface{}{
			"username": userName,
			"password": password,
		}

		w := helper.PerformRequest(testApp.Router(), "POST", "/login", payload, "")

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		authToken = response["data"].(string)
		assert.NotEmpty(t, authToken)
	})

	t.Run("Invalid Login", func(t *testing.T) {
		payload := map[string]interface{}{
			"username": userName,
			"password": "wrongpassword",
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/login", payload, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Create Blog Post
	t.Run("Create Post", func(t *testing.T) {
		payload := map[string]interface{}{
			"username":    userName,
			"title":       "Test Blog Post",
			"content":     "This is a test blog post content",
			"description": "This is a test blog post description",
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/post", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		postID = response["data"].(map[string]interface{})["id"].(string)
	})

	// Add Comment
	t.Run("Add Comment", func(t *testing.T) {
		payload := map[string]interface{}{
			"username": userName,
			"post_id":  postID,
			"comment":  "This is a test comment",
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/comments", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// Add Like
	t.Run("Add Like", func(t *testing.T) {
		payload := map[string]interface{}{
			"post_id":       postID,
			"username_from": userName,
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/likes", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("Get User Summary", func(t *testing.T) {
		payload := map[string]interface{}{
			"username": userName,
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/summary", payload, authToken)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Create Blog Post Pub Sub
	t.Run("Create Post Pub Sub", func(t *testing.T) {
		payload := map[string]interface{}{
			"username":    userName,
			"title":       "Test Blog Post",
			"content":     "This is a test blog post content",
			"description": "This is a test blog post description",
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/post/pubsub", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
	})

	// Add Like Pub Sub
	t.Run("Add Like Pub Sub", func(t *testing.T) {
		payload := map[string]interface{}{
			"post_id":       postID,
			"username_from": userName,
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/likes/pubsub", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

	// Add Comment Pub sub
	t.Run("Add Comment Pub", func(t *testing.T) {
		payload := map[string]interface{}{
			"username": userName,
			"post_id":  postID,
			"comment":  "This is a test comment",
		}
		w := helper.PerformRequest(testApp.Router(), "POST", "/comments/pubsub", payload, authToken)
		assert.Equal(t, http.StatusCreated, w.Code)
	})

}

func setupTest() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	testApp = app.NewApp(cfg)
	return nil
}

func closeApp() {
	if err := testApp.Close(); err != nil {
		fmt.Printf("Failed to close app: %v", err)
	}
}
