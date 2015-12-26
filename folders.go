package main

import "fmt"

type Folder struct {
	ID      RecordID  `json:"id"`
	Name    string    `json:"title"`
	Items   []Torrent `json:"items"`
	ownerID RecordID
}

func FolderCreate(user User, token string, folder *Folder) error {
	if folder.ID == "" {
		folder.ID = newID()
	}

	if folder.Items == nil {
		folder.Items = []Torrent{}
	}

	folder.ownerID = user.ID

	tx, err := services.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO folders(id,owner_id,name) VALUES($1,$2,$)",
		folder.ID, folder.ownerID, folder.Name)
	if err != nil {
		if isUniqueError(err) {
			return ErrUniqueViolation
		} else {
			return fmt.Errorf("unable to create folder %#v: %v", folder, err)
		}
	}

	// err = addTorrents(user, folder.ID, folder.Items, tx)
	// if err != nil {
	// 	return err
	// }

	// invalidateCache(folderCacheHint{user, folder.ID})

	return tx.Commit()
}

func FolderGetAll(user User) (Cacheable, error) {
	cacheHint := cacheHint{userID: user.ID, table: "folders"}
	return dbFind([]Folder{}, cacheHint, "SELECT id,name from folders where owner_id=?", user.ID)
}

func FolderGet(user User, id RecordID) (Cacheable, error) {
	cacheHint := cacheHint{userID: user.ID, table: "folders", recordID: id}
	return dbFindOne(Folder{}, cacheHint, "SELECT id,name from folders where id=? AND owner_id=?", id, user.ID)
}
