package main

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/pivotal-golang/bytefmt"
	"github.com/zeebo/bencode"
)

var (
	torrentMagnetInfoHashPattern = regexp.MustCompile("^urn:btih:([a-fA-F-0-9]+).*")
)

type Torrent struct {
	ID        RecordID      `json:"id"`
	OwnerID   RecordID      `json:"-" db:"owner_id"`
	Folder    string        `json:"folder"`
	Title     string        `json:"title"`
	Trackers  []string      `json:"trackers"`
	InfoHash  string        `json:"infoHash" db:"info_hash"`
	Data      []byte        `json:"-"`
	SourceURL string        `json:"sourceURL,omitempty" db:"source_url"`
	Status    TorrentStatus `json:"status,omitempty"`
}

type TorrentFile struct {
	Path   string        `json:"path"`
	Length uint64        `json:"length"`
	Status TorrentStatus `json:"status,omitempty"`
}

type TorrentStatus string

const (
	TorrentStatusNew         TorrentStatus = "new"
	TorrentStatusDownloading TorrentStatus = "downloading"
	TorrentStatusDownloaded  TorrentStatus = "downloaded"
	TorrentStatusFailed      TorrentStatus = "failed"
)

func TorrentGetAll(user User) (Cacheable, error) {
	return dbFind(Torrent{OwnerID: user.ID}, torrentSelect("owner_id=$1"), user.ID)
}

func TorrentGetByFolder(user User, folder string) (Cacheable, error) {
	return dbFind(Torrent{OwnerID: user.ID, Folder: folder}, torrentSelect("owner_id=$1 AND folder=$2"), user.ID, folder)
}

func TorrentGet(user User, id RecordID) (Cacheable, error) {
	return dbFindOne(Torrent{OwnerID: user.ID, ID: id}, torrentSelect("id=$1 AND owner_id=$2"), id, user.ID)
}

func TorrentGetData(user User, id RecordID) ([]byte, error) {
	var data []byte
	err := services.db.
		QueryRow("SELECT data FROM torrents WHERE owner_id=$1 AND id=$2", user.ID, id).
		Scan(&data)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("unable to get data for torrent %v of %v: %v", id, user.ID, err)
	}

	return data, nil
}

func torrentSelect(where string) string {
	return "SELECT id,folder,title,owner_id,info_hash,source_url,status FROM torrents WHERE " + where
}

func TorrentCreateFromInfoHash(user User, folder string, infoHash string) (Torrent, error) {
	return TorrentCreate(Torrent{OwnerID: user.ID, Folder: folder, InfoHash: infoHash})
}

func TorrentCreate(torrent Torrent) (Torrent, error) {
	err := dbExecOnRecord(`
			INSERT INTO torrents(id, owner_id, folder, info_hash, data, source_url)
			VALUES(:id, :owner_id, :folder, :info_hash, :data, :source_url)
		`, &torrent)

	return torrent, err
}

func TorrentCreateFromURL(user User, folder string, urlStr string) (Torrent, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return Torrent{}, fmt.Errorf("Invalid url %v: %v", urlStr, err)
	}

	t := Torrent{OwnerID: user.ID, Folder: folder}
	if u.Scheme == "magnet" {
		err = torrentFromMagnetURL(u, &t)
	} else {
		err = torrentFromHTTPURL(u, &t)
	}

	if err != nil {
		return Torrent{}, err
	}

	return TorrentCreate(t)
}

func torrentFromMagnetURL(u *url.URL, t *Torrent) error {
	q := u.Query()
	match := torrentMagnetInfoHashPattern.FindStringSubmatch(q.Get("xt"))
	if len(match) == 0 {
		return fmt.Errorf("invalid info hash in magnet url: %v", u)
	}

	t.InfoHash = match[1]
	t.Title = q.Get("dn")
	t.Trackers = q["tr"]
	return nil
}

func torrentFromHTTPURL(u *url.URL, t *Torrent) error {
	data, err := torrentFetchData(u)
	if err != nil {
		return err
	}
	t.Data = data

	err = torrentParse(data, t)
	if err != nil {
		return fmt.Errorf("invalid torrent data at %v: %v", u, err)
	}

	return nil
}

func torrentFetchData(u *url.URL) ([]byte, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %v: %v", u, err)
	}
	req.Header.Add("User-Agent", "tormgr/1.0") //some torrent servers do U/A filtering and don't like go's default
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create torrent from source url %v: %v", u, err)
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data at %v: %v", u, err)
	}
	return data, nil
}

func torrentParse(data []byte, t *Torrent) error {
	var torrentData struct {
		Info struct {
			Name   string
			Length uint64
			Files  []struct {
				Path   string
				Length uint64
			}
		}
	}

	err := bencode.NewDecoder(bytes.NewBuffer(data)).Decode(&torrentData)
	if err != nil {
		return fmt.Errorf("torrentParse: failed to parse as bencoded dictionary: %v", err)
	}
	info := torrentData.Info

	var files []TorrentFile

	if info.Name != "" { //single-file
		files = []TorrentFile{TorrentFile{Path: info.Name, Length: info.Length}}
	} else { //multi-file
		l := len(info.Files)
		if l == 0 {
			return fmt.Errorf("0 file in torrent %#v", torrentData)
		}
		files = make([]TorrentFile, l)
		root := info.Name
		for i, file := range info.Files {
			path := path.Join(root, file.Path)
			files[i] = TorrentFile{Path: path, Length: file.Length}
		}
	}

	t.Title = torrentTitleFromFiles(files)

	infoHash, err := torrentComputeInfoHash(info)
	if err != nil {
		return err
	}
	t.InfoHash = infoHash

	return nil
}

func torrentTitleFromFiles(files []TorrentFile) string {
	var title string
	if len(files) == 1 {
		title = fmt.Sprintf("%v (%v)", files[0].Path, bytefmt.ByteSize(files[0].Length))
	} else {
		root := path.Dir(files[0].Path)
		var totLen uint64 = 0
		for _, f := range files {
			totLen += f.Length
		}
		title = fmt.Sprintf("%v (%d files, %v)", root, len(files), bytefmt.ByteSize(totLen))
	}

	return title
}

func torrentComputeInfoHash(info interface{}) (string, error) {
	var b bytes.Buffer
	err := bencode.NewEncoder(&b).Encode(info)
	if err != nil {
		return "", fmt.Errorf("unable to bencode %#v: %v", info, err)
	}

	infoHash := strings.ToUpper(fmt.Sprintf("%x", sha1.Sum(b.Bytes())))
	return infoHash, nil
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

	return dbExecOnRecord("UPDATE torrents SET "+sets+" WHERE id = :id AND owner_id = :owner_id", &torrent)
}

func TorrentDelete(torrent Torrent) error {
	return dbExecOnRecord("DELETE FROM torrents where id = :id AND owner_id = :owner_id", &torrent)
}

func (t Torrent) cacheHint() cacheHint {
	params := map[string]interface{}{"ID": t.ID, "Folder": t.Folder, "InfoHash": t.InfoHash, "Status": t.Status}
	return cacheHintMake("torrents", t.OwnerID, params)

}

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
