package main

import "net/http"

func setupServer() {
	m := NewMux("/api/1.0")
	setupMiddlewares(m)
	routeUsers(m)
	routeFolders(m)
	routeTorrents(m)
	m.Serve()
}

func setupMiddlewares(m *Mux) {
	m.Use(cors)
	m.Use(authenticate)
}

func cors(c *TMContext, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept,Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD")
	return nil
}

type UserRequest struct {
	Email    string `validate:"nonzero,regexp=^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[[:alnum:]]{2,}$"`
	Password string `validate:"nonzero,min=6"`
}

func routeUsers(m *Mux) {

	m.Post("/users", func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		var userReq UserRequest
		if err := parseAndValidate(r, &userReq); err != nil {
			return err
		}

		user, err := UserCreate(userReq.Email, userReq.Password)
		if err == ErrUniqueViolation {
			return HttpError{StatusCode: 400, StatusText: "User already exists"}
		} else if err != nil {
			return err
		}

		token, err := AccessTokenCreateFull(user)
		if err != nil {
			return err
		}

		res := map[string]interface{}{
			"user":  user,
			"token": token,
		}

		return jsonify(w, res)
	})

	m.Post("/users/tokens", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		var access Access
		if err := access.Parse(r.URL.Query().Get("access")); err != nil {
			return NewHttpError(http.StatusBadRequest)
		}

		token, err := AccessTokenCreate(user, access)
		if err != nil {
			return err
		}

		res := map[string]interface{}{
			"user":  user,
			"token": token,
		}

		return jsonify(w, res)
	}))

	m.Get("/users/me", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		return jsonify(w, user)
	}))
}

func routeFolders(m *Mux) {

	m.Get("/folders", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		cacheable, err := FolderGetAll(user)
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	m.Get("/folders/:folderID", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		folderID := c.URLParams["folderID"]

		cacheable, err := FolderGet(user, RecordID(folderID))
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	m.Post("/folders", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		folder := Folder{OwnerID: user.ID}

		if err := parseFolderRequest(r, &folder); err != nil {
			return err
		}

		return jsonify(w, folder)
	}))

	m.Delete("/folders/:folderID", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		id := c.URLParams["folderID"]
		folder := Folder{OwnerID: user.ID, ID: RecordID(id)}

		if err := FolderDelete(folder); err != nil {
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

	m.Put("/folders/:folderID", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		id := c.URLParams["folderID"]
		folder := Folder{OwnerID: user.ID, ID: RecordID(id)}

		if err := parseFolderRequest(r, &folder); err != nil {
			return err
		}

		if err := FolderRename(folder); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

}

type folderRequest struct {
	Name string `validate:"nonzero,min=1"`
}

func parseFolderRequest(r *http.Request, folder *Folder) error {
	var folderReq folderRequest
	if err := parseAndValidate(r, &folderReq); err != nil {
		return err
	}
	folder.Name = folderReq.Name
	return nil
}

func routeTorrents(m *Mux) {

	m.Get("/torrents", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		cacheable, err := TorrentGetAll(user)
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	m.Get("/torrents/:torrentID", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		torrentID := c.URLParams["torrentID"]

		cacheable, err := TorrentGet(user, RecordID(torrentID))
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	m.Post("/torrents", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		createReq, err := parseTorrentCreateRequest(r)
		if err != nil {
			return err
		}

		var torrent Torrent
		if createReq.URL != "" {
			torrent, err = TorrentCreateFromURL(user, createReq.Folder, createReq.URL)
		} else if createReq.InfoHash != "" {
			torrent, err = TorrentCreate(Torrent{OwnerID: user.ID, Folder: createReq.Folder, InfoHash: createReq.InfoHash})
		}
		if err != nil {
			return err
		}

		return jsonify(w, torrent)
	}))

	m.Delete("/torrents/:torrentID", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		id := c.URLParams["torrentID"]
		torrent := Torrent{OwnerID: user.ID, ID: RecordID(id)}

		if err := TorrentDelete(torrent); err != nil {
			return err
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

	m.Put("/torrents/:torrentID", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		id := c.URLParams["torrentID"]
		torrent := Torrent{OwnerID: user.ID, ID: RecordID(id)}

		if err := parseTorrentEditRequest(r, &torrent); err != nil {
			return err
		}

		if err := TorrentUpdate(torrent); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
}

type torrentCreateRequest struct {
	Folder   string `validate:"nonzero,min=1"`
	InfoHash string `validate:"nonzero,min=40"`
	URL      string `validate:"nonzero"`
}

func parseTorrentCreateRequest(r *http.Request) (torrentCreateRequest, error) {
	var createReq torrentCreateRequest
	if err := parseAndValidate(r, &createReq); err != nil {
		return createReq, err
	}
	return createReq, nil
}

type torrentEditRequest struct {
	Folder string `validate:"nonzero,min=1"`
	Status string `validate:"nonzero,min=1"`
}

func parseTorrentEditRequest(r *http.Request, torrent *Torrent) error {
	var editReq torrentEditRequest
	if err := parseAndValidate(r, &editReq); err != nil {
		return err
	}
	torrent.Status = TorrentStatus(editReq.Status)
	torrent.Folder = editReq.Folder
	return nil
}
