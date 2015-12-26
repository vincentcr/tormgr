package main

type Torrent struct {
	ID        RecordID `json:"id"`
	Folder    string   `json:"folder"`
	SourceURL string   `json:"sourceURL"` //original url used to retrieve the torrent
	InfoHash  string   `json:"infoHash"`
	Data      []byte   `json:"data,omitifempty"`
	ownerID   RecordID
}

func TorrentGetAll(user User) (Cacheable, error) {
	cacheHint := cacheHint{userID: user.ID, table: "torrents"}
	return dbFind([]Torrent{}, cacheHint, "SELECT id,name from torrents where owner_id=?", user.ID)
}

func TorrentGetByFolder(user User, folder string) (Cacheable, error) {
	cacheHint := cacheHint{userID: user.ID, table: "torrents"}
	return dbFind([]Torrent{}, cacheHint, "SELECT id,name from torrents where owner_id=? AND folder=?", user.ID, folder)
}

func TorrenGet(user User, id RecordID) (Cacheable, error) {
	cacheHint := cacheHint{userID: user.ID, table: "torrents", recordID: id}
	return dbFindOne(Torrent{}, cacheHint, "SELECT id,name from torrents where id=? AND owner_id=?", id, user.ID)
}

func TorrentAddFromURL(url string) error {

}
