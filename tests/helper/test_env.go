package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
)

func PerformRequest(r http.Handler, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var req *http.Request
	w := httptest.NewRecorder()

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Add("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
		req.Header.Add("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	r.ServeHTTP(w, req)
	return w
}

func SetTestEnv() {
	os.Setenv("APPLICATION_NAME", "blog-service")
	os.Setenv("APPLICATION_PORTS", "8080")
	os.Setenv("GOOGLE_CLOUD_PROJECT_ID", "replix-394315")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_DATABASE_ID", "blogdb-yanu-widodo")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_USERS", "users")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_POSTS", "posts")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_COMMENTS", "comments")
	os.Setenv("GOOGLE_CLOUD_FIRESTORE_COLLECTION_LIKES", "likes")
	os.Setenv("GOOGLE_CLOUD_PUBSUB_TOPIC", "blogpubsub-yanu-widodo")
	os.Setenv("GOOGLE_CLOUD_PUBSUB_SUBSCRIPTION", "blogpubsub-yanu-widodo-sub")
	os.Setenv("JWT_SECRET", "LmogQeUKR3rL7JaGG2UtrPJ0TrZyTfFm")
	os.Setenv("SESSION_SECRET", "secret")
}
