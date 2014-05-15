package main

import (
	"encoding/json"
	"github.com/go-martini/martini"
	"labix.org/v2/mgo/bson"
	"log"
	"net/http"
)

type UserDB interface {
	Get(id bson.ObjectId) *User
	GetUsername(username string) *User
	// GetAll() []*User
	AddToFavorites(id bson.ObjectId, favId bson.ObjectId) error
	RemoveFromFavorites(id bson.ObjectId, favId bson.ObjectId) error
	Add(u *User) error
	GetFavorites(id bson.ObjectId) []*User
	Update(u *User) error
	// Delete(id bson.ObjectId) error
	AddGuest(id bson.ObjectId, guest bson.ObjectId) error
	GetAllGuests(id bson.ObjectId) ([]*User, error)
	SendMessage(origin bson.ObjectId, destination bson.ObjectId, text string) error
}

type TokenStorage interface {
	Get(hexToken string) (*Token, error)
	Generate(user *User) (*Token, error)
	Remove(token *Token) error
}

type TokenInterface interface {
	Get() (*Token, error)
}

const (
	JSON_HEADER     = "application/json; charset=utf-8"
	FORM_TARGET     = "target"
	FORM_EMAIL      = "email"
	FORM_PASSWORD   = "password"
	FORM_FIRSTNAME  = "firstname"
	FORM_SECONDNAME = "secondname"
	FORM_PHONE      = "phone"
	FORM_TEXT       = "text"
)

func JsonEncoder(c martini.Context, w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", JSON_HEADER)
}

func Render(value interface{}) (int, []byte) {
	// trying to marshal to json
	j, err := json.Marshal(value)
	if err != nil {
		j, err = json.Marshal(ErrorMarshal)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		return ErrorMarshal.Code, j
	}
	switch v := value.(type) {
	case Error:
		if v.Code == http.StatusInternalServerError {
			log.Println(v)
		}
		return v.Code, j
	default:
		return http.StatusOK, j
	}
}

func TokenWrapper(c martini.Context, r *http.Request, tokens TokenStorage, w http.ResponseWriter) {
	var t TokenAbstract

	q := r.URL.Query()
	hexToken := q.Get(TOKEN_URL_PARM)
	token, err := tokens.Get(hexToken)
	if err != nil {
		log.Println(err)
		code, data := Render(ErrorBackend)
		http.Error(w, string(data), code) // todo: set content-type
	}

	t = TokenHanlder{err, token}
	c.Map(t)
}

func GetUser(db UserDB, parms martini.Params, token TokenInterface) (int, []byte) {
	hexId := parms["id"]

	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()

	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)

	user := db.Get(id)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	// hiding private fields for non-owner
	if t == nil || t.Id != id {
		user.CleanPrivate()
	}

	return Render(user)
}

func AddToFavorites(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	user := db.Get(id)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	hexId = r.FormValue(FORM_TARGET)
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	favId := bson.ObjectIdHex(hexId)
	friend := db.Get(favId)
	if friend == nil {
		return Render(ErrorUserNotFound)
	}

	err := db.AddToFavorites(user.Id, friend.Id)
	if err != nil {
		return Render(ErrorBadRequest)
	}

	return Render("updated")
}

func RemoveFromFavorites(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	user := db.Get(id)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	hexId = r.FormValue(FORM_TARGET)
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	favId := bson.ObjectIdHex(hexId)
	friend := db.Get(favId)
	if friend == nil {
		return Render(ErrorUserNotFound)
	}

	err := db.RemoveFromFavorites(user.Id, friend.Id)
	if err != nil {
		return Render(ErrorBadRequest)
	}

	return Render("removed")
}

func GetFavorites(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	favorites := db.GetFavorites(id)
	if favorites == nil {
		return Render(ErrorUserNotFound)
	}

	for key, _ := range favorites {
		favorites[key].CleanPrivate()
	}

	return Render(favorites)
}

func GetGuests(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	guests, err := db.GetAllGuests(id)

	if err != nil {
		return Render(ErrorBackend)
	}

	if guests == nil {
		return Render(ErrorUserNotFound)
	}

	for key, _ := range guests {
		guests[key].CleanPrivate()
	}

	return Render(guests)
}

func AddToGuests(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	user := db.Get(id)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	hexId = r.FormValue(FORM_TARGET)
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	guestId := bson.ObjectIdHex(hexId)
	guest := db.Get(guestId)
	if guest == nil {
		return Render(ErrorUserNotFound)
	}

	go func() {
		err := db.AddGuest(user.Id, guest.Id)
		if err != nil {
			log.Println(err)
		}
	}()

	return Render("added to guests")
}

func Login(db UserDB, r *http.Request, tokens TokenStorage) (int, []byte) {
	username, password := r.FormValue(FORM_EMAIL), r.FormValue(FORM_PASSWORD)
	user := db.GetUsername(username)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	if user.Password != getHash(password) {
		log.Printf("%s != %s", user.Password, getHash(password))
		return Render(ErrorAuth)
	}

	t, err := tokens.Generate(user)
	if err != nil {
		return Render(ErrorBackend)
	}

	return Render(t)
}

func Logout(db UserDB, r *http.Request, tokens TokenStorage, token TokenInterface) (int, []byte) {
	t, _ := token.Get()

	if t == nil {
		return Render(ErrorAuth)
	}

	err := tokens.Remove(t)

	if err != nil {
		return Render(ErrorBackend)
	}

	return Render("logged out")
}

func Register(db UserDB, r *http.Request, tokens TokenStorage) (int, []byte) {
	u := UserFromForm(r)
	uDb := db.GetUsername(u.Email)
	if uDb != nil {
		return Render(ErrorBadRequest) // todo: change error name
	}

	err := db.Add(u)

	if err != nil {
		log.Println(err)
		return Render(ErrorBadRequest) // todo: change error name
	}

	t, err := tokens.Generate(u)
	if err != nil {
		return Render(ErrorBackend)
	}

	return Render(t)
}

func Update(db UserDB, r *http.Request, token TokenInterface, parms martini.Params) (int, []byte) {
	hexId := parms["id"]
	if !bson.IsObjectIdHex(hexId) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	id := bson.ObjectIdHex(hexId)
	if t.Id != id {
		return Render(ErrorNotAllowed)
	}

	user := db.Get(id)
	if user == nil {
		return Render(ErrorUserNotFound)
	}

	UpdateUserFromForm(r, user)
	err := db.Update(user)

	if err != nil {
		return Render(ErrorBackend)
	}
	return Render(user)
}

func SendMessage(db UserDB, parms martini.Params, r *http.Request, token TokenInterface) (int, []byte) {
	text := r.FormValue(FORM_TEXT)

	if text == BLANK {
		return Render(ErrorBadRequest)
	}

	destinationHex := parms["id"]

	if !bson.IsObjectIdHex(destinationHex) {
		return Render(ErrorBadId)
	}

	t, _ := token.Get()
	if t == nil {
		return Render(ErrorAuth)
	}

	destination := bson.ObjectIdHex(destinationHex)
	origin := t.Id

	go func() {
		err := db.SendMessage(origin, destination, text)
		if err != nil {
			log.Println(err)
		}
	}()

	return Render("message sent")
}
