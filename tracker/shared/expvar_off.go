// +build !expvar

package shared

const expvarOn = false

var (
	Expvar expvals
)

type expvarIPmap struct {
	// M map[PeerIP]int8
}

func (e *expvarIPmap) Lock()               {}
func (e *expvarIPmap) Unlock()             {}
func (e *expvarIPmap) delete(ip PeerIP)    {}
func (e *expvarIPmap) inc(ip PeerIP)       {}
func (e *expvarIPmap) dec(ip PeerIP)       {}
func (e *expvarIPmap) dead(ip PeerIP) bool { return false }

type expvals struct {
	Connects    int64
	ConnectsOK  int64
	Announces   int64
	AnnouncesOK int64
	Scrapes     int64
	ScrapesOK   int64
	Errs        int64
	Clienterrs  int64
	Seeds       int64
	Leeches     int64
	IPs         expvarIPmap
}

func AddExpval(num *int64, inc int64) {}
func InitExpvar(peerdb *PeerDatabase) {}
