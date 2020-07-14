package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/FusionAuth/go-client/pkg/fusionauth"
	"github.com/joho/godotenv"
)

var (
	AppId        string
	AppName      string
	ClientID     string
	ClientSecret string
	APIKey       string
	FAHost       string
	FAPort       string
	BaseUrl      string
	httpClient   = &http.Client{
		Timeout: time.Second * 10,
	}
	faClient *fusionauth.FusionAuthClient
)

func upsertFusionAuthApplication() {
	log.Printf("Upserting FusionAuth application with ID: %+v", AppId)

	// Check if the app already exists
	app, err := faClient.RetrieveApplication(AppId)
	if err != nil {
		// Create the app if it doesn't exist
		req := fusionauth.ApplicationRequest{
			Application: fusionauth.Application{
				Active: true,
				Name:   AppName,
			},
		}
		log.Print("FusionAuth application not found. Creating...")
		app, errors, err := faClient.CreateApplication(AppId, req)
		if err != nil {
			log.Print("Failed to create FusionAuth application")
			log.Print(errors)
			log.Fatal(err)
		}

		log.Print("FusionAuth application created successfully")
		ClientID = app.Application.OauthConfiguration.ClientId
		ClientSecret = app.Application.OauthConfiguration.ClientSecret
	} else {
		log.Print("FusionAuth application already exists")
		ClientID = app.Application.OauthConfiguration.ClientId
		ClientSecret = app.Application.OauthConfiguration.ClientSecret
	}

	log.Printf("ClientID: %+v", ClientID)
	log.Printf("ClientSecret: %+v", ClientSecret)
}

func main() {
	godotenv.Load(".env")
	AppId = os.Getenv("FA_APP_ID")
	AppName = os.Getenv("FUSIONAUTH_APP_NAME")
	APIKey = os.Getenv("FA_API_KEY")
	FAHost = os.Getenv("FA_HOST")
	FAPort = os.Getenv("FA_PORT")
	BaseUrl = os.Getenv("BASE_URL")

	publicPort := os.Getenv("PUBLIC_PORT")

	host := fmt.Sprintf("http://%s:%s", FAHost, FAPort)
	baseURL, _ := url.Parse(host)
	faClient = fusionauth.NewClient(httpClient, baseURL, APIKey)
	faClient.Debug = true

	// Won't be able to seed the oauth application until FusionAuth is done booting
	time.Sleep(15 * time.Second)
	upsertFusionAuthApplication()

	r := setupRouter()
	// Listen and Serve on 0.0.0.0:8080
	r.Run(fmt.Sprintf(":%s", publicPort))
}
