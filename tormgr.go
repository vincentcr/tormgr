package main

import "net/http"

func main() {
	if err := InitServices(); err != nil {
		panic(err)
	}
	setupServer()
}

func setupServer() {
	m := NewMux("/api/1.0")
	setupMiddlewares(m)
	routeUsers(m)
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

		return jsonify(res, w)
	})

	m.Post("/users/tokens", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		token, err := AccessTokenCreateFull(user)
		if err != nil {
			return err
		}

		res := map[string]interface{}{
			"user":  user,
			"token": token,
		}

		return jsonify(res, w)
	}))

	m.Get("/users/me", mustAuthenticateRW(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		return jsonify(user, w)
	}))
}

func routeTorrents(m *Mux) {
	m.Get("/folders", mustAuthenticateR(func(c *TMContext, w http.ResponseWriter, r *http.Request) error {
		user := c.MustGetUser()
		cacheable, err := FolderGetAll(user)
		if err != nil {
			return err
		}
		return writeCacheable(r, w, "application/json", cacheable)
	}))

	// m.Post("/folders")
	// m.Get("/folders/:folderID")
	// m.Delete("/folders/:folderID")
	// m.Put("/folders/:folderID")
	//
	// m.Get("/torrents")
	// m.Post("/torrents")
	// m.Get("/torrents/:torrentID")
	// m.Put("/torrents/:torrentID")
	// m.Delete("/torrents/:torrentID")
}
