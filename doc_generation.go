package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type DocGenData struct {
	Template   string
	TargetPath string
}

const doc_gen_script = "./scripts/gen_doc.sh"
const static_content_path = "../blog/_site/autotests/"

func ResultPath(doc_type string, suggested_doc_name *string) string {
	if suggested_doc_name == nil {
		uuid_str := uuid.New().String()
		return fmt.Sprintf("generated/%s/%s-%s.pdf", doc_type, doc_type, uuid_str)
	}
	return fmt.Sprintf("generated/%s/%s.pdf", doc_type, *suggested_doc_name)
}

const template_path = "doc_templates"

func TemplatePath(doc_type string) string {
	return fmt.Sprintf("%s/%s.tex", template_path, doc_type)
}
func GenTemplatePath(doc_type string) string {
	return fmt.Sprintf("%s/%s-generated.tex", template_path, doc_type)
}

func GenTaskDesc(task *Task, source_code *string, test_result *string) (*M, error) {
	input_strings := []string{}
	has_dimensions := false
	for _, input := range task.Input {
		if input.TotalCount > 1 {
			has_dimensions = true
			break
		}
	}

	if has_dimensions {
		input_strings = append(input_strings, "\\begin{tabular}{ |c|c|c|c| }\n\\hline\nНазвание & Размер массива & Тип & Допустимые значения \\\\ \n \\hline\n")
	} else {
		input_strings = append(input_strings, "\\begin{tabular}{ |c|c|c|c| }\n\\hline\nНазвание & Тип & Допустимые значения \\\\ \n \\hline\n")
	}

	for _, input := range task.Input {
		if has_dimensions {
			dimensions := []string{}
			cur_d := 0
			for _, d := range *input.Dimensions {
				if d != 0 {
					cur_d = d
				}
				dimensions = append(dimensions, strconv.Itoa(cur_d))
			}
			var dimensions_str string
			if (*input.Dimensions)[0] > 1 {
				dimensions_str = strings.Join(dimensions, "x")
			} else {
				dimensions_str = "Не массив"
			}
			input_strings = append(input_strings, fmt.Sprintf("%s & %s & %s & [%s, %s] \\\\ \n \\hline\n", input.Name, dimensions_str, input.Type, input.Range[0], input.Range[1]))
		} else {
			input_strings = append(input_strings, fmt.Sprintf("%s & %s & [%s, %s] \\\\ \n \\hline\n", input.Name, input.Type, input.Range[0], input.Range[1]))
		}
	}

	input_strings = append(input_strings, fmt.Sprintf("\n\\end{tabular}\n"))
	input_string := strings.Join(input_strings, "")

	page_content := M{}
	page_content["`TASK-NUMBER`"] = strconv.Itoa(task.Number)
	page_content["`TASK-NAME`"] = task.Name
	page_content["`TASK-DESC`"] = task.Desc
	page_content["`TASK-INPUT`"] = input_string
	page_content["`TASK-OUTPUT`"] = task.Output
	if test_result != nil {
		test_table_column_count := len(task.Input) + 1
		test_table_header_definition := []string{}
		for i := 0; i < test_table_column_count; i++ {
			test_table_header_definition = append(test_table_header_definition, "c")
		}
		test_table_header := []string{}
		for i := 0; i < len(task.Input); i++ {
			test_table_header = append(test_table_header, task.Input[i].Name)
		}
		test_table_header = append(test_table_header, "Результат")

		test_table := fmt.Sprintf(
			"\\begin{tabular}{ |%s| }\n\\hline\n%s\\\\\n\\hline\n%s\\end{tabular}\n",
			strings.Join(test_table_header_definition, "|"),
			strings.Join(test_table_header, " & "),
			*test_result)
		page_content["`TASK-TESTS`"] = test_table
	}
	if source_code != nil {
		page_content["`TASK-SOURCE-CODE`"] = *source_code
	}
	return &page_content, nil
}

type M map[string]string

func GenDoc(doc_type string, pages_content []M, user_data *UserData, suggested_doc_name *string) (*string, error) {
	title_page_gen_path := ""
	if user_data != nil {
		template_text_raw, err := ioutil.ReadFile(TemplatePath("title-page"))
		if err != nil {
			return nil, err
		}
		template_text := string(template_text_raw)
		//template_text = strings.Replace(template_text, "`STUDENT-NAME`", user_data.Name, -1)
		//template_text = strings.Replace(template_text, "`GROUP-NAME`", user_data.Group, -1)
		//template_text = strings.Replace(template_text, "`TEACHER-NAME`", user_data.Teacher, -1)

		title_page_gen_path = GenTemplatePath("title-page")
		err = ioutil.WriteFile(title_page_gen_path, []byte(template_text), 0777)
		if err != nil {
			return nil, err
		}
	}
	template_text_raw, err := ioutil.ReadFile(TemplatePath(doc_type))
	if err != nil {
		return nil, err
	}
	template_text := string(template_text_raw)
	result_text := ""
	for _, page_content := range pages_content {
		page_text := template_text
		for k, v := range page_content {
			page_text = strings.Replace(page_text, k, v, -1)
		}
		result_text += page_text
	}

	gen_template_path := GenTemplatePath(doc_type)
	err = ioutil.WriteFile(gen_template_path, []byte(result_text), 0777)
	if err != nil {
		return nil, err
	}

	gen_result_path := ResultPath(doc_type, suggested_doc_name)
	cmd_str := fmt.Sprintf("%s %s %s %s/%s %s", doc_gen_script, doc_type, gen_template_path, static_content_path, gen_result_path, title_page_gen_path)
	_, err = ExecCmd(cmd_str)
	if err != nil {
		return nil, fmt.Errorf("Error in PDF report generation.\nError log: %s", err.Error())
	}
	return &gen_result_path, nil
}
