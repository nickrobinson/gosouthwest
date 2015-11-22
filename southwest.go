package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mattbaird/gochimp"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//Southwest structure
type Southwest struct {
	FirstName          string
	LastName           string
	ConfirmationNumber string
	Url                string
}

type SouthwestResponse struct {
	Intermsg                         string        `json:"interMsg"`
	FormInput                        []interface{} `json:"form_input"`
	Hazmatnotecontent                string        `json:"hazmatnoteContent"`
	Departitinerary                  []interface{} `json:"departItinerary"`
	Title                            string        `json:"title"`
	Hazviewmorelinkcontent3B         string        `json:"hazViewMoreLinkContent3b"`
	Hazviewmorelinkcontent3A         string        `json:"hazViewMoreLinkContent3a"`
	CodaPnr                          string        `json:"CODA_PNR"`
	Errmsg                           string        `json:"errmsg"`
	Hazviewmorelinkcontent4Linktext  string        `json:"hazViewMoreLinkContent4LinkText"`
	Httpstatuscode                   int           `json:"httpStatusCode"`
	Hazviewmorelinkcontent4          string        `json:"hazViewMoreLinkContent4"`
	Hazmatnotecontentbold            string        `json:"hazmatnoteContentBold"`
	Hazmatnotelinktext               string        `json:"hazmatnoteLinkText"`
	Returnoperator                   []interface{} `json:"returnOperator"`
	Departoperator                   []interface{} `json:"departOperator"`
	Opstatus                         string        `json:"opstatus"`
	Hazmatnotetitle                  string        `json:"hazmatnoteTitle"`
	FlightcheckinURL                 string        `json:"flightcheckin_url"`
	Hazviewmorelinkcontent4Linkvalue string        `json:"hazViewMoreLinkContent4LinkValue"`
	Hazviewmorelinkcontentbold       string        `json:"hazViewMoreLinkContentBold"`
	Hazviewmorelinkcontent1          string        `json:"hazViewMoreLinkContent1"`
	Hazmatnotelinkvalue              string        `json:"hazmatnoteLinkValue"`
	Hazviewmorelinkcontent2          string        `json:"hazViewMoreLinkContent2"`
	PassengerNames                   []interface{} `json:"passenger_names"`
	Hazviewmorelinkheading           string        `json:"hazViewMoreLinkHeading"`
	Returnitinerary                  []interface{} `json:"returnItinerary"`
}

func NewSouthwest(firstName string, lastName string,
	confirmationNumber string,
	url string) *Southwest {
	southwest := Southwest{FirstName: firstName, LastName: lastName, ConfirmationNumber: confirmationNumber, Url: url}
	return &southwest
}

func (s *Southwest) CheckIn() (res *SouthwestResponse, err error) {
	//Create x-www-form-url-encoded
	// URL package
	v := url.Values{}
	v.Set("platform", "android")
	v.Set("firstName", s.FirstName)
	v.Set("lastName", s.LastName)
	v.Set("recordLocator", s.ConfirmationNumber)
	v.Set("serviceID", "flightcheckin_new")
	v.Set("appID", "swa")
	v.Set("appver", "2.4.1")
	v.Set("platformver", "5.0.GA_v201403042054")
	v.Set("channel", "rc")

	req, err := http.NewRequest("POST", s.Url, strings.NewReader(v.Encode()))

	if err != nil {
		log.Panic(err)
		return nil, err
	}

	req.Header.Add("Content-type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Panic(err)
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	log.Printf("Status code: %s\n", resp.StatusCode)
	log.Printf("Body: %s\n", string(body))

	res = new(SouthwestResponse)
	json.Unmarshal([]byte(string(body)), &res)

	return res, nil
}

func main() {
	var firstName string
	var lastName string
	var confirmationNumber string
	var email string
	url := "http://mobile.southwest.com/middleware/MWServlet"

	flag.StringVar(&firstName, "firstName", "", "First name for check in")
	flag.StringVar(&lastName, "lastName", "", "Last name for check in")
	flag.StringVar(&confirmationNumber, "confirmationNumber", "", "Confirmation Number for check in")
	flag.StringVar(&email, "email", "", "Email address to receive notifications")

	flag.Parse()

	if firstName == "" || lastName == "" || confirmationNumber == "" {
		log.Panic("Please ensure first name, last name and confirmation number are filled out")
		os.Exit(1)
	}

	s := NewSouthwest(firstName, lastName, confirmationNumber, url)
	resp, err := s.CheckIn()
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	if email != "" {
		apiKey := os.Getenv("MANDRILL_KEY")
		mandrillApi, err := gochimp.NewMandrill(apiKey)

		if err != nil {
			fmt.Println("Error instantiating client")
		}

		templateName := "notification"
		content := []gochimp.Var{
			gochimp.Var{"header", "<h1>Howdy and welcome!</h1>"},
			gochimp.Var{"main", fmt.Sprintf("<div>%s</div>", resp.Errmsg)},
		}

		renderedTemplate, err := mandrillApi.TemplateRender(templateName, content, nil)

		if err != nil {
			fmt.Println(err)
			fmt.Println("Error rendering template")
		}
		recipients := []gochimp.Recipient{
			gochimp.Recipient{Email: email},
		}

		message := gochimp.Message{
			Html:      renderedTemplate,
			Subject:   "All Set!",
			FromEmail: "checkin@isengard.io",
			FromName:  "Checkin Agent",
			To:        recipients,
		}

		_, err = mandrillApi.MessageSend(message, false)

		if err != nil {
			fmt.Println("Error sending message")
		}
	}

	log.Println(resp.Errmsg)
}
