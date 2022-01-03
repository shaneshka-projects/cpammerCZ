package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"
	"time"
)

// define email interface, and implemented auth and send method
type Mail interface {
	Auth()
	Send(message Message) error
}

type SendMail struct {
	user     string
	password string
	host     string
	port     string
	auth     smtp.Auth
}

type Attachment struct {
	name        string
	contentType string
}

type Message struct {
	from        string
	to          []string
	cc          []string
	bcc         []string
	subject     string
	body        string
	contentType string
	attachments []Attachment
}

type MM struct {
	mail    Mail
	message Message
}

var TO = "student_yekaterinburg@mzv.cz"
var start time.Time
var finish time.Time
func main() {
	fmt.Println("Хочу в Чехию")

	l, err := time.LoadLocation("Local")
	if err != nil {
		panic(err)
	}

	start = time.Date(2022, 1, 3, 8, 59, 58, 0, l)
	finish = time.Date(2022, 1, 3, 9, 0, 2, 0, l)

	fmt.Println("Будем слать с", start, "до", finish, "На адрес", TO)

	if time.Until(start) <= 0 {
		fmt.Println("завершено " + time.Until(start).String())
		return
	}

	mailAndMessages := []MM{
		create("shaneshka_test@inbox.ru", "xxxxxxxxxxxxxxxxx", "smtp.mail.ru", "587"),
    //"smtp.yandex.ru", "587"),
    //"smtp.gmail.com", "587"),
	}

	wait(start)

	fmt.Println("go go go go go go go")

	send(mailAndMessages)

	time.Sleep(time.Minute)
	fmt.Println("Конец")
}

func send(mailAndMessages []MM) {
	for _, mm := range mailAndMessages {
		go func(mm MM, finish time.Time) {
			i := 1
			for time.Until(finish) > 0  {
				fmt.Println("начинаем посылать с", mm.message.from, "№", i)
				mm.mail.Send(mm.message) //todo copy
				fmt.Println("закончили посылать с ", mm.message.from, "№", i)

				//r := rand.Intn(1000)
				//time.Sleep(time.Duration(r) * time.Millisecond)
				i++
			}
		}(mm, finish)
	}
}

func create(user, pass, host, port string) MM {
	var mail Mail
	mail = &SendMail{user: user, password: pass, host: host, port: port}
	message := Message{from: user,
		to:      []string{TO},
		cc:      []string{},
		bcc:     []string{},
		subject: "запись - XXX",
		body: `
Фамилия - 
Имя - 
Дата рождения - 
Номер загранпаспорта - 
Телефонный номер - 
Адрес электронной почты - 
Цель поездки - долговременное проживание - обучение.`,
		contentType: "text/plain;charset=utf-8",
		attachments: []Attachment{
			{
				name:        "XXX.pdf",
				contentType: "application/pdf",
			},
			{
				name:        "XXX2.pdf",
				contentType: "application/pdf",
			},
		},
	}
	fmt.Println("подготовили письмо для", user)
	return MM{mail, message}
}

func wait(start time.Time) {
	for time.Until(start) > 0 {
		t := time.Until(start)
		fmt.Println("ждем, старт в", start.String(), ", через", t.String())
		d := time.Second
		if t > time.Minute {
			d = time.Minute
		}
		time.Sleep(d)
	}
}

func (mail *SendMail) Auth() {
	mail.auth = smtp.PlainAuth("", mail.user, mail.password, mail.host)
}

func (mail SendMail) Send(message Message) error {
	mail.Auth()
	buffer := bytes.NewBuffer(nil)
	boundary := "GoBoundary"
	Header := make(map[string]string)
	Header["From"] = message.from
	Header["To"] = strings.Join(message.to, ";")
	Header["Cc"] = strings.Join(message.cc, ";")
	Header["Bcc"] = strings.Join(message.bcc, ";")
	Header["Subject"] = message.subject
	Header["Content-Type"] = "multipart/mixed;boundary=" + boundary
	Header["Mime-Version"] = "1.0"
	//Header["Date"] = time.Now().String()//todo
	mail.writeHeader(buffer, Header)

	body := "\r\n--" + boundary + "\r\n"
	body += "Content-Type:" + message.contentType + "\r\n"
	body += "\r\n" + message.body + "\r\n"
	buffer.WriteString(body)

	if message.attachments != nil {
		for _, att := range message.attachments {
			attachment := "\r\n--" + boundary + "\r\n"
			attachment += "Content-Transfer-Encoding:base64\r\n"
			attachment += "Content-Disposition:attachments\r\n"
			attachment += "Content-Type:" + att.contentType + ";name=\"" + att.name + "\"\r\n"
			buffer.WriteString(attachment)
			mail.writeFile(buffer, att.name)
		}
	}

	buffer.WriteString("\r\n--" + boundary + "--")

	err := smtp.SendMail(mail.host+":"+mail.port, mail.auth, message.from, message.to, buffer.Bytes())
	if err != nil {
		fmt.Println(message.from, err)
	}

	return nil
}

func (mail SendMail) writeHeader(buffer *bytes.Buffer, Header map[string]string) string {
	header := ""
	for key, value := range Header {
		header += key + ":" + value + "\r\n"
	}
	header += "\r\n"
	buffer.WriteString(header)
	return header
}

// read and write the file to buffer
func (mail SendMail) writeFile(buffer *bytes.Buffer, fileName string) {
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err.Error())
	}
	payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
	base64.StdEncoding.Encode(payload, file)
	buffer.WriteString("\r\n")
	for index, line := 0, len(payload); index < line; index++ {
		buffer.WriteByte(payload[index])
		if (index+1)%76 == 0 {
			buffer.WriteString("\r\n")
		}
	}
}
