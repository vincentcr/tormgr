package main

import "fmt"

type Folder struct {
	ID      RecordID  `json:"id"`
	Name    string    `json:"title"`
	Items   []Torrent `json:"items"`
	ownerID RecordID
}

func CreateFolder(user User, token string, folder *Folder) error {
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
