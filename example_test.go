package hrr

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
)

type (
	DatabaseMethods interface {
		NewMonster(name string, cuteness int) (Monster, Error)
		ReadAllMonsters(sorted_by string) ([]Monster, Error)
		ReadMonster(id int64) (Monster, Error)
		UpdateMonster(id int64, change Monster) (Monster, Error)
		RemoveMonster(id int64) Error
	}

	Monster struct {
		ID         int64     `json:"id"`
		Name       string    `json:"name" validate:"required"`
		Cuteness   int       `json:"cuteness" validate:"requried"`
		CreateDate time.Time `json:"created_at"`
	}

	Skill struct {
		ID         int64     `json:"id"`
		Name       string    `json:"name" validate:"required"`
		Force      int       `json:"force" validate:"requried"`
		CreateDate time.Time `json:"created_at"`
	}

	sqlitePool struct {
		db *sqlx.DB
	}
)

// Prüft ob User und Passwort mit den Daten in der Datenbank übereinstimmt
func VerifyUser(user, password string) (bool, error) {
	return true, nil
}

func Example_LittleMonstersDefault() {
	pool, err := sqlx.Connect("sqlite3", ":memory")
	if err != nil {
		log.Fatal(err)
	}
	db := sqlitePool{pool}

	Logger = log.New()
	// Logt jeden Request aufruf
	LogAllRequests = true

	router := httprouter.New()
	router.POST("/v0/monster", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body := Monster{}
		req := Request(r).
			Log().
			DecodeBody(&body).
			ValidateBody().
			BaseAuth(VerifyUser)
		if err := req.Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).
			StatusCode(http.StatusCreated).
			Data(func() (interface{}, Error) {
				return db.NewMonster(body.Name, body.Cuteness)
			})

	})

	router.GET("/v0/monsters", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		Request(r).Process()

		v := r.URL.Query()

		Response(w, r).Data(func() (interface{}, Error) {
			return db.ReadAllMonsters(v.Get("sorted_by"))
		})
	})

	router.GET("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		if err := Request(r).ParamInt64(p, "monsterID", &id).Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Data(func() (interface{}, Error) {
			return db.ReadMonster(id)
		})
	})

	router.PUT("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		var monster Monster
		req := Request(r).
			ParamInt64(p, "monsterID", &id).
			DecodeBody(&monster).ValidateBody().
			BaseAuth(VerifyUser)

		if err := req.Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Data(func() (interface{}, Error) {
			return db.UpdateMonster(id, monster)
		})
	})

	router.DELETE("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		err := Request(r).ParamInt64(p, "monsterID", &id).Process()
		if err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).OK()
	})
}

func Example_LittleMonstersShortcuts() {
	pool, err := sqlx.Connect("sqlite3", ":memory")
	if err != nil {
		log.Fatal(err)
	}
	db := sqlitePool{pool}

	Logger = log.New()
	// Logt jeden Request aufruf
	LogAllRequests = true

	router := httprouter.New()

	router.POST("/v0/monster", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body := Monster{}
		if err := Request(r).Post(&body).BaseAuth(VerifyUser).Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Post(func() (interface{}, Error) {
			return db.NewMonster(body.Name, body.Cuteness)
		})
	})

	router.GET("/v0/monsters", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if err := Request(r).Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Data(func() (interface{}, Error) {
			v := r.URL.Query()
			return db.ReadAllMonsters(v.Get("sorted_by"))
		})
	})

	router.GET("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		if err := Request(r).ParamInt64(p, "monsterID", &id).Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Data(func() (interface{}, Error) {
			return db.ReadMonster(id)
		})
	})

	router.PUT("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		var monster Monster

		req := Request(r).Put(&monster, p, "monsterID", &id).BaseAuth(VerifyUser)
		if err := req.Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).Data(func() (interface{}, Error) {
			return db.UpdateMonster(id, monster)
		})
	})

	router.DELETE("/v0/monster/:monsterID", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var id int64
		err := Request(r).ParamInt64(p, "monsterID", &id).Process()
		if err != nil {
			Response(w, r).Error(err)
			return
		}

		Response(w, r).OK()
	})
}

func Example_LittleMonstersVerifyUser_v2() {
	verifyUser := func(user *string) func(string, string) (bool, error) {
		return func(u, pass string) (bool, error) {
			ok, err := VerifyUser(u, pass)
			if !ok || err != nil {
				return false, err
			}

			*user = u

			return true, nil
		}
	}

	router := httprouter.New()

	router.POST("/v0/monster", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		var user string
		var monster Monster
		if err := Request(r).Post(&monster).BaseAuth(verifyUser(&user)).Process(); err != nil {
			Response(w, r).Error(err)
			return
		}

		Logger.Println("User %v create new monster", user)

		Response(w, r).OK()
	})
}

func (sqlitePool) NewMonster(name string, cuteness int) (Monster, Error) {
	return Monster{
		ID:         1,
		Name:       name,
		Cuteness:   cuteness,
		CreateDate: time.Now(),
	}, nil
}

func (sqlitePool) ReadAllMonsters(sorted_by string) ([]Monster, Error) {
	return []Monster{
		{
			ID:         1,
			Name:       "X10",
			Cuteness:   -100,
			CreateDate: time.Now(),
		},
	}, nil
}

func (sqlitePool) ReadMonster(id int64) (Monster, Error) {
	return Monster{
		ID:         id,
		Name:       "Fluffy",
		Cuteness:   100,
		CreateDate: time.Now(),
	}, nil
}

func (sqlitePool) UpdateMonster(id int64, change Monster) (Monster, Error) {
	return Monster{
		ID:         id,
		Name:       change.Name,
		Cuteness:   change.Cuteness,
		CreateDate: time.Now(),
	}, nil
}
