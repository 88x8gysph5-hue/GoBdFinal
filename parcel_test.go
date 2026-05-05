package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// randSource источник псевдо случайных чисел.
	// Для повышения уникальности в качестве seed
	// используется текущее время в unix формате (в виде числа)
	randSource = rand.NewSource(time.Now().UnixNano())
	// randRange использует randSource для генерации случайных чисел
	randRange = rand.New(randSource)
)

// getTestParcel возвращает тестовую посылку
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// TestAddGetDelete проверяет добавление, получение и удаление посылки
func TestAddGetDelete(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	n, err := store.Add(parcel)
	require.NoError(t, err)
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	assert.NotEmpty(t, n)
	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	pars, err := store.Get(n)
	require.NoError(t, err)
	assert.Equal(t, parcel.Address, pars.Address)
	assert.Equal(t, parcel.Client, pars.Client)
	assert.Equal(t, parcel.CreatedAt, pars.CreatedAt)
	assert.Equal(t, n, pars.Number)
	assert.Equal(t, parcel.Status, pars.Status)
	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(n)
	require.NoError(t, err)

	pars, err = store.Get(n)

	assert.Empty(t, pars)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	n, err := store.Add(parcel)
	require.NoError(t, err)
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	assert.NotEmpty(t, n)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(n, newAddress)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	pars, err := store.Get(n)
	require.NoError(t, err)
	assert.Equal(t, newAddress, pars.Address)

	err = store.Delete(n)
	require.NoError(t, err)

	pars, err = store.Get(n)

	assert.Empty(t, pars)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcel := getTestParcel()
	n, err := store.Add(parcel)
	require.NoError(t, err)
	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	assert.NotEmpty(t, n)
	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newStat := "new stat"
	err = store.SetStatus(n, newStat)
	require.NoError(t, err)
	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	pars, err := store.Get(n)
	require.NoError(t, err)
	assert.Equal(t, newStat, pars.Status)

	err = store.Delete(n)
	require.NoError(t, err)

	pars, err = store.Get(n)
	assert.Empty(t, pars)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, err := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	require.NoError(t, err)
	store := NewParcelStore(db)
	parcels := []Parcel{
		getTestParcel(),
		getTestParcel(),
		getTestParcel(),
	}
	parcelMap := map[int]Parcel{}

	// задаём всем посылкам один и тот же идентификатор клиента
	client := randRange.Intn(10_000_000)
	parcels[0].Client = client
	parcels[1].Client = client
	parcels[2].Client = client

	// add
	for i := 0; i < len(parcels); i++ {
		id, err := store.Add(parcels[i])
		require.NoError(t, err)
		assert.NotEmpty(t, id)
		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id
		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.NoError(t, err)
	assert.Len(t, storedParcels, len(parcels))
	// check
	for _, parcel := range storedParcels {
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
		exp, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		assert.Equal(t, exp.Address, parcel.Address)
		assert.Equal(t, exp.CreatedAt, parcel.CreatedAt)
		assert.Equal(t, exp.Client, parcel.Client)
		assert.Equal(t, exp.Status, parcel.Status)
	}
}
