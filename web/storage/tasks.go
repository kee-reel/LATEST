package storage

import (
	"encoding/json"
	"fmt"
	"late/models"
	"late/utils"
	"log"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func (s *Storage) GetTasks(token *models.Token, task_ids []int) *[]models.Task {
	tasks := make([]models.Task, len(task_ids))
	units := map[int]*models.Unit{}
	projects := map[int]*models.Project{}
	for task_index, task_id := range task_ids {
		query, err := s.db.Prepare(`SELECT t.project_id, t.unit_id, t.position, t.extention, t.folder_name,
			t.name, t.description, t.input, t.output, t.score
			FROM tasks AS t WHERE t.id = $1`)
		utils.Err(err)

		var task models.Task
		var in_params_str []byte
		var project_id int
		var unit_id int
		err = query.QueryRow(task_id).Scan(
			&project_id, &unit_id, &task.Number, &task.Extention, &task.FolderName,
			&task.Name, &task.Desc, &in_params_str, &task.Output, &task.Score)
		utils.Err(err)

		project, ok := projects[project_id]
		if !ok {
			project = s.GetProject(project_id)
			projects[project_id] = project
		}
		task.Project = project
		task.ProjectId = project.Id

		unit, ok := units[unit_id]
		if !ok {
			unit = s.GetUnit(unit_id)
			units[unit_id] = unit
		}
		task.Unit = unit
		task.UnitId = unit.Id

		query, err = s.db.Prepare(`SELECT MAX(s.completion) FROM solutions AS s
			WHERE s.user_id = $1 AND s.task_id = $2 GROUP BY s.task_id`)
		utils.Err(err)

		task.Completion = 0
		_ = query.QueryRow(token.UserId, task_id).Scan(&task.Completion)

		err = json.Unmarshal(in_params_str, &task.Input)
		utils.Err(err)

		for i := range task.Input {
			input := &task.Input[i]
			param_type := input.Type
			param_range := input.Range
			if len(input.Dimensions) == 0 {
				input.TotalCount = 1
				input.Dimensions = []int{1}
			} else {
				last_d := 0
				input.TotalCount = 1
				for _, d := range input.Dimensions {
					if d != 0 {
						last_d = d
					}
					input.TotalCount *= last_d
				}
			}
			switch param_type {
			case "float", "double":
				float_range := make([]float64, 2)
				float_range[0], _ = strconv.ParseFloat(param_range[0], 64)
				float_range[1], _ = strconv.ParseFloat(param_range[1], 64)
				input.FloatRange = &float_range
			case "int":
				int_range := make([]int, 2)
				int_range[0], _ = strconv.Atoi(param_range[0])
				int_range[1], _ = strconv.Atoi(param_range[1])
				input.IntRange = &int_range
			default:
				utils.Err(fmt.Errorf("Param type \"%s\" not supported", param_type))
			}
		}

		task.Id = task_id
		tasks[task_index] = task
	}

	return &tasks
}

func (s *Storage) GetTaskTestData(task_id int) (string, string) {
	query, err := s.db.Prepare(`SELECT t.source_code, t.fixed_tests FROM tasks AS t WHERE t.id = $1`)
	utils.Err(err)

	var source_code string
	var fixed_tests string
	err = query.QueryRow(task_id).Scan(&source_code, &fixed_tests)
	utils.Err(err)

	return source_code, fixed_tests
}

func (s *Storage) GetTaskTemplate(lang string, task_id *int) string {
	query, err := s.db.Prepare(`SELECT t.source_code FROM solution_templates AS t WHERE t.extention = $1`)
	utils.Err(err)

	var source_code string
	err = query.QueryRow(lang).Scan(&source_code)
	utils.Err(err)

	return source_code
}

func (s *Storage) GetTaskIdsByFolder(folder_names *[]string) (*[]int, int) {
	var sb strings.Builder
	sb.WriteString("SELECT t.id FROM tasks AS t")

	var err error
	var project_id int
	var unit_id int
	folders_count := len(*folder_names)
	if folders_count >= 1 {
		query, err := s.db.Prepare("SELECT p.id FROM projects AS p WHERE p.folder_name = $1")
		utils.Err(err)
		err = query.QueryRow((*folder_names)[0]).Scan(&project_id)
		if err != nil {
			return nil, 0
		}
		if folders_count != 3 {
			sb.WriteString(" WHERE t.project_id = ")
			sb.WriteString(strconv.Itoa(project_id))
		}
	}
	if folders_count >= 2 {
		query, err := s.db.Prepare("SELECT u.id FROM units AS u WHERE u.project_id = $1 AND u.folder_name = $2")
		utils.Err(err)
		err = query.QueryRow(project_id, (*folder_names)[1]).Scan(&unit_id)
		if err != nil {
			return nil, 1
		}
		if folders_count != 3 {
			sb.WriteString(" AND t.unit_id = ")
			sb.WriteString(strconv.Itoa(unit_id))
		}
	}
	if folders_count == 3 {
		query, err := s.db.Prepare("SELECT t.id FROM tasks AS t WHERE t.project_id = $1 AND t.unit_id = $2 AND t.folder_name = $3")
		utils.Err(err)
		var task_id int
		err = query.QueryRow(project_id, unit_id, (*folder_names)[2]).Scan(&task_id)
		if err != nil {
			return nil, 2
		}
		task_ids := []int{task_id}
		return &task_ids, -1
	}

	rows, err := s.db.Query(sb.String())
	utils.Err(err)

	defer rows.Close()
	task_ids := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		utils.Err(err)
		task_ids = append(task_ids, task_id)
	}
	return &task_ids, -1
}

func (s *Storage) GetTaskIdsById(task_str_ids *[]string) (*[]int, bool) {
	var sb strings.Builder
	sb.WriteString("SELECT t.id FROM tasks AS t")
	if len(*task_str_ids) > 0 {
		sb.WriteString(" WHERE ")
		last_i := len(*task_str_ids) - 1
		for i, task_str_id := range *task_str_ids {
			_, err := strconv.Atoi(task_str_id)
			if err != nil {
				return nil, false
			}
			sb.WriteString("t.id=")
			sb.WriteString(task_str_id)
			if i != last_i {
				sb.WriteString(" OR ")
			}
		}
	}

	rows, err := s.db.Query(sb.String())
	utils.Err(err)

	defer rows.Close()
	task_ids := []int{}
	for rows.Next() {
		var task_id int
		err := rows.Scan(&task_id)
		utils.Err(err)
		task_ids = append(task_ids, task_id)
	}
	if len(*task_str_ids) != 0 && len(*task_str_ids) != len(task_ids) {
		log.Print("Requested:", task_str_ids, "Got:", task_ids)
		return nil, true
	}
	return &task_ids, true
}

func (s *Storage) GetProject(project_id int) *models.Project {
	query, err := s.db.Prepare(`SELECT s.name, s.folder_name FROM projects AS s WHERE s.id = $1`)
	utils.Err(err)

	var project models.Project
	err = query.QueryRow(project_id).Scan(&project.Name, &project.FolderName)
	utils.Err(err)
	project.Id = project_id
	return &project
}

func (s *Storage) GetUnit(unit_id int) *models.Unit {
	query, err := s.db.Prepare(`SELECT u.name, u.project_id, u.folder_name FROM units AS u WHERE u.id = $1`)
	utils.Err(err)

	var unit models.Unit
	err = query.QueryRow(unit_id).Scan(&unit.Name, &unit.ProjectId, &unit.FolderName)
	utils.Err(err)
	unit.Id = unit_id
	return &unit
}
