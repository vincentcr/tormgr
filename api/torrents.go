package main

type Torrent struct {
	ID        RecordID `json:"id"`
	Folder    string   `json:"folder"`
	SourceURL string   `json:"sourceURL"` //original url used to retrieve the torrent
	InfoHash  string   `json:"infoHash"`
	Data      []byte   `json:"data,omitifempty"`
	ownerID   RecordID
}
