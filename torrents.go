package main

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/jackpal/bencode-go"
)

var (
	torrentMagnetURIPattern = regexp.MustCompile("^magnet:?.*\\bxt=urn:btih:([a-fA-F-0-9]+).*")
)

type Torrent struct {
	ID        RecordID      `json:"id"`
	OwnerID   RecordID      `json:"omit" db:"owner_id"`
	Folder    string        `json:"folder"`
	InfoHash  string        `json:"infoHash" db:"info_hash"`
	Data      []byte        `json:"data,omitifempty"`
	SourceURL string        `json:"sourceURL" db:"source_url"`
	Status    TorrentStatus `json:"status"`
}

type TorrentStatus string

const (
	TorrentStatusNew         TorrentStatus = "new"
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusDownloaded  TorrentStatus = "downloaded"
	TorrentStatusFailed      TorrentStatus = "failed"
)

func (status TorrentStatus) String() string {
	return string(status)
}

func (status TorrentStatus) Value() (driver.Value, error) {
	return string(status), nil
}

func (status *TorrentStatus) Scan(val interface{}) error {
	bytes, ok := val.([]byte)
	if !ok {
		return fmt.Errorf("Cast error: expected TorrentStatus bytes, got %v", val)
	}
	*status = TorrentStatus(string(bytes))
	return nil
}

func (t *Torrent) cacheHint() cacheHint {
	return cacheHint{userID: t.OwnerID, table: "torrents", recordID: t.ID}
}

func TorrentGetAll(user User) (Cacheable, error) {
	return dbFind(&Torrent{OwnerID: user.ID}, "SELECT id,name from torrents where owner_id=?", user.ID)
}

func TorrentGetByFolder(user User, folder string) (Cacheable, error) {
	return dbFind(&Torrent{OwnerID: user.ID, Folder: folder}, "SELECT id,name from torrents where owner_id=? AND folder=?", user.ID, folder)
}

func TorrentGet(user User, id RecordID) (Cacheable, error) {
	return dbFindOne(&Torrent{OwnerID: user.ID, ID: id}, "SELECT id,name from torrents where id=? AND owner_id=?", id, user.ID)
}

func TorrentCreateFromURL(user User, folder string, url string) (Torrent, error) {
	var data []byte
	infoHash, ok := torrentInfoHashFromMagnetURL(url)
	if !ok {
		var err error
		data, err = torrentDataFromHTTPURL(url)
		if err != nil {
			return Torrent{}, err
		}
	}

	torrent := Torrent{ID: newID(), OwnerID: user.ID, Folder: folder, InfoHash: infoHash, Data: data, SourceURL: url}
	return TorrentCreate(torrent)
}

func TorrentCreate(torrent Torrent) (Torrent, error) {
	err := dbExecOnRecord("insert", `
			INSERT INTO torrents(id, owner_id, folder, info_hash, data, source_url)
			VALUES(:id, :owner_id, :folder, :info_hash, :data, :source_url)
		`, &torrent)

	return torrent, err
}

func torrentDataFromHTTPURL(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent from source url %v: %v", url, err)
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at %v: %v", url, err)
	}

	//a torrent data is valid if it can be parsed as a bencoded dictionary
	//TODO: could be interesting to index the info part for searching
	var bdict map[string]interface{}
	err = bencode.Unmarshal(bytes.NewBuffer(data), bdict)
	if err != nil {
		return nil, fmt.Errorf("failed to parse data at '%v' as torrent file: %v", url, err)
	}

	return data, nil
}

func torrentInfoHashFromMagnetURL(magnetURL string) (string, bool) {
	match := torrentMagnetURIPattern.FindStringSubmatch(magnetURL)
	if len(match) > 0 {
		return match[1], true
	} else {
		return "", false
	}
}

func TorrentUpdate(torrent Torrent) error {
	sets := ""
	if torrent.Status != "" {
		sets += "status = :status"
	}
	if torrent.Folder != "" {
		if len(sets) > 0 {
			sets += ", "
		}
		sets += "folder = :folder"
	}

	return dbExecOnRecord("update", "UPDATE torrents SET "+sets+" WHERE id = :id AND owner_id = :owner_id", &torrent)
}

func TorrentDelete(torrent Torrent) error {
	return dbExecOnRecord("delete", "DELETE FROM torrents where id = :id AND owner_id = :owner_id", &torrent)
}
