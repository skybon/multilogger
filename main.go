package multilogger

import (
	"log"
	"time"

	"github.com/skybon/mgoHelpers"

	"gopkg.in/mgo.v2/bson"
)

// LogMessageType identifies the message's severity.
type LogMessageType int

const (
	MSG_MAJOR = LogMessageType(iota)
	MSG_MINOR
	MSG_DEBUG
)

type LogMessageBase struct {
	Dt   int64          `json:"date" bson:"date"`
	Type LogMessageType `json:"type" bson:"type"`
	Text string         `json:"text" bson:"text"`
}

type LogMessage struct {
	mgoHelpers.DbEntryBase
	LogMessageBase
}

func MakeLogMessage(date time.Time, t LogMessageType, text string) LogMessage {
	m := LogMessage{LogMessageBase: LogMessageBase{Dt: date.Unix(), Type: t, Text: text}}
	m.SetBsonID(bson.NewObjectId())

	return m
}

// LoggingModes determines the place to store logs.
type LoggingModes struct {
	Mongo  bool
	Mem    bool
	Stdout bool
}

type LogCollectionMongo struct {
	mongo mgoHelpers.MongoStorageInfo
}

type LogCollectionBase struct {
	activeModes LoggingModes
	data        []LogMessage
	inputChan   chan LogMessage
}

type LogCollection struct {
	LogCollectionMongo
	LogCollectionBase
}

func (c *LogCollection) Add(m LogMessage) {
	c.inputChan <- m
}

func (c *LogCollection) Logs() []LogMessage {
	return c.data
}

func (c *LogCollection) Close() {
	close(c.inputChan)
}

func (c *LogCollection) addLoop() {
	for {
		entry, ok := <-c.inputChan
		if ok {
			if c.activeModes.Mem {
				c.data = append(c.data, entry)
			}
			if c.activeModes.Mongo {
				c.mongo.Database.Insert(c.mongo.Collection, entry)
			}
			if c.activeModes.Stdout {
				log.Println(entry.Text)
			}

		} else {
			return
		}
	}
}

func MakeLogCollection(modes LoggingModes, mongoInstance *mgoHelpers.MongoDb) *LogCollection {
	var mongoStorage mgoHelpers.MongoStorageInfo
	if mongoInstance == nil {
		modes.Mongo = false
	} else {
		mongoStorage.Database = mongoInstance
		mongoStorage.Collection = "Logs"
	}

	c := LogCollection{LogCollectionMongo{mongo: mongoStorage}, LogCollectionBase{activeModes: modes, data: []LogMessage{}, inputChan: make(chan LogMessage)}}
	go c.addLoop()

	return &c
}
