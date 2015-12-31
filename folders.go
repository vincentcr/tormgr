package main

type Folder struct {
	ID      RecordID `json:"id"`
	Name    string   `json:"name"`
	OwnerID RecordID `json:"-" db:"owner_id" `
}

func FolderGetAll(user User) (Cacheable, error) {
	return dbFind(Folder{OwnerID: user.ID}, "SELECT id,name from folders WHERE owner_id=$1", user.ID)
}

func FolderGetByID(user User, id RecordID) (Cacheable, error) {
	return dbFindOne(Folder{OwnerID: user.ID, ID: id}, "SELECT * from folders WHERE owner_id=$1 AND id=$2", user.ID, id)
}

func FolderGetByName(user User, name string) (Cacheable, error) {
	return dbFindOne(Folder{OwnerID: user.ID, Name: name}, "SELECT * from folders WHERE owner_id=$1 AND name=$2", user.ID, name)
}

func FolderGet(f Folder) (Cacheable, error) {
	return dbFindOne(f, "SELECT * from folders WHERE owner_id=$1 AND (id=$2 OR name=$2)", f.OwnerID, f.ID)
}

func FolderCreate(user User, name string) (Folder, error) {
	f := Folder{ID: newID(), OwnerID: user.ID, Name: name}

	err := dbExecOnRecord("INSERT INTO folders(id, owner_id,name) VALUES(:id, :owner_id, :name)", &f)
	return f, err
}

func FolderRename(f Folder) error {
	return dbExecOnRecord("UPDATE folders SET name=:name WHERE (name=:id OR id=:id) AND owner_id=:owner_id", &f)
}

func FolderDelete(f Folder) error {
	return dbExecOnRecord("DELETE FROM folders WHERE (name=:id OR id=:id) AND owner_id=:owner_id", &f)
}

func (f Folder) cacheHint() cacheHint {
	params := map[string]interface{}{"ID": f.ID, "Name": f.Name}
	return cacheHintMake("folders", f.OwnerID, params)
}
