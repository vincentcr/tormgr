package main

type Torrent struct {
	ID        RecordID `json:"id"`
	FolderID  RecordID `json:"folderID,omitifempty"`
	Hash      string   `json:"hash"`
	SourceURL string   `json:"sourceURL"`
	ownerID   RecordID
}
