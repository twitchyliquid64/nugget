package serv

import (
	"net"
	"sync"

	"github.com/twitchyliquid64/nugget"
	"github.com/twitchyliquid64/nugget/logger"
)

// Manager is the concrete type representing the network side of a server,
// and managing client connections.
type Manager struct {
	wg                  sync.WaitGroup
	isOnline            bool
	listener            net.Listener
	logger              *logger.Logger
	provider            nugget.DataSourceSink
	isOptimisedProvider bool
}
