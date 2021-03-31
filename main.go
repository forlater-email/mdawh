package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/emersion/go-message/mail"
	"github.com/microcosm-cc/bluemonday"
)

type Mail struct {
	From        string
	Date        string
	Body        string
	ContentType string
}

func makeReq(j []byte) {
	req, err := http.NewRequest("POST", "http://localhost:8001/webhook", bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func cleanHTML(body string) string {
	bm := bluemonday.StrictPolicy()
	bm.AddSpaceWhenStrippingTag(true)
	clean := bm.Sanitize(body)
	return clean
}

func jsonMail(p *mail.Part, mr *mail.Reader) ([]byte, error) {
	b, _ := ioutil.ReadAll(p.Body)
	body := string(b)
	ct := p.Header.Get("Content-Type")
	from := mr.Header.Get("From")
	date := mr.Header.Get("Date")

	// Prefer plaintext part over html
	if strings.Contains(ct, "text/plain") {
		m := Mail{from, date, body, "text/plain"}
		j, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}
		return j, nil

		// If there's no plaintext, fallback on html
		// Clean up HTML junk using bluemonday
	} else if strings.Contains(ct, "text/html") {
		cleanBody := cleanHTML(body)
		m := Mail{from, date, cleanBody, "text/html"}
		j, err := json.Marshal(m)
		if err != nil {
			log.Fatal(err)
		}
		return j, nil
	} else {
		return nil, errors.New("no plaintext or html parts")
	}
}

func main() {
	r := bufio.NewReader(os.Stdin)
	mr, err := mail.CreateReader(r)

	if err != nil {
		log.Fatal(err)
	}

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)

		}
		switch h := p.Header.(type) {
		case *mail.InlineHeader:
			jm, err := jsonMail(p, mr)
			fmt.Println(string(jm))
			if err != nil {
				log.Fatal(err)
			}
			makeReq(jm)

		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Printf("Got attachment: %v\n", filename)
		}
	}
}
