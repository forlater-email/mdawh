package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/emersion/go-message/mail"
)

type P map[string]string

type Mail struct {
	From  string
	Date  string
	Parts []P
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

func main() {
	r := bufio.NewReader(os.Stdin)
	mr, err := mail.CreateReader(r)

	newmail := Mail{}
	newmail.Date = mr.Header.Get("Date")
	newmail.From = mr.Header.Get("From")

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
			ct := p.Header.Get("Content-Type")
			b, _ := ioutil.ReadAll(p.Body)
			part := P{ct: string(b)}
			newmail.Parts = append(newmail.Parts, part)

		case *mail.AttachmentHeader:
			filename, _ := h.Filename()
			log.Printf("Got attachment: %v\n", filename)
		}
	}

	j, err := json.Marshal(newmail)
	if err != nil {
		log.Fatal(err)
	}
	makeReq(j)
	log.Printf("sent webhook: %v\n", newmail.From)
}
