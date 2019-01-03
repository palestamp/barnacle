package postgres

import (
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

// const testDbURI = "postgresql://postgres@localhost:5434/barnacle"

// func TestPostgres(t *testing.T) {

// 	store, err := NewPostgresStore(testDbURI)
// 	assert.NoError(t, err)

// 	metadata := api.QueueMetadata{QueueID: "robocop", Table: "ut_robocop"}
// 	// err = store.CreateTopic(metadata)
// 	assert.NoError(t, err)

// 	tp, err := store.Connect(metadata)
// 	assert.NoError(t, err)

// 	// for i := 0; i < 20; i++ {
// 	// 	_, err = tp.Add(topic.EventInput{
// 	// 		Data:  fmt.Sprintf("data-%d", i),
// 	// 		Delay: 0,
// 	// 	})
// 	// 	assert.NoError(t, err)
// 	// }

// 	events, err := api.Poll(tp, 10, 10*time.Second, 40*time.Second)
// 	assert.NoError(t, err)

// 	litter.Dump(events)
// }

// func TestQueueMetadataStore(t *testing.T) {
// 	mstore, err := NewPostgresMetadataStore(testDbURI)
// 	assert.NoError(t, err)
// 	store, err := NewPostgresStore(testDbURI)
// 	assert.NoError(t, err)

// }
