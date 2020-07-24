package natslib

type NATSUpdateEHeader struct {
	Mode       string `json:"mode"`
	Entity     string `json:"entity"`
	EntityPath string `json:"entity_path"`
	ServerID   string `json:"serverID,omitempty"`
}

type NATSEntityUpdate struct {
	Header NATSUpdateEHeader `json:"header"`
	Buffer []byte            `json:"buffer"`
}

type NATSUpdateDHeader struct {
	Created      bool     `json:"created,omitempty"`
	Timestamp    int64    `json:"timestamp,omitempty"`
	Path         string   `json:"path,omitempty"`
	Doc          string   `json:"docId,omitempty"`
	DocVersion   string   `json:"docVersion,omitempty"`
	Expiry       int64    `json:"expiry"`
	ServerID     string   `json:"serverID,omitempty"`
	Entity       string   `json:"entity"`
	EntityPath   string   `json:"entityPath,omitempty"`
	DocPath      string   `json:"docPath,omitempty"`
	TArray       []string `json:"tarray,omitempty"`
	EntityAccess string   `json:"entityAccess"`
	RDID         string   `json:"rdid,omitempty"`
}

type NATSDataUpdate struct {
	Header NATSUpdateDHeader `json:"header"`
	Buffer []byte            `json:"buffer"`
}

type NATSResponseHeader struct {
	Created      bool   `json:"created,omitempty"`
	Timestamp    int64  `json:"timestamp,omitempty"`
	Path         string `json:"path,omitempty"`
	Doc          string `json:"docId,omitempty"`
	DocVersion   string `json:"docVersion,omitempty"`
	Status       int    `json:"status"`
	ErrorStr     string `json:"error_str,omitempty"`
	ServerID     string `json:"serverID,omitempty"`
	Chunks       int    `json:"chunks,omitempty"`
	EncryptedHdr []byte `json:"encrypted_hdr,omitempty"`
}

type NATSReqHeader struct {
	Mode          string                 `json:"mode"`
	Path          string                 `json:"path"`
	Flags         map[string]interface{} `json:"flags"`
	Authorization string                 `json:"authorization"`
	Accept        string                 `json:"accept"`
}

type NATSRequest struct {
	Header NATSReqHeader `json:"header"`
	Body   []byte        `json:"body"`
}

type NATSResponse struct {
	Header   NATSResponseHeader `json:"header"`
	Response string             `json:"response"`
}
