package utils

import "os"

func MLServiceURL() string {
	url := os.Getenv("ML_SERVICE_URL")
	if url == "" {
		return "http://localhost:8000"
	}
	return url
}
