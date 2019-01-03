package configuration

// func TestFSStorage(t *testing.T) {
// 	dir, err := ioutil.TempDir("", "barnacle-test")
// 	assert.NoError(t, err)

// 	storage, err := NewFSStorage(dir)
// 	assert.NoError(t, err)

// 	config := Config{
// 		"a": 1,
// 	}

// 	err = storage.Create("first", config)
// 	assert.NoError(t, err)

// 	c2, err := storage.Read("first")
// 	assert.NoError(t, err)

// 	assert.Equal(t, config, c2)
// }
