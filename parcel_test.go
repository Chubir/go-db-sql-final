package main

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

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
	db, _ := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	i, err := store.Add(parcel)
	require.Nil(t, err)
	require.Greater(t, i, 0)

	// get
	// получите только что добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что значения всех полей в полученном объекте совпадают со значениями полей в переменной parcel
	np, err := store.Get(i)
	require.Nil(t, err)
	require.Equal(t, parcel.Address, np.Address)
	require.Equal(t, parcel.Client, np.Client)
	require.Equal(t, parcel.CreatedAt, np.CreatedAt)
	require.Equal(t, parcel.Status, np.Status)

	// delete
	// удалите добавленную посылку, убедитесь в отсутствии ошибки
	// проверьте, что посылку больше нельзя получить из БД
	err = store.Delete(i)
	require.Nil(t, err)
	s, _ := store.Get(i)
	require.Equal(t, s, Parcel{})
}

// TestSetAddress проверяет обновление адреса
func TestSetAddress(t *testing.T) {
	// prepare
	db, _ := sql.Open("sqlite", "tracker.db")
	store := NewParcelStore(db)
	parcel := getTestParcel() // настройте подключение к БД
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	i, err := store.Add(parcel)
	require.Nil(t, err)
	require.Greater(t, i, 0)

	// set address
	// обновите адрес, убедитесь в отсутствии ошибки
	newAddress := "new test address"
	err = store.SetAddress(i, newAddress)
	require.Nil(t, err)

	// check
	// получите добавленную посылку и убедитесь, что адрес обновился
	p, err := store.Get(i)
	require.Nil(t, err)
	require.Equal(t, p.Address, newAddress)
}

// TestSetStatus проверяет обновление статуса
func TestSetStatus(t *testing.T) {
	// prepare
	db, _ := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	store := NewParcelStore(db)
	parcel := getTestParcel()
	defer db.Close()

	// add
	// добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
	i, err := store.Add(parcel)
	require.Nil(t, err)
	require.Greater(t, i, 0)

	// set status
	// обновите статус, убедитесь в отсутствии ошибки
	err = store.SetStatus(i, ParcelStatusSent)
	require.Nil(t, err)

	// check
	// получите добавленную посылку и убедитесь, что статус обновился
	p, err := store.Get(i)
	require.Nil(t, err)
	require.Equal(t, p.Status, ParcelStatusSent)

}

// TestGetByClient проверяет получение посылок по идентификатору клиента
func TestGetByClient(t *testing.T) {
	// prepare
	db, _ := sql.Open("sqlite", "tracker.db") // настройте подключение к БД
	store := NewParcelStore(db)
	defer db.Close()

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
		id, err := store.Add(parcels[i]) // добавьте новую посылку в БД, убедитесь в отсутствии ошибки и наличии идентификатора
		require.Nil(t, err)

		// обновляем идентификатор добавленной у посылки
		parcels[i].Number = id

		// сохраняем добавленную посылку в структуру map, чтобы её можно было легко достать по идентификатору посылки
		parcelMap[id] = parcels[i]
	}

	// get by client
	storedParcels, err := store.GetByClient(client) // получите список посылок по идентификатору клиента, сохранённого в переменной client
	// убедитесь в отсутствии ошибки
	// убедитесь, что количество полученных посылок совпадает с количеством добавленных
	require.Nil(t, err)
	require.Len(t, storedParcels, len(parcels))

	// check
	for _, parcel := range storedParcels {
		parcelInMap, found := parcelMap[parcel.Number]
		require.True(t, found)
		require.Equal(t, parcel.Address, parcelInMap.Address)
		require.Equal(t, parcel.Client, parcelInMap.Client)
		require.Equal(t, parcel.CreatedAt, parcelInMap.CreatedAt)
		require.Equal(t, parcel.Status, parcelInMap.Status)
		// в parcelMap лежат добавленные посылки, ключ - идентификатор посылки, значение - сама посылка
		// убедитесь, что все посылки из storedParcels есть в parcelMap
		// убедитесь, что значения полей полученных посылок заполнены верно
	}
}
