package database

import (
	"fmt"
	"github.com/ernado/poputchiki/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sort"
)

func (db *DB) GetMessagesFromUser(userReciever bson.ObjectId, userOrigin bson.ObjectId) (messages []*models.Message, err error) {
	err = db.messages.Find(bson.M{"user": userReciever, "chat": userOrigin}).Sort("time").All(&messages)
	return messages, err
}

func (db *DB) AddMessage(m *models.Message) error {
	return db.messages.Insert(m)
}

func (db *DB) RemoveMessage(id bson.ObjectId) error {
	return db.messages.RemoveId(id)
}

func (db *DB) GetMessage(id bson.ObjectId) (*models.Message, error) {
	message := &models.Message{}
	err := db.messages.FindId(id).One(message)
	return message, err
}
func (db *DB) setRead(query bson.M) error {
	messages := []models.Message{}
	change := mgo.Change{Update: bson.M{"$set": bson.M{"read": true}}}
	_, err := db.messages.Find(query).Apply(change, &messages)
	return err
}

func (db *DB) SetRead(user, id bson.ObjectId) error {
	query := bson.M{"_id": id, "user": user}
	return db.setRead(query)
}

func (db *DB) SetReadMessagesFromUser(userReciever bson.ObjectId, userOrigin bson.ObjectId) error {
	query := bson.M{"user": userReciever, "chat": userOrigin}
	return db.setRead(query)
}

func (db *DB) GetUnreadCount(id bson.ObjectId) (int, error) {
	query := bson.M{"user": id, "read": false}
	return db.messages.Find(query).Count()
}

type Dialogs []*models.Dialog

func (a Dialogs) Len() int {
	return len(a)
}

func (a Dialogs) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a Dialogs) Less(i, j int) bool {
	return a[j].Time.Before(a[i].Time)
}

func (db *DB) GetChats(id bson.ObjectId) ([]*models.Dialog, error) {
	var users []*models.User
	var ids []bson.ObjectId
	var result []*models.Dialog

	first := func(key string) bson.M {
		return bson.M{"$first": fmt.Sprintf("$%s", key)}
	}
	// preparing query
	match := bson.M{"$match": bson.M{"user": id}}
	s := bson.M{"$sort": bson.M{"time": -1}}
	group := bson.M{"$group": bson.M{"_id": bson.M{"chat": "$chat"}, "time": first("time"), "text": first("text"), "origin": first("origin")}}
	project := bson.M{"$project": bson.M{"_id": "$_id.chat", "time": "$time", "text": "$text", "origin": "$origin"}}
	pipeline := []bson.M{match, s, group, project}
	pipe := db.messages.Pipe(pipeline)
	iter := pipe.Iter()
	iter.All(&result)
	for i := range result {
		ids = append(ids, result[i].Id)
		ids = append(ids, result[i].Origin)
	}
	if err := db.users.Find(bson.M{"_id": bson.M{"$in": ids}}).All(&users); err != nil {
		return nil, err
	}

	usersMap := make(map[bson.ObjectId]*models.User)

	for _, u := range users {
		usersMap[u.Id] = u
	}

	for i := range result {
		result[i].User = usersMap[result[i].Id]
		result[i].OriginUser = usersMap[result[i].Origin]
	}

	sort.Sort(Dialogs(result))
	return result, nil
}
