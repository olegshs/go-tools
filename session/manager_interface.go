package session

import (
	"net/http"

	"github.com/olegshs/go-tools/events"
)

type ManagerInterface interface {
	Start(r *http.Request, w http.ResponseWriter) SessionInterface
	Restart(s SessionInterface, w http.ResponseWriter)
	Open(id string) SessionInterface
	Id(r *http.Request) string
	SetId(w http.ResponseWriter, id string)
	GenerateId() string
	Events() *events.Dispatcher
}
