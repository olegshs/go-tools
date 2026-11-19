package session

import (
	"net/http"

	"github.com/olegshs/go-tools/events"
)

// ManagerInterface is the interface of a session manager. It is implemented
// by Manager; see its methods for details.
type ManagerInterface interface {
	Start(r *http.Request, w http.ResponseWriter) SessionInterface
	Restart(s SessionInterface, w http.ResponseWriter)
	Open(id string) SessionInterface
	Id(r *http.Request) string
	SetId(w http.ResponseWriter, id string)
	GenerateId() string
	Events() *events.Dispatcher
}
