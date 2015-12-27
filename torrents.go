package main

import (
	"bytes"
	"crypto/sha1"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/zeebo/bencode"
)

var (
	torrentMagnetURIPattern = regexp.MustCompile("^magnet:?.*\\bxt=urn:btih:([a-fA-F-0-9]+).*")
)

type Torrent struct {
	ID        RecordID      `json:"id"`
	OwnerID   RecordID      `json:"-" db:"owner_id"`
	Folder    string        `json:"folder"`
	InfoHash  string        `json:"infoHash" db:"info_hash"`
	Data      []byte        `json:"data,omitifempty"`
	SourceURL string        `json:"sourceURL" db:"source_url"`
	Status    TorrentStatus `json:"status"`
}

func (t Torrent) cacheHint() cacheHint {
	return cacheHint{userID: t.OwnerID, table: "torrents", recordID: t.ID}
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

func TorrentGetAll(user User) (Cacheable, error) {
	return dbFind(Torrent{OwnerID: user.ID}, "SELECT * from torrents where owner_id=$1", user.ID)
}

func TorrentGetByFolder(user User, folder string) (Cacheable, error) {
	return dbFind(Torrent{OwnerID: user.ID, Folder: folder}, "SELECT * from torrents where owner_id=$1 AND folder=$2", user.ID, folder)
}

func TorrentGet(user User, id RecordID) (Cacheable, error) {
	return dbFindOne(Torrent{OwnerID: user.ID, ID: id}, "SELECT * from torrents where id=$1 AND owner_id=$2", id, user.ID)
}

func TorrentCreateFromURL(user User, folder string, url string) (Torrent, error) {
	var data []byte
	infoHash, ok := torrentInfoHashFromMagnetURL(url)
	if !ok {
		var err error
		data, infoHash, err = torrentDataFromHTTPURL(url)
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

func torrentDataFromHTTPURL(url string) ([]byte, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request for %v: %v", url, err)
	}
	req.Header.Add("User-Agent", "tormgr/1.0") //some torrent servers do U/A filtering
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create torrent from source url %v: %v", url, err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read data at %v: %v", url, err)
	}

	infoHash, err := torrentComputeInfoHash(data)
	if err != nil {
		return nil, "", fmt.Errorf("invalid torrent data at %v: %v", url, err)
	}

	return data, infoHash, err
}

func torrentComputeInfoHash(data []byte) (string, error) {
	var bdict map[string]interface{}

	err := bencode.NewDecoder(bytes.NewBuffer(data)).Decode(&bdict)
	if err != nil {
		return "", fmt.Errorf("failed to parse as bencoded dictionary: %v", err)
	}

	info, ok := bdict["info"]
	if !ok {
		return "", fmt.Errorf("info section not found in: %#v", data)
	}
	var b bytes.Buffer
	err = bencode.NewEncoder(&b).Encode(info)
	if err != nil {
		return "", fmt.Errorf("unable to bencode %#v: %v", info, err)
	}

	infoHash := strings.ToUpper(fmt.Sprintf("%x", sha1.Sum(b.Bytes())))
	return infoHash, nil
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
