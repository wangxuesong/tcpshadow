package model

type DataForward int

const (
	ClientToServer DataForward = iota
	ServerToClient
)

type Data struct {
	Forward DataForward
	Buffer  []byte
}

func (f DataForward) String() string {
	names := []string{
		"C->S",
		"S->C",
	}

	return names[f]
}

type SavePackage struct {
	Number  int         `json:"number"`
	Forward DataForward `json:"forward"`
	Length  int         `json:"length"`
	Buffer  []byte      `json:"buffer"`
}
