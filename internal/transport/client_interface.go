package transport

import "github.com/yourusername/avocado-exchange-server/internal/domain"

type MessageClient interface {
	domain.ClientConnection
	SetTeamName(teamName string)
	GetRemoteAddr() string
	RegisterWithServer(teamName string)
}
