package pubsub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// MessageBody is PubSubからPushされたMessageのBody
type MessageBody struct {
	Message      Message `json:"message"`
	Subscription string  `json:"subscription"`
}

type messageBody struct {
	Message      message `json:"message"`
	Subscription string  `json:"subscription"`
}

// Message is PubSubからPushされたMessageの中で、Messageに関連すること
type Message struct {
	Data        MessageData `json:"data"`
	Attributes  Attributes  `json:"attributes"`
	MessageID   string      `json:"messageId"`
	PublishTime time.Time   `json:"publishTime"`
}

type message struct {
	Data        string     `json:"data"`
	Attributes  attributes `json:"attributes"`
	MessageID   string     `json:"messageId"`
	PublishTime time.Time  `json:"publishTime"`
}

type messageData struct {
	Kind                    string    `json:"kind"`
	ID                      string    `json:"id"`
	SelfLink                string    `json:"selfLink"`
	Name                    string    `json:"name"`
	Bucket                  string    `json:"bucket"`
	Generation              string    `json:"generation"`
	Metageneration          string    `json:"metageneration"`
	ContentType             string    `json:"contentType"`
	TimeCreated             time.Time `json:"timeCreated"`
	Updated                 time.Time `json:"updated"`
	StorageClass            string    `json:"storageClass"`
	TimeStorageClassUpdated time.Time `json:"timeStorageClassUpdated"`
	Size                    string    `json:"size"`
	MD5Hash                 string    `json:"md5hash"`
	MediaLink               string    `json:"mediaLink"`
	CRC32C                  string    `json:"crc32c"`
	Etag                    string    `json:"etag"`
}

// MessageData is PubSubからPushされたMessageのObjectに関連する内容
type MessageData struct {
	Kind                    string           `json:"kind"`
	ID                      string           `json:"id"`
	SelfLink                string           `json:"selfLink"`
	Name                    string           `json:"name"`
	Bucket                  string           `json:"bucket"`
	Generation              int              `json:"generation"`
	Metageneration          int              `json:"metageneration"`
	ContentType             string           `json:"contentType"`
	TimeCreated             time.Time        `json:"timeCreated"`
	Updated                 time.Time        `json:"updated"`
	StorageClass            StorageClassType `json:"storageClass"`
	TimeStorageClassUpdated time.Time        `json:"timeStorageClassUpdated"`
	Size                    int              `json:"size"`
	MD5Hash                 string           `json:"md5hash"`
	MediaLink               string           `json:"mediaLink"`
	CRC32C                  string           `json:"crc32c"`
	Etag                    string           `json:"etag"`
}

// Attributes is PubSubからPushされたMessageのObjectの変更に関連する内容
type Attributes struct {
	BucketID                string                 `json:"bucketId"`
	ObjectID                string                 `json:"objectId"`
	ObjectGeneration        string                 `json:"objectGeneration"`
	EventTime               time.Time              `json:"eventTime"`
	EventType               StorageNotifyEventType `json:"eventType"`
	PayloadFormat           string                 `json:"payloadFormat"`
	NotificationConfig      string                 `json:"notificationConfig"`
	OverwrittenByGeneration int                    `json:"overwrittenByGeneration"`
	OverwroteGeneration     int                    `json:"overwroteGeneration"`
}

type attributes struct {
	BucketID                string    `json:"bucketId"`
	ObjectID                string    `json:"objectId"`
	ObjectGeneration        string    `json:"objectGeneration"`
	EventTime               time.Time `json:"eventTime"`
	EventType               string    `json:"eventType"`
	PayloadFormat           string    `json:"payloadFormat"`
	NotificationConfig      string    `json:"notificationConfig"`
	OverwrittenByGeneration int       `json:"overwrittenByGeneration"`
	OverwroteGeneration     int       `json:"overwroteGeneration"`
}

// ReadBody is PubSubからPushされたリクエストのBodyを読み込む
func ReadBody(body io.Reader) (*MessageBody, error) {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	var b messageBody
	if err := json.Unmarshal(buf, &b); err != nil {
		return nil, err
	}

	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b.Message.Data))
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var md messageData
	if err := json.Unmarshal(d, &md); err != nil {
		return nil, err
	}
	psmd := MessageData{}
	psmd.Kind = md.Kind
	psmd.ID = md.ID
	psmd.SelfLink = md.SelfLink
	psmd.Name = md.Name
	psmd.Bucket = md.Bucket
	psmd.ContentType = md.ContentType
	psmd.TimeCreated = md.TimeCreated
	psmd.Updated = md.Updated
	psmd.TimeStorageClassUpdated = md.TimeStorageClassUpdated
	psmd.MD5Hash = md.MD5Hash
	psmd.MediaLink = md.MediaLink
	psmd.CRC32C = md.CRC32C
	psmd.Etag = md.Etag

	sct, err := ParseStorageClassType(md.StorageClass)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed ParseStorageClassType. v=%s", md.StorageClass))
	}
	psmd.StorageClass = sct

	size, err := strconv.Atoi(md.Size)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed Size ParseInt. Size = %s", md.Size))
	}
	psmd.Size = size

	g, err := strconv.Atoi(md.Generation)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed Generation Atoi. Generation = %s", md.Generation))
	}
	psmd.Generation = g

	mg, err := strconv.Atoi(md.Metageneration)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed Metageneration Atoi. Metageneration = %s", md.Metageneration))
	}
	psmd.Metageneration = mg

	a := Attributes{
		BucketID:                b.Message.Attributes.BucketID,
		ObjectID:                b.Message.Attributes.ObjectID,
		ObjectGeneration:        b.Message.Attributes.ObjectGeneration,
		EventTime:               b.Message.Attributes.EventTime,
		PayloadFormat:           b.Message.Attributes.PayloadFormat,
		NotificationConfig:      b.Message.Attributes.NotificationConfig,
		OverwrittenByGeneration: b.Message.Attributes.OverwrittenByGeneration,
		OverwroteGeneration:     b.Message.Attributes.OverwroteGeneration,
	}
	et, err := ParseStorageNotifyEventType(b.Message.Attributes.EventType)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed ParseStorageNotifyEventType. v=%s", b.Message.Attributes.EventType))
	}
	a.EventType = et

	return &MessageBody{
		Message: Message{
			Data:        psmd,
			Attributes:  a,
			MessageID:   b.Message.MessageID,
			PublishTime: b.Message.PublishTime,
		},
		Subscription: b.Subscription,
	}, nil
}
