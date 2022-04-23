package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"stocks/alerts"
	"stocks/utils"
	"strings"
	"time"
)

type NotifierRequest struct {
	Alerts         []alerts.Alert
	Subscribers    []alerts.Subscriber
	Title          string
	AlertGroupName string
}

type Notifier interface {
	Send(ctx context.Context, request NotifierRequest) (bool, error)
	SendAll(ctx context.Context, requests []NotifierRequest) (bool, error)
}

type Config struct {
	TempDirectory string
}

type emailer struct {
	config Config
}

func (e *emailer) SendAll(ctx context.Context, requests []NotifierRequest) (bool, error) {
	for _, alert := range requests {
		b, err := e.Send(ctx, alert)
		if err != nil {
			return b, err
		}
	}
	return true, nil
}

var emailTemplate = `
<body>
	<h2>
		Alerts for %s on date %s
	</h2>
	<p>
		Hi %s!
		Here are your alerts:
	</p>
	%s
</body>
`

type metadata struct {
	Email        string `json:"email"`
	FileLocation string `json:"file_location"`
	Subject      string `json:"subject"`
}

func (e *emailer) Send(_ context.Context, request NotifierRequest) (bool, error) {
	// Writes to file today in a selected directory, there is a separate github action to actually send the emails
	reg, err := regexp.Compile("[^a-zA-Z\\d]+")
	if err != nil {
		log.Fatal(err)
	}
	processedTitle := reg.ReplaceAllString(request.Title, "_")
	for _, subscriber := range request.Subscribers {
		processedName := reg.ReplaceAllString(subscriber.Name, "_")
		htmlString := fmt.Sprintf(emailTemplate, request.AlertGroupName, time.Now().Format("01-02-2006"), subscriber.Name, strings.Join(request.Alerts, ""))
		dirPathAddr := fmt.Sprintf("%s/%s_%s_tmp", e.config.TempDirectory, processedName, processedTitle)
		b, err := utils.MakeDirs([]string{dirPathAddr})
		if err != nil {
			return b, err
		}
		emailFileAddr := fmt.Sprintf("%s/email.html", dirPathAddr)
		err = os.WriteFile(emailFileAddr, []byte(htmlString), 0744)
		if err != nil {
			return false, err
		}

		var emailJSON = metadata{
			Email:        subscriber.Email,
			FileLocation: emailFileAddr,
			Subject:      request.Title,
		}

		fileData, err := json.MarshalIndent(emailJSON, "", " ")
		if err != nil {
			return false, err
		}

		jsonFileAddr := fmt.Sprintf("%s/metadata.json", dirPathAddr)
		err = ioutil.WriteFile(jsonFileAddr, fileData, 0744)
	}

	return true, nil
}

func New(config Config) Notifier {
	return &emailer{config: config}
}
