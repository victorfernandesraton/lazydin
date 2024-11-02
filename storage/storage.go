package storage

type Storage interface {
	CreateTable() error
}

func CreateAllTables(storages []Storage) error {
	for _, storage := range storages {
		if err := storage.CreateTable(); err != nil {
			return err
		}
	}
	return nil
}
