package mailer

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	host string
	port string
	from string
}

func New(host, port, from string) *Mailer {
	return &Mailer{host: host, port: port, from: from}
}

func (m *Mailer) Enviar(destinatario, assunto, corpo string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.from, destinatario, assunto, corpo)
	addr := m.host + ":" + m.port
	return smtp.SendMail(addr, nil, m.from, []string{destinatario}, []byte(msg))
}
