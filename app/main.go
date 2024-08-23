package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	oneloop "github.com/OneLoop-HQ/oneloop-go"
	"github.com/OneLoop-HQ/oneloop-go/client"
	"github.com/OneLoop-HQ/oneloop-go/option"
	"github.com/joho/godotenv"
)

// Handler function
func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!")
}

func main() {
	godotenv.Load(".env")
	
	oneloopSDKKey := os.Getenv("ONELOOP_SDK_KEY")
	
	fmt.Println("ONELOOP_SDK_KEY: ", oneloopSDKKey)

	oneloopClient := client.NewClient(
		option.WithToken(oneloopSDKKey),
	)

	// Middleware function to log request details
	oneloopMiddleware := func (next http.HandlerFunc, scopes []*oneloop.ApiKeyScope) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			request := oneloop.VerifyApiKeyRequest{
				Key: r.Header.Get("Authorization"),
				RequestedScopes: scopes,
			}

			response, err := oneloopClient.VerifyApiKey(r.Context(), &request)

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if response.Status == "RATE_LIMITED" {
				http.Error(w, "Rate limited", http.StatusTooManyRequests)
				return
			} 

			if response.Status == "INVALID_SCOPES" {
				http.Error(w, "Invalid scopes", http.StatusForbidden)
				return
			}
			
			if response.Status != "VALID" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
	}
}

	scopes := []*oneloop.ApiKeyScope{
		{
			Representation: "billing",
			Read: true,
			Create: true,
			Update: true,
			Del: true,
		},
	}

	http.HandleFunc("/", oneloopMiddleware(helloHandler, scopes))

	fmt.Println("Server is running on http://localhost:8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}