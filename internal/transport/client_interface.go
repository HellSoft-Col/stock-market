package transport

import "github.com/HellSoft-Col/stock-market/internal/domain"

type MessageClient interface {
	domain.ClientConnection
	SetTeamName(teamName string)
	GetRemoteAddr() string
	RegisterWithServer(teamName string)
}
