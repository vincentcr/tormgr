package main

import (
	"net/http"
	"regexp"

	"github.com/zenazn/goji"
)

const apiVersion = "1.0"

func setupServer() {
	api := NewMux("/api/" + apiVersion)
	setupMiddlewares(api)
	setupRoutes(api)
	goji.Serve()
}

func setupMiddlewares(m *Mux) {
	m.Use(panicRecovery)
	m.Use(cors)
	m.Use(authenticate)
}

func cors(c *TMContext, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization,Accept,Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD")
	return nil
}

func setupRoutes(m *Mux) {

	m.Get("/", func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		return jsonify(w, struct {
			AppName string
			Version string
		}{"tormgr", apiVersion})
	})

	//////////// USERS ////////////

	m.Post("/users", func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		var userReq userRequest
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

	//////////// FOLDERS ////////////

	m.Get("/folders", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		cacheable, err := FolderGetAll(user)
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	m.Post("/folders", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		var folderReq folderRequest
		if err := parseAndValidate(r, &folderReq); err != nil {
			return err
		}

		folder, err := FolderCreate(user, folderReq.Name)
		if err != nil {
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
		var folderReq folderRequest
		if err := parseAndValidate(r, &folderReq); err != nil {
			return err
		}

		folderID := c.URLParams["folderID"]
		folder := Folder{OwnerID: user.ID, ID: RecordID(folderID), Name: folderReq.Name}

		if err := FolderRename(folder); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))

	m.Get("/folders/:folderName", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		folderName := c.URLParams["folderName"]

		cacheable, err := TorrentGetByFolder(user, folderName)
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

	//////////// TORRENTS ////////////

	m.Get("/folders/:folder/torrents", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		cacheable, err := TorrentGetByFolder(user, c.URLParams["folder"])
		if err != nil {
			return err
		}

		return writeCacheable(r, w, "application/json", cacheable)
	}))

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

	m.Get("/torrents/:torrentID/data", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		torrentID := c.URLParams["torrentID"]

		data, err := TorrentGetData(user, RecordID(torrentID))
		if err != nil {
			return err
		}

		return writeAs(w, "application/x-bittorrent", data)
	}))

	m.Post("/torrents", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()

		var createReq torrentCreateRequest
		if err := parseAndValidate(r, &createReq); err != nil {
			return err
		}

		var err error
		var torrent Torrent
		if torrentInfoHashPattern.MatchString(createReq.URLOrInfoHash) {
			torrent, err = TorrentCreateFromInfoHash(user, createReq.Folder, createReq.URLOrInfoHash)
		} else if torrentURLPattern.MatchString(createReq.URLOrInfoHash) {
			torrent, err = TorrentCreateFromURL(user, createReq.Folder, createReq.URLOrInfoHash)
		} else {
			return NewHttpErrorWithText(http.StatusBadRequest, "urlOrInfoHash must be either a magnet or http url, or a info-hash")
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

		var editReq torrentEditRequest
		if err := parseAndValidate(r, &editReq); err != nil {
			return err
		}

		torrent := Torrent{
			ID:      RecordID(c.URLParams["torrentID"]),
			OwnerID: user.ID,
			Status:  TorrentStatus(editReq.Status),
			Folder:  editReq.Folder,
		}

		if err := TorrentUpdate(torrent); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	}))
}

type userRequest struct {
	Email    string `validate:"nonzero,regexp=^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[[:alnum:]]{2,}$"`
	Password string `validate:"nonzero,min=6"`
}

type folderRequest struct {
	Name string `validate:"nonzero,min=1"`
}

var (
	torrentURLPattern      = regexp.MustCompile("^(magnet|https?):.*")
	torrentInfoHashPattern = regexp.MustCompile("^[a-fA-F-0-9]{40}$")
)

type torrentCreateRequest struct {
	Folder        string `validate:"nonzero,min=1"`
	URLOrInfoHash string `validate:"nonzero,min=1"`
}

type torrentEditRequest struct {
	Folder string
	Status string
}
