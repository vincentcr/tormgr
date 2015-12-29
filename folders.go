package main

type Folder struct {
	ID      RecordID `json:"id"`
	Name    string   `json:"name"`
	OwnerID RecordID `json:"-" db:"owner_id" `
}

func FolderGetAll(user User) (Cacheable, error) {
	return dbFind(Folder{OwnerID: user.ID}, "SELECT id,name from folders where owner_id=$1", user.ID)
}

func FolderGetByID(user User, id RecordID) (Cacheable, error) {
	return dbFindOne(Folder{OwnerID: user.ID, ID: id}, "SELECT * from folders where owner_id=$1 AND id=$2", user.ID, id)
}

func FolderGetByName(user User, name string) (Cacheable, error) {
	return dbFindOne(Folder{OwnerID: user.ID, Name: name}, "SELECT * from folders where owner_id=$1 AND name=$2", user.ID, name)
}

func FolderCreate(user User, name string) (Folder, error) {
	folder := Folder{ID: newID(), OwnerID: user.ID, Name: name}

	err := dbExecOnRecord("insert", "INSERT INTO folders(id, owner_id,name) VALUES(:id, :owner_id, :name)", &folder)
	return folder, err
}

func FolderRename(folder Folder) error {
	return dbExecOnRecord("update", "UPDATE folders SET name = :name WHERE id = :id AND owner_id = :owner_id", &folder)
}

func FolderDelete(folder Folder) error {
	return dbExecOnRecord("delete", "DELETE FROM folders where id = :id AND owner_id = :owner_id", &folder)
}

func (f Folder) cacheHint() cacheHint {
	params := map[string]interface{}{"ID": f.ID, "Name": f.Name}
	return cacheHintMake("folders", f.OwnerID, params)
}
