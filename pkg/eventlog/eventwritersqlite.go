package eventlog

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Driver for the sql database
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/filecoin-project/mir/pkg/logging"
	t "github.com/filecoin-project/mir/pkg/types"
)

const (
	newTable string = `CREATE TABLE IF NOT EXISTS events (
    ts INTEGER,
    nodeid INTEGER NOT NULL,
    evtype TEXT,
    evdata TEXT
);`
	insert string = `INSERT INTO events VALUES (?,?,?,?);`
)

type sqliteWriter struct {
	db     *sql.DB
	nodeID t.NodeID
	logger logging.Logger
}

func NewSqliteWriter(filename string, nodeID t.NodeID, logger logging.Logger) (EventWriter, error) {
	db, err := sql.Open("sqlite3", filename+".db")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(newTable); err != nil {
		return nil, err
	}

	return &sqliteWriter{
		db:     db,
		nodeID: nodeID,
		logger: logger,
	}, nil
}

func (w sqliteWriter) Write(record EventRecord) error {
	// For each incoming event
	iter := record.Events.Iterator()
	for event := iter.Next(); event != nil; event = iter.Next() {
		jsonData, err := protojson.Marshal(event)
		if err != nil {
			return err
		}

		_, err = w.db.Exec(
			insert,
			record.Time,
			w.nodeID,
			fmt.Sprintf("%T", event.Type)[len("*eventpb.Event_"):],
			jsonData,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w sqliteWriter) Close() error {
	return w.db.Close()
}
