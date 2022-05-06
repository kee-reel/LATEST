package storage

import "web/utils"

func (s *Storage) GetLanguages() map[int]string {
	rows, err := s.db.Query(`SELECT l.id, l.extention FROM languages AS l`)
	utils.Err(err)

	defer rows.Close()
	languages := map[int]string{}
	for rows.Next() {
		var id int
		var extention string
		err := rows.Scan(&id, &extention)
		utils.Err(err)
		languages[id] = extention
	}
	return languages
}
